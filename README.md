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

| Command | Description |
|---------|-------------|
| `make docker-up` | Start |
| `make docker-down` | Stop |
| `make docker-logs` | View logs |
| `make deploy` | Full deploy |

## API Endpoints

| Method | Endpoint | Auth |
|--------|----------|------|
| GET | /health | No |
| GET | /v1/models | Yes |
| GET | /v1/backends | Yes |
| POST | /v1/chat/completions | Yes |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DOMAIN | localhost | Domain for SSL |
| PORT | 8090 | Gateway port |
| API_KEY | - | Your API key |
| AUTH_ENABLED | true | Enable auth |
| OPENCODE_HOST | host.docker.internal | OpenCode host |
| OPENCODE_PORT | 3001 | OpenCode port |

## Requirements

- Docker & Docker Compose
- OpenCode: `curl -fsSL https://opencode.ai/install | bash`

## License

MIT
