# Moonshot - Dynamic DCA Crypto Bot

Moonshot is an intelligent Dollar Cost Averaging (DCA) bot for cryptocurrency that dynamically adjusts investment amounts based on market sentiment using the Fear & Greed Index. The bot automatically manages a portfolio of BTC (80%) and ETH (20%) through Coinbase Advanced.

## Features

- **Dynamic DCA**: Investment amounts automatically adjust based on market sentiment
- **Fear & Greed Index Integration**: Uses market sentiment to determine buy amounts
- **Smart Portfolio Management**: Automatically maintains 80% BTC / 20% ETH allocation
- **Dynamic Buffer System**: Automatically adjusts cash buffer based on market conditions
- **Official Coinbase SDK**: Uses the official Coinbase Advanced Trade SDK for reliable API integration
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

**IMPORTANT**: The Coinbase Advanced Trade API requires different credentials than the old Coinbase Pro API. You need to generate credentials through the Coinbase Developer Platform.

1. Go to [Coinbase Developer Platform](https://portal.cdp.coinbase.com/)
2. Create a new application
3. Generate API credentials with the following permissions:
   - View account information
   - Place orders
   - View orders
4. Copy your **API Key ID** and **Private PEM Key** (not a simple secret!)
5. Set the environment variables using one of these methods:

**Method 1: JSON format (RECOMMENDED)**
```bash
COINBASE_CREDENTIALS_JSON={"accessKey":"your_api_key_id","privatePemKey":"-----BEGIN EC PRIVATE KEY-----\nyour_key_here\n-----END EC PRIVATE KEY-----"}
```

**Method 2: Individual credentials (fallback for local development)**
```bash
COINBASE_API_KEY=your_api_key_id
COINBASE_API_SECRET=-----BEGIN EC PRIVATE KEY-----
your_base64_encoded_private_key_here
-----END EC PRIVATE KEY-----
```

**Important Notes**:
- The private key must be in PEM format (starts with `-----BEGIN EC PRIVATE KEY-----`)
- For AWS Lambda, use the JSON format because environment variables are single-line strings
- In the JSON format, use `\n` for newlines in the `privatePemKey` field
- If you're getting "failed to parse PEM block" errors, you're likely using the wrong credential format or need to use the JSON format

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
