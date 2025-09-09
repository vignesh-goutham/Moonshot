package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"moonshot/bot"
	"moonshot/services"
	"moonshot/types"

	"github.com/aws/aws-lambda-go/lambda"
)

// LambdaResponse represents the Lambda response
type LambdaResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Global bot instance
var dcaBot *bot.DCABot

// init initializes the bot when Lambda container starts
func init() {
	log.Println("Initializing Moonshot DCA Bot...")

	// Load configuration from environment variables
	botConfig, coinbaseConfig, err := loadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := validateConfig(botConfig, coinbaseConfig); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("BTC Allocation: %s%%", botConfig.BTCAllocation.String())
	log.Printf("ETH Allocation: %s%%", botConfig.ETHAllocation.String())
	log.Printf("Weekly Base Investment: %s USDC", botConfig.WeeklyBaseInvestment.String())

	// Initialize services
	coinbaseService := services.NewCoinbaseService(coinbaseConfig)
	fngService := services.NewFNGService("https://api.alternative.me/fng/")

	// Initialize bot
	dcaBot = bot.NewDCABot(botConfig, coinbaseService, fngService)

	log.Println("Moonshot DCA Bot initialized successfully")
}

// handleRequest handles EventBridge scheduler triggers
func handleRequest(ctx context.Context, event interface{}) (LambdaResponse, error) {
	log.Println("EventBridge trigger received - executing DCA bot...")

	// Execute the DCA bot logic
	result, err := dcaBot.Execute()
	if err != nil {
		log.Printf("Bot execution failed: %v", err)
		return LambdaResponse{
			Success:   false,
			Message:   "Bot execution failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	log.Printf("Bot execution completed successfully")
	log.Printf("Total invested: %s USDC", result.TotalInvested.String())
	log.Printf("Total sold: %s USDC", result.TotalSold.String())

	return LambdaResponse{
		Success:   result.Success,
		Message:   "DCA execution completed successfully",
		Data:      result,
		Timestamp: time.Now(),
	}, nil
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() (*types.BotConfig, *types.CoinbaseConfig, error) {
	// Load bot configuration
	botConfig := &types.BotConfig{}

	// Asset allocations
	btcAlloc := getEnvFloat("BTC_ALLOCATION", 80.0)
	ethAlloc := getEnvFloat("ETH_ALLOCATION", 20.0)

	// Validate allocations sum to 100
	if btcAlloc+ethAlloc != 100.0 {
		return nil, nil, fmt.Errorf("BTC and ETH allocations must sum to 100, got %.1f + %.1f", btcAlloc, ethAlloc)
	}

	botConfig.BTCAllocation = types.DecimalFromFloat(btcAlloc)
	botConfig.ETHAllocation = types.DecimalFromFloat(ethAlloc)
	botConfig.WeeklyBaseInvestment = types.DecimalFromFloat(getEnvFloat("WEEKLY_BASE_INVESTMENT", 100.0))

	botConfig.FNGBuyThreshold = getEnvInt("FNG_BUY_THRESHOLD", 25)
	botConfig.MinMultiplier = types.DecimalFromFloat(getEnvFloat("MIN_MULTIPLIER", 0.5))
	botConfig.MaxMultiplier = types.DecimalFromFloat(getEnvFloat("MAX_MULTIPLIER", 2.0))
	botConfig.InvestmentFrequency = getEnvString("INVESTMENT_FREQUENCY", "weekly")
	botConfig.ExecutionTime = getEnvString("EXECUTION_TIME", "09:00")

	// Load Coinbase configuration using the new credential loading method
	creds, err := services.LoadCredentialsFromEnv()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Coinbase credentials: %w", err)
	}

	coinbaseConfig := &types.CoinbaseConfig{
		APIKey:    creds.AccessKey,
		APISecret: creds.PrivatePemKey,
		Sandbox:   getEnvBool("COINBASE_SANDBOX", false),
	}

	return botConfig, coinbaseConfig, nil
}

// validateConfig validates the loaded configuration
func validateConfig(botConfig *types.BotConfig, coinbaseConfig *types.CoinbaseConfig) error {
	// Validate bot configuration
	if botConfig.WeeklyBaseInvestment.LessThanOrEqual(types.DecimalZero()) {
		return fmt.Errorf("weekly base investment must be positive")
	}

	if botConfig.FNGBuyThreshold < 0 || botConfig.FNGBuyThreshold > 100 {
		return fmt.Errorf("FNG buy threshold must be between 0 and 100")
	}

	if botConfig.MinMultiplier.LessThanOrEqual(types.DecimalZero()) {
		return fmt.Errorf("minimum multiplier must be positive")
	}

	if botConfig.MaxMultiplier.LessThanOrEqual(types.DecimalZero()) {
		return fmt.Errorf("maximum multiplier must be positive")
	}

	if botConfig.MinMultiplier.GreaterThan(botConfig.MaxMultiplier) {
		return fmt.Errorf("minimum multiplier cannot be greater than maximum multiplier")
	}

	return nil
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := types.StringToInt(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := types.StringToFloat(value); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := types.StringToBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// main function for Lambda
func main() {
	lambda.Start(handleRequest)
}
