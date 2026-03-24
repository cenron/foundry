import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../api/swagger/swagger.json",
  output: "src/api/generated",
  plugins: [
    "@hey-api/typescript",
    {
      name: "@hey-api/client-fetch",
      runtimeConfigPath: "../hey-api",
    },
    "@hey-api/sdk",
    {
      name: "@tanstack/react-query",
      queryOptions: true,
      mutationOptions: true,
    },
  ],
})
