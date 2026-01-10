# AI Gateway

Personal OpenAI-compatible API gateway using free AI models via OpenCode backend.

**Created by Kashif Khan**

## Features

- OpenAI-compatible API
- Free AI models via OpenCode
- Streaming support (SSE)
- API key authentication
- Auto SSL with Caddy
- Docker deployment

## Deploy

```bash
git clone https://github.com/KashifKhn/ai-gateway.git
cd ai-gateway
cp .env.example .env
nano .env  # Set API_KEY and DOMAIN
./scripts/deploy.sh
```

## Commands

| Command            | Description |
| ------------------ | ----------- |
| `make docker-up`   | Start       |
| `make docker-down` | Stop        |
| `make docker-logs` | View logs   |
| `make deploy`      | Full deploy |

## API Endpoints

All endpoints except `/health` require `Authorization: Bearer <API_KEY>` header.

### GET /health

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "backends": { "opencode": "connected" },
  "uptime": 194
}
```

### GET /v1/models

```json
{
  "object": "list",
  "data": [{ "id": "big-pickle", "backend": "opencode", "free": true }]
}
```

### GET /v1/backends

```json
{
  "object": "list",
  "data": [{ "id": "opencode", "name": "OpenCode", "status": "active" }]
}
```

### POST /v1/chat/completions

**Request:**

```json
{
  "model": "big-pickle",
  "messages": [{ "role": "user", "content": "Hello" }],
  "stream": false
}
```

**Response:**

```json
{
  "id": "chatcmpl-xxx",
  "model": "big-pickle",
  "choices": [{ "message": { "role": "assistant", "content": "Hi there!" } }]
}
```

## Requirements

- Docker & Docker Compose
- OpenCode: `curl -fsSL https://opencode.ai/install | bash`

## License

MIT
