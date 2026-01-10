# Kashif AI Gateway

Your personal OpenAI/Gemini alternative powered by OpenCode's free models.

## Quick Start

### 1. Prerequisites

- Go 1.21+
- OpenCode installed (`curl -fsSL https://opencode.ai/install | bash`)

### 2. Build

```bash
cd ai-gateway
go mod tidy
go build -o ai-gateway ./cmd/server
```

### 3. Start OpenCode Server

In a separate terminal:
```bash
opencode serve --port 3001
```

### 4. Start AI Gateway

```bash
./ai-gateway
```

Or use the combined script:
```bash
./scripts/start.sh
```

## API Usage

### Default API Key
```
sk-kashif-ai-gateway-secret-key-2024
```

### Health Check
```bash
curl http://localhost:8080/health
```

### List Models
```bash
curl -H "Authorization: Bearer sk-kashif-ai-gateway-secret-key-2024" \
     http://localhost:8080/v1/models
```

### Chat Completion (Non-streaming)
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-kashif-ai-gateway-secret-key-2024" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "big-pickle",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Chat Completion (Streaming)
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-kashif-ai-gateway-secret-key-2024" \
  -H "Content-Type: application/json" \
  -N \
  -d '{
    "model": "big-pickle",
    "messages": [{"role": "user", "content": "Tell me a joke"}],
    "stream": true
  }'
```

## Using with OpenAI SDK

### Python
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-kashif-ai-gateway-secret-key-2024"
)

response = client.chat.completions.create(
    model="big-pickle",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)
```

### JavaScript/TypeScript
```typescript
import OpenAI from 'openai'

const client = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'sk-kashif-ai-gateway-secret-key-2024'
})

const response = await client.chat.completions.create({
  model: 'big-pickle',
  messages: [{ role: 'user', content: 'Hello!' }]
})
console.log(response.choices[0].message.content)
```

## Available Models

| Model | Aliases | Free |
|-------|---------|------|
| big-pickle | pickle, bp | Yes |
| grok-code-fast-1 | grok, grok-fast | Yes |
| glm-4.7 | glm, glm4 | Yes |
| minimax-m2.1 | minimax, mm | Yes |

## Configuration

Edit `config/config.yaml` to customize:
- Server port and host
- API keys
- Rate limiting
- Backend configurations

## Custom API Key

Generate a new key:
```bash
./scripts/generate-key.sh
```

Or set via environment:
```bash
export AI_GATEWAY_API_KEY="your-custom-key"
./ai-gateway
```

## Directory Structure

```
ai-gateway/
├── cmd/server/          # Main entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── adapters/        # Backend adapters
│   ├── auth/            # Authentication
│   ├── config/          # Configuration
│   └── models/          # Data models
├── config/              # Config files
├── scripts/             # Utility scripts
├── Makefile
└── README.md
```

## License

MIT
