package config

import (
	"fmt"
	"moonshot/types"

	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*types.BotConfig, *types.CoinbaseConfig, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Load bot configuration
	botConfig := &types.BotConfig{}

	btcAlloc := viper.GetFloat64("bot.btc_allocation")
	ethAlloc := viper.GetFloat64("bot.eth_allocation")

	// Validate allocations sum to 100
	if btcAlloc+ethAlloc != 100.0 {
		return nil, nil, fmt.Errorf("BTC and ETH allocations must sum to 100, got %.1f + %.1f", btcAlloc, ethAlloc)
	}

	botConfig.BTCAllocation = decimal.NewFromFloat(btcAlloc)
	botConfig.ETHAllocation = decimal.NewFromFloat(ethAlloc)
	botConfig.WeeklyBaseInvestment = decimal.NewFromFloat(viper.GetFloat64("bot.weekly_base_investment"))
	botConfig.DipBuyingBuffer = decimal.NewFromFloat(viper.GetFloat64("bot.dip_buying_buffer"))
	botConfig.FNGBuyThreshold = viper.GetInt("bot.fng_buy_threshold")
	botConfig.FNGSellThreshold = viper.GetInt("bot.fng_sell_threshold")
	botConfig.MinMultiplier = decimal.NewFromFloat(viper.GetFloat64("bot.min_multiplier"))
	botConfig.MaxMultiplier = decimal.NewFromFloat(viper.GetFloat64("bot.max_multiplier"))
	botConfig.SellPercentage = decimal.NewFromFloat(viper.GetFloat64("bot.sell_percentage"))
	botConfig.InvestmentFrequency = viper.GetString("bot.investment_frequency")
	botConfig.ExecutionTime = viper.GetString("bot.execution_time")

	// Load Coinbase configuration
	coinbaseConfig := &types.CoinbaseConfig{
		APIKey:     viper.GetString("coinbase.api_key"),
		APISecret:  viper.GetString("coinbase.api_secret"),
		Passphrase: viper.GetString("coinbase.passphrase"),
		Sandbox:    viper.GetBool("coinbase.sandbox"),
	}

	// Validate required Coinbase credentials
	if coinbaseConfig.APIKey == "" || coinbaseConfig.APISecret == "" || coinbaseConfig.Passphrase == "" {
		return nil, nil, fmt.Errorf("Coinbase API credentials are required")
	}

	return botConfig, coinbaseConfig, nil
}

// ValidateConfig validates the loaded configuration
func ValidateConfig(botConfig *types.BotConfig, coinbaseConfig *types.CoinbaseConfig) error {
	// Validate bot configuration
	if botConfig.WeeklyBaseInvestment.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("weekly base investment must be positive")
	}

	if botConfig.DipBuyingBuffer.LessThan(decimal.Zero) || botConfig.DipBuyingBuffer.GreaterThan(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("dip buying buffer must be between 0 and 1")
	}

	if botConfig.FNGBuyThreshold < 0 || botConfig.FNGBuyThreshold > 100 {
		return fmt.Errorf("FNG buy threshold must be between 0 and 100")
	}

	if botConfig.FNGSellThreshold < 0 || botConfig.FNGSellThreshold > 100 {
		return fmt.Errorf("FNG sell threshold must be between 0 and 100")
	}

	if botConfig.MinMultiplier.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("minimum multiplier must be positive")
	}

	if botConfig.MaxMultiplier.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("maximum multiplier must be positive")
	}

	if botConfig.MinMultiplier.GreaterThan(botConfig.MaxMultiplier) {
		return fmt.Errorf("minimum multiplier cannot be greater than maximum multiplier")
	}

	if botConfig.SellPercentage.LessThan(decimal.Zero) || botConfig.SellPercentage.GreaterThan(decimal.NewFromFloat(1.0)) {
		return fmt.Errorf("sell percentage must be between 0 and 1")
	}

	return nil
}
