# UptimeGuardian

Go-based uptime/health check service with transition-based webhook alerts.

- Periodically checks HTTP endpoints (GET/HEAD)
- Stores checks + last results in a JSON file
- Sends a webhook only when status changes (DOWN, optional RECOVERY)
- Simple CRUD API to manage checks

---

## Quick Start (Docker)

1) Create an env file:

```bash
cp .env.example .env
```

2) Start the service:

### Docker Compose v2

```bash
docker compose up -d --build
```

### Docker Compose v1

```bash
docker-compose up -d --build
```

3) Health check:

```bash
curl http://localhost:8080/health
```

Expected:

```json
{"status":"ok"}
```

---

## Configuration

All configuration is via environment variables (typically in `.env`).

- `PORT` (default: `8080`) — API port
- `DATA_FILE` (default: `./data/checks.json`) — where checks are persisted
- `ALERT_WEBHOOK_URL` (default: empty) — if empty, alerting is disabled
- `REQUEST_USER_AGENT` (default: `go-healthcheck-bot/1.0`) — User-Agent for check requests

### API Authentication (recommended if exposed)

- `API_KEY` (default: empty)
  - If **empty**, the API is **unauthenticated** (convenient for localhost dev)
  - If **set**, every request must include one of:
    - `X-API-Key: <API_KEY>`
    - `Authorization: Bearer <API_KEY>`

### Network timeouts

- `HTTP_TIMEOUT_SEC` (default: `10`) — overall HTTP timeout for endpoint checks
- `WEBHOOK_TIMEOUT_SEC` (default: `5`) — timeout when posting alert webhooks

---

## API

Base URL: `http://localhost:8080`

### Authentication examples

If `API_KEY` is set:

```bash
curl -H 'X-API-Key: <API_KEY>' http://localhost:8080/checks
# or
curl -H 'Authorization: Bearer <API_KEY>' http://localhost:8080/checks
```

### Create a check

```bash
curl -X POST http://localhost:8080/checks \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "main-api",
    "url": "https://example.com/health",
    "method": "GET",
    "intervalSec": 30,
    "timeoutSec": 5,
    "expectedStatus": 200,
    "notifyOnRecovery": true,
    "enabled": true
  }'
```

### List checks

```bash
curl http://localhost:8080/checks
```

### Get a check

```bash
curl http://localhost:8080/checks/<id>
```

### Update a check

```bash
curl -X PUT http://localhost:8080/checks/<id> \
  -H 'Content-Type: application/json' \
  -d '{"intervalSec":60,"enabled":true}'
```

### Delete a check

```bash
curl -X DELETE http://localhost:8080/checks/<id>
```

### Run a check now

```bash
curl -X POST http://localhost:8080/checks/<id>/run
```

---

## Alerts (Webhook)

Alerts are sent **only on status transitions**:

- DOWN transition → always notify
- UP transition → notify only if `notifyOnRecovery=true`

Webhook payload example:

```json
{
  "checkId": "abc123",
  "checkName": "main-api",
  "url": "https://example.com/health",
  "previousStatus": "up",
  "currentStatus": "down",
  "statusCode": 500,
  "latencyMs": 88,
  "checkedAt": "2026-03-07T21:00:00Z",
  "error": ""
}
```

### Sending alerts to Discord/Slack

UptimeGuardian sends a **generic JSON payload**. Some platforms (Discord/Slack) expect a different schema.

Common approaches:

1) **Use a webhook relay** (recommended)
- Send `ALERT_WEBHOOK_URL` to your own small endpoint (Cloudflare Worker, AWS Lambda, a tiny server)
- Convert the payload into Discord/Slack webhook format

2) **Use automation tools**
- n8n / Zapier / Make.com: Receive the JSON payload → transform → post to Discord/Slack

#### Minimal relay example (Node.js)

Create `relay.js`:

```js
import express from "express";

const app = express();
app.use(express.json({ limit: "1mb" }));

// Set these as env vars in production
const DISCORD_WEBHOOK_URL = process.env.DISCORD_WEBHOOK_URL;

app.post("/uptimeguardian", async (req, res) => {
  const p = req.body;
  const text = `[${p.currentStatus?.toUpperCase?.() ?? ""}] ${p.checkName}\n${p.url}\nstatus=${p.statusCode} latency=${p.latencyMs}ms`;

  if (DISCORD_WEBHOOK_URL) {
    await fetch(DISCORD_WEBHOOK_URL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content: text }),
    });
  }

  res.json({ ok: true });
});

app.listen(3000, () => console.log("relay listening on :3000"));
```

Run it:

```bash
npm i express
node relay.js
```

Then set:

- `ALERT_WEBHOOK_URL=http://<relay-host>:3000/uptimeguardian`

> Note: For real deployments, add authentication and HTTPS on the relay.

---

## Persistence

Checks and their last results are stored in `DATA_FILE` (JSON). With the default `docker-compose.yml`, `./data` is mounted into the container.

---

## Operations

### Logs

```bash
docker compose logs -f
# or
docker-compose logs -f
```

### Stop / start / restart

```bash
docker compose stop && docker compose start
# or
docker-compose stop && docker-compose start
```

---

## Korean Guide

For a more detailed Korean walkthrough, see: `USER_GUIDE_KO.md`
