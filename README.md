# VibeTrade Claude AI Trading Assistant

An AI-powered options trading assistant that integrates Claude into the VibeTrade platform, providing intelligent trade recommendations, risk analysis, and educational insights.

## Overview

This integration brings the power of Claude AI to options traders, similar to the Reddit user who achieved a 100% win rate using ChatGPT for options trading. Our implementation provides:

- 🤖 AI-generated trade recommendations based on real-time market data
- 📊 Comprehensive risk analysis and portfolio optimization
- 🎓 Educational explanations of trading strategies
- 🔒 Built-in safety features and risk management
- 🔐 Secure API key management with encryption

## Features

### AI Trade Recommendations
- Analyzes 50+ data points including fundamentals, technicals, and options Greeks
- Generates exactly 5 high-probability trades per analysis
- Filters trades by strict criteria (≥65% POP, proper risk/reward)
- Provides clear thesis for each recommendation (≤30 words)

### Risk Management
- Enforces position size limits (max 0.5% portfolio risk per trade)
- Portfolio concentration limits (max 10% per symbol)
- Requires minimum probability of profit (65%)
- All trades require manual approval (no auto-execution)

### Market Data Integration
- Real-time quotes from Alpaca Markets
- Options chain data with full Greeks
- Technical indicators (SMA, RSI, MACD, etc.)
- Market statistics (VIX, sector performance)

## Installation

### Prerequisites
- Go 1.21+
- Node.js 18+
- Anthropic API key
- Alpaca Markets account (paper or live)

### Backend Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/vibetrade-claude.git
cd vibetrade-claude

# Install Go dependencies
go mod download

# Set environment variables
export ANTHROPIC_API_KEY="your-api-key"
export ALPACA_API_KEY="your-alpaca-key"
export ALPACA_SECRET_KEY="your-alpaca-secret"

# Run the server
go run cmd/unified_oauth_server/main.go
```

### Frontend Setup

```bash
# Install dependencies
cd frontend
npm install

# Start development server
npm run dev
```

## Usage

### Connecting Claude

1. Navigate to the AI Trading Assistant section in the dashboard
2. Click "Connect Claude Assistant"
3. Enter your Anthropic API key
4. Complete Turnstile verification
5. Your AI assistant is ready!

### Getting Trade Recommendations

The AI analyzes market data every 5 minutes and provides:
- Ticker symbol
- Strategy type (covered call, credit spread, etc.)
- Option legs description
- Investment thesis
- Probability of profit
- Risk/reward metrics

### Example Recommendation

```json
{
  "ticker": "SPY",
  "strategy": "Bull Put Credit Spread",
  "legs": "Sell 440P/Buy 435P exp 7d",
  "thesis": "SPY above 200-day MA with strong momentum, collect premium on pullback support",
  "pop": 0.72,
  "max_loss": 500,
  "max_profit": 150,
  "score": 0.85
}
```

## API Endpoints

### Claude Connection
- `POST /api/claude-code/connect` - Connect Claude API
- `GET /api/claude-code/status` - Check connection status
- `DELETE /api/claude-code/disconnect` - Remove connection

### Trading Features
- `GET /api/claude-code/recommendations` - Get AI trade recommendations
- `POST /api/claude-code/analyze-risk` - Analyze position risks
- `POST /api/claude-code/explain-strategy` - Get strategy explanations

## Safety Features

### Risk Limits
- Maximum 2% total portfolio risk
- Maximum 0.5% risk per position
- Minimum 65% probability of profit
- Minimum 1:3 risk/reward ratio

### Security
- API keys encrypted at rest
- Turnstile verification for all connections
- JWT authentication for all API calls
- No auto-execution of trades

### Disclaimers
- Clear "AI Generated - Not Financial Advice" labels
- Comprehensive risk disclosures
- Educational purpose disclaimers
- Manual approval required for all trades

## Architecture

```
vibetrade-claude/
├── internal/
│   └── ai_assistant/
│       ├── claude_client.go      # Claude API integration
│       ├── trading_assistant.go   # Trade analysis logic
│       ├── prompt_templates.go    # AI prompts
│       ├── market_data_aggregator.go # Data collection
│       └── risk_management.go     # Safety features
├── frontend/
│   └── ClaudeTradeAssistant.tsx  # React component
├── cmd/
│   └── unified_oauth_server/
│       ├── ai_handlers.go         # API endpoints
│       └── main_ai_routes.go      # Route registration
└── README.md
```

## Performance Tracking

The system tracks:
- AI recommendation success rate
- Win/loss ratio
- Average return per trade
- Risk-adjusted returns
- Comparison vs manual trading

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file

## Disclaimer

**IMPORTANT**: This software is for educational purposes only. AI-generated trading recommendations should not be considered financial advice. Always consult with a qualified financial advisor before making investment decisions. Options trading involves substantial risk of loss and is not suitable for all investors.

## Support

- Documentation: See `/docs` directory
- Issues: GitHub Issues
- Email: support@vibetrade.com

---

*Inspired by the Reddit trader who achieved 100% success rate with AI-assisted options trading*