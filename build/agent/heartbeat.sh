#!/bin/bash
# Writes epoch timestamp to /foundry/state/heartbeat every 10 seconds.
# The control plane's health monitor reads this to detect unresponsive containers.

while true; do
    date +%s > /foundry/state/heartbeat
    sleep 10
done
