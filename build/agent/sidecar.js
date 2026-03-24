const amqp = require("amqplib");
const http = require("http");
const fs = require("fs");
const path = require("path");

const RABBITMQ_URL = process.env.RABBITMQ_URL || "amqp://guest:guest@localhost:5672/";
const PROJECT_ID = process.env.PROJECT_ID || "unknown";
const SIDECAR_PORT = 3000;

let channel = null;

async function connectRabbitMQ() {
  const conn = await amqp.connect(RABBITMQ_URL);
  channel = await conn.createChannel();

  // Declare exchanges
  await channel.assertExchange("foundry.commands", "topic", { durable: true });
  await channel.assertExchange("foundry.events", "topic", { durable: true });
  await channel.assertExchange("foundry.logs", "topic", { durable: true });

  // Subscribe to commands for this project
  const q = await channel.assertQueue(`commands.${PROJECT_ID}`, { durable: true });
  await channel.bindQueue(q.queue, "foundry.commands", `commands.${PROJECT_ID}.*`);

  channel.consume(q.queue, (msg) => {
    if (!msg) return;

    try {
      const command = JSON.parse(msg.content.toString());
      console.log(`[sidecar] received command: ${command.type}`);
      handleCommand(command);
      channel.ack(msg);
    } catch (err) {
      console.error(`[sidecar] error handling command: ${err.message}`);
      channel.nack(msg, false, false);
    }
  });

  console.log(`[sidecar] connected to RabbitMQ, listening for commands on project ${PROJECT_ID}`);
}

function handleCommand(command) {
  switch (command.type) {
    case "assign_task": {
      // Validate agent_role to prevent path traversal.
      const agentRole = command.agent_role || "";
      if (!/^[a-z0-9-]+$/.test(agentRole)) {
        console.error(`[sidecar] invalid agent_role: ${agentRole}`);
        break;
      }
      // Write task assignment to shared volume for the target agent
      const taskFile = path.join("/shared", "tasks", `${agentRole}.json`);
      fs.mkdirSync(path.dirname(taskFile), { recursive: true });
      fs.writeFileSync(taskFile, JSON.stringify(command, null, 2));
      publishEvent("task_assigned", command);
      break;
    }

    case "pause_agent":
      // Signal the supervisor to pause
      fs.writeFileSync("/foundry/state/pause-signal", command.agent_role || "all");
      publishEvent("pause_requested", command);
      break;

    default:
      console.log(`[sidecar] unknown command type: ${command.type}`);
  }
}

function publishEvent(eventType, payload) {
  if (!channel) return;

  const routingKey = `events.${PROJECT_ID}.${eventType}`;
  const msg = JSON.stringify({
    project_id: PROJECT_ID,
    type: eventType,
    payload,
    timestamp: new Date().toISOString(),
  });

  channel.publish("foundry.events", routingKey, Buffer.from(msg));
}

function publishLog(agentId, line) {
  if (!channel) return;

  const routingKey = `logs.${PROJECT_ID}.${agentId}`;
  channel.publish("foundry.logs", routingKey, Buffer.from(line));
}

// HTTP notification endpoint — agents/supervisor POST here to signal events
const server = http.createServer((req, res) => {
  if (req.method === "POST" && req.url === "/notify") {
    let body = "";
    req.on("data", (chunk) => (body += chunk));
    req.on("end", () => {
      try {
        const event = JSON.parse(body);
        publishEvent(event.type || "notification", event);
        res.writeHead(200);
        res.end("ok");
      } catch (err) {
        res.writeHead(400);
        res.end("invalid json");
      }
    });
    return;
  }

  if (req.method === "GET" && req.url === "/health") {
    res.writeHead(200);
    res.end(JSON.stringify({ status: "ok", project_id: PROJECT_ID }));
    return;
  }

  res.writeHead(404);
  res.end("not found");
});

// Watch /shared/status/ for agent completion signals
function watchStatusDir() {
  const statusDir = "/shared/status";
  fs.mkdirSync(statusDir, { recursive: true });

  setInterval(() => {
    try {
      const files = fs.readdirSync(statusDir).filter((f) => f.endsWith(".json"));
      for (const file of files) {
        const filePath = path.join(statusDir, file);
        const content = JSON.parse(fs.readFileSync(filePath, "utf-8"));

        if (content._processed) continue;

        console.log(`[sidecar] agent status update: ${file} -> ${content.status}`);
        publishEvent(`agent_${content.status}`, content);

        // Mark as processed
        content._processed = true;
        fs.writeFileSync(filePath, JSON.stringify(content, null, 2));
      }
    } catch (err) {
      // Ignore errors during polling
    }
  }, 2000);
}

async function main() {
  try {
    await connectRabbitMQ();
  } catch (err) {
    console.error(`[sidecar] RabbitMQ connection failed: ${err.message}`);
    console.log("[sidecar] running without RabbitMQ (standalone mode)");
  }

  server.listen(SIDECAR_PORT, () => {
    console.log(`[sidecar] HTTP server listening on port ${SIDECAR_PORT}`);
  });

  watchStatusDir();
}

main().catch(console.error);
