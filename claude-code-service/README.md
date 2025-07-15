# Claude Code Service

This Node.js service integrates the Claude Code SDK to provide natural language trading assistance for VibeTrade.

## Features

- Natural language trading conversations
- Real-time market analysis
- Trading strategy recommendations
- Educational explanations
- WebSocket support for streaming responses

## Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Copy `.env.example` to `.env` and configure:
   ```bash
   cp .env.example .env
   ```

3. Set your Anthropic API key:
   ```
   ANTHROPIC_API_KEY=your-api-key-here
   ```

4. Start the service:
   ```bash
   npm start
   ```

   For development with auto-reload:
   ```bash
   npm run dev
   ```

## Architecture

- **server.js**: WebSocket server and HTTP endpoints
- **claude-handler.js**: Claude Code SDK integration and message processing

## API Endpoints

- `GET /health` - Health check
- `WS /ws` - WebSocket endpoint for real-time communication

## WebSocket Protocol

### Client to Server Messages

```json
{
  "type": "chat",
  "content": "Find me profitable iron condors on tech stocks"
}
```

### Server to Client Messages

```json
{
  "type": "assistant",
  "content": "I'll analyze tech stocks for iron condor opportunities...",
  "session_id": "abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Integration with VibeTrade

The service connects to the VibeTrade backend to access:
- Real-time market data
- Options chains
- Portfolio information
- Trading capabilities

## Environment Variables

- `PORT` - Server port (default: 3001)
- `ANTHROPIC_API_KEY` - Your Anthropic API key
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `VIBETRADE_API_URL` - VibeTrade backend URL
- `CLAUDE_CODE_MAX_TURNS` - Maximum conversation turns
- `CLAUDE_CODE_MODEL` - Claude model to use

## Security

- API keys are stored server-side only
- WebSocket connections require authentication
- Rate limiting is implemented
- All trading actions require confirmation