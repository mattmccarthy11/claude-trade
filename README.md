# VibeTrade Claude Integration

This open-source project provides Claude AI integration for the VibeTrade trading platform. It connects to the closed-source VibeTrade backend API to provide intelligent trading recommendations and analysis.

## Architecture

- **vibetrade-claude** (this repo): Open-source Claude AI integration layer
- **vibetrade**: Closed-source backend API that provides real market data and trading capabilities

## Features

- AI-powered trading recommendations using Claude
- Real-time options data integration
- Risk analysis and portfolio optimization
- Educational explanations of trading strategies
- Market data aggregation and analysis

## Setup

### Prerequisites

- Go 1.21 or higher
- Access to VibeTrade backend API
- Claude API key (optional - users can provide their own)

### Environment Variables

Configure the following environment variables to connect to the VibeTrade backend:

```bash
# VibeTrade API Configuration
export VIBETRADE_API_URL=http://localhost:8090  # URL of the VibeTrade backend
export VIBETRADE_USER_ID=your-user-id           # Your VibeTrade user ID

# Claude API Configuration (optional)
export ANTHROPIC_API_KEY=your-claude-api-key    # If using server-side API key
```

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/vibetrade-claude.git
   cd vibetrade-claude
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build ./cmd/unified_oauth_server
   ```

## Usage

### Running the Server

Start the vibetrade-claude server:

```bash
./unified_oauth_server
```

The server will start on port 8090 by default.

### API Endpoints

#### Claude AI Endpoints

- `POST /api/claude-code/connect` - Connect Claude API
- `GET /api/claude-code/recommendations` - Get AI trading recommendations
- `POST /api/claude-code/analyze-risk` - Analyze position risks
- `POST /api/claude-code/explain-strategy` - Get educational explanations

### Frontend Integration

The project includes a React component (`ClaudeTradeAssistant.tsx`) that can be integrated into your trading UI:

```tsx
import ClaudeTradeAssistant from './components/ClaudeTradeAssistant';

function App() {
  return (
    <ClaudeTradeAssistant 
      positions={userPositions}
      accountValue={accountValue}
    />
  );
}
```

## Options Data Integration

The MarketDataAggregator automatically fetches real options data from the VibeTrade backend when configured:

1. **With VibeTrade API**: Fetches real-time options chains, quotes, and Greeks
2. **Without VibeTrade API**: Falls back to mock data for development/testing

### Supported Options Data

- Options chains with bid/ask spreads
- Real-time option quotes
- Greeks (Delta, Gamma, Theta, Vega, Rho)
- Implied volatility
- Open interest and volume

## Development

### Project Structure

```
vibetrade-claude/
├── cmd/
│   └── unified_oauth_server/    # Server implementation
├── frontend/
│   └── ClaudeTradeAssistant.tsx # React UI component
├── internal/
│   ├── ai_assistant/            # Claude AI integration
│   │   ├── claude_client.go     # Claude API client
│   │   ├── market_data_aggregator.go # Market data fetching
│   │   ├── trading_assistant.go # Trading recommendations
│   │   └── risk_management.go  # Risk analysis
│   └── vibetrade/              # VibeTrade API client
│       └── client.go           # HTTP client for VibeTrade backend
└── go.mod                      # Go module definition
```

### Adding New Features

1. **New AI Capabilities**: Add to `internal/ai_assistant/trading_assistant.go`
2. **New API Endpoints**: Add handlers in `cmd/unified_oauth_server/ai_handlers.go`
3. **New Market Data**: Extend `internal/vibetrade/client.go`

## Testing

Run tests:

```bash
go test ./...
```

Test with mock data (no VibeTrade backend required):

```bash
# Don't set VIBETRADE_API_URL to use mock data
go run ./cmd/unified_oauth_server
```

## Contributing

We welcome contributions! Please:

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Security

- Never commit API keys or credentials
- The VibeTrade API uses header-based authentication (`X-User-ID`)
- Claude API keys can be provided by users or configured server-side
- All sensitive data is handled securely

## License

This project is open source and available under the MIT License.

## Support

For issues and questions:
- Open an issue on GitHub
- Check the [API documentation](docs/API_OPTIONS.md)
- Contact the maintainers

## Acknowledgments

- Built with the Claude AI assistant
- Integrates with the VibeTrade trading platform
- Uses real-time market data from SnapTrade