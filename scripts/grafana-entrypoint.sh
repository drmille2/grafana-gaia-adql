#!/bin/sh
set -e

PLUGIN_ID="davidmillerpf9-gaia-adql-datasource"
ADMIN_USER="${GF_SECURITY_ADMIN_USER:-admin}"
ADMIN_PASS="${GF_SECURITY_ADMIN_PASSWORD:-admin}"

# Start Grafana in background using the standard entrypoint
/run.sh &
GRAFANA_PID=$!

# Wait for the Grafana HTTP API to become ready
echo "grafana-entrypoint: waiting for Grafana to start..."
i=0
while [ $i -lt 60 ]; do
  if curl -sf http://localhost:3000/api/health > /dev/null 2>&1; then
    echo "grafana-entrypoint: Grafana is ready"
    break
  fi
  sleep 2
  i=$((i + 1))
done

# Enable the plugin so it appears in the UI
echo "grafana-entrypoint: enabling plugin ${PLUGIN_ID}"
curl -sf \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"enabled":true,"pinned":true}' \
  "http://${ADMIN_USER}:${ADMIN_PASS}@localhost:3000/api/plugins/${PLUGIN_ID}/settings" \
  && echo "grafana-entrypoint: plugin enabled" \
  || echo "grafana-entrypoint: WARNING - could not enable plugin (may already be enabled)"

# Hand off to the Grafana process
wait $GRAFANA_PID
