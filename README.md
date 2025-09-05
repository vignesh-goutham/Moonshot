# Moonshot - Dynamic DCA Crypto Bot

Moonshot is an intelligent Dollar Cost Averaging (DCA) bot for cryptocurrency that dynamically adjusts investment amounts based on market sentiment using the Fear & Greed Index. The bot automatically manages a portfolio of BTC (80%) and ETH (20%) through Coinbase Advanced.

## Features

- **Dynamic DCA**: Investment amounts automatically adjust based on market sentiment
- **Fear & Greed Index Integration**: Uses market sentiment to determine buy amounts
- **Smart Portfolio Management**: Automatically maintains 80% BTC / 20% ETH allocation
- **Dip Buying Strategy**: Reserves funds for buying during market dips
- **Coinbase Advanced Integration**: Direct trading through Coinbase's advanced trading platform
- **Lambda Ready**: Deploys to AWS Lambda for automated execution
- **Environment Variables**: Simple configuration through environment variables

## How It Works

### Investment Multiplier System
The bot uses the Fear & Greed Index to calculate investment multipliers:

- **Extreme Fear (0-20)**: 1.5x - 2.0x multiplier (buy more)
- **Fear (21-40)**: 1.2x - 1.5x multiplier (buy more)
- **Neutral (41-60)**: 1.0x multiplier (normal DCA)
- **Greed (61-80)**: 0.7x - 1.0x multiplier (buy less)
- **Extreme Greed (81-100)**: 0.5x - 0.7x multiplier (buy less)

### Dynamic Buffer System
The bot automatically adjusts how much cash to reserve based on market sentiment:

- **Extreme Fear (0-20)**: **0% buffer** - Go all in when markets are fearful
- **Fear (21-40)**: **5-10% buffer** - Small reserve, still aggressive buying
- **Neutral (41-60)**: **15-20% buffer** - Normal reserve for standard DCA
- **Greed (61+)**: **20% buffer** - Conservative reserve, capped maximum

This ensures optimal buying during market crashes while maintaining reasonable protection during euphoric periods.

### Pure DCA Strategy
- **No Selling**: The bot only buys, never sells
- **Automatic Allocation**: Maintains target portfolio balance through DCA
- **Dynamic Buffer System**: Automatically adjusts cash reserves based on market sentiment
- **EventBridge Triggered**: Runs automatically on schedule via AWS Lambda

## Prerequisites

- Go 1.21 or higher
- Coinbase Advanced account with API access
- USDC balance for trading
- AWS account for Lambda deployment

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd moonshot
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the bot:
```bash
make build
```

## Configuration

The bot uses environment variables for configuration. Copy the template and fill in your values:

```bash
cp env.template .env
# Edit .env with your Coinbase API credentials
```

### Required Environment Variables
```bash
# Coinbase Advanced API Credentials
COINBASE_API_KEY=your_api_key
COINBASE_API_SECRET=your_api_secret
COINBASE_PASSPHRASE=your_passphrase
```

### Optional Environment Variables (with defaults)
```bash
# Asset allocation (must sum to 100)
BTC_ALLOCATION=80.0
ETH_ALLOCATION=20.0

# Investment parameters
WEEKLY_BASE_INVESTMENT=100.0  # Base weekly investment in USDC

# Fear & Greed Index thresholds
FNG_BUY_THRESHOLD=20          # Buy more when F&G < 20

# Multiplier ranges
MIN_MULTIPLIER=0.5            # Minimum investment multiplier
MAX_MULTIPLIER=2.0            # Maximum investment multiplier

# AWS Lambda settings
LAMBDA_REGION=us-east-2
LAMBDA_NAME=moonshot-dca-bot
```

### Coinbase API Setup

1. Log into your Coinbase Advanced account
2. Go to API settings
3. Create a new API key with the following permissions:
   - View account information
   - Place orders
   - View orders
4. Save your API key, secret, and passphrase

## Usage

### Local Testing
Test the bot locally:
```bash
make test-lambda
```

### Deploy to AWS Lambda
Deploy the bot to AWS Lambda:
```bash
make deploy-lambda
```

### Command Line Options

- `make build` - Build the Lambda function
- `make package-lambda` - Package for deployment
- `make deploy-lambda` - Deploy to AWS Lambda
- `make help` - View all available commands

## AWS Lambda Setup

### EventBridge Rule
Create an EventBridge rule to trigger the Lambda function:

```json
{
  "schedule": "cron(0 9 ? * MON *)",  // Every Monday at 9 AM UTC
  "target": {
    "arn": "arn:aws:lambda:us-east-2:...",
    "id": "MoonshotDCATrigger"
  }
}
```

### IAM Role
Ensure your Lambda execution role has:
- Basic Lambda execution permissions
- CloudWatch Logs permissions

## Safety Features

- **Buy-Only Strategy**: No selling, only accumulating assets
- **Dip Buying Buffer**: Always reserves funds for buying during dips
- **Configurable Thresholds**: Adjustable buy triggers
- **Portfolio Limits**: Prevents over-investment
- **Error Handling**: Graceful failure handling and logging

## Monitoring

The bot provides comprehensive logging:
- Portfolio status updates
- Fear & Greed Index values
- Investment decisions and reasoning
- Trade execution results
- Error reporting

## Risk Management

- **Never invest more than you can afford to lose**
- Start with small amounts to test the system
- Monitor the bot's performance regularly
- Adjust parameters based on your risk tolerance
- The bot only buys, so you control when to sell

## Disclaimer

This software is for educational and informational purposes only. Cryptocurrency trading involves substantial risk and may result in the loss of your invested capital. The authors are not responsible for any financial losses incurred through the use of this software.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Open an issue on GitHub
- Check the configuration examples
- Review the logs for error details

## Roadmap

- [ ] Additional technical indicators
- [ ] Portfolio rebalancing strategies
- [ ] Web dashboard
- [ ] Mobile app
- [ ] Advanced risk management
- [ ] Multi-exchange support
