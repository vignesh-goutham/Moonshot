package types

import (
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// Asset represents a cryptocurrency asset
type Asset struct {
	Symbol     string          `json:"symbol"`
	Name       string          `json:"name"`
	Allocation decimal.Decimal `json:"allocation"`
	Balance    decimal.Decimal `json:"balance"`
	Price      decimal.Decimal `json:"price"`
	Value      decimal.Decimal `json:"value"`
}

// Portfolio represents the current portfolio state
type Portfolio struct {
	TotalValue  decimal.Decimal   `json:"total_value"`
	USDCBalance decimal.Decimal   `json:"usdc_balance"`
	Assets      map[string]*Asset `json:"assets"`
	LastUpdated time.Time         `json:"last_updated"`
}

// FearGreedIndex represents the Fear & Greed Index data
type FearGreedIndex struct {
	Value          int             `json:"value"`
	Classification string          `json:"classification"`
	Timestamp      time.Time       `json:"timestamp"`
	Multiplier     decimal.Decimal `json:"multiplier"`
}

// InvestmentDecision represents a decision made by the bot
type InvestmentDecision struct {
	Asset     string          `json:"asset"`
	Action    string          `json:"action"` // buy only
	Amount    decimal.Decimal `json:"amount"`
	Price     decimal.Decimal `json:"price"`
	Reason    string          `json:"reason"`
	Timestamp time.Time       `json:"timestamp"`
}

// MarketData represents current market information
type MarketData struct {
	Symbol    string          `json:"symbol"`
	Price     decimal.Decimal `json:"price"`
	Volume24h decimal.Decimal `json:"volume_24h"`
	Change24h decimal.Decimal `json:"change_24h"`
	Timestamp time.Time       `json:"timestamp"`
}

// BotConfig represents the bot configuration
type BotConfig struct {
	BTCAllocation        decimal.Decimal `json:"btc_allocation"`
	ETHAllocation        decimal.Decimal `json:"eth_allocation"`
	WeeklyBaseInvestment decimal.Decimal `json:"weekly_base_investment"`
	FNGBuyThreshold      int             `json:"fng_buy_threshold"`
	MinMultiplier        decimal.Decimal `json:"min_multiplier"`
	MaxMultiplier        decimal.Decimal `json:"max_multiplier"`
	InvestmentFrequency  string          `json:"investment_frequency"`
	ExecutionTime        string          `json:"execution_time"`
}

// CoinbaseConfig represents Coinbase Advanced API configuration
type CoinbaseConfig struct {
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	Passphrase string `json:"passphrase"`
	Sandbox    bool   `json:"sandbox"`
}

// ExecutionResult represents the result of a bot execution
type ExecutionResult struct {
	Success       bool                 `json:"success"`
	Decisions     []InvestmentDecision `json:"decisions"`
	Portfolio     *Portfolio           `json:"portfolio"`
	FNGIndex      *FearGreedIndex      `json:"fng_index"`
	TotalInvested decimal.Decimal      `json:"total_invested"`
	TotalSold     decimal.Decimal      `json:"total_sold"`
	Timestamp     time.Time            `json:"timestamp"`
	Error         string               `json:"error,omitempty"`
}

// Helper functions for decimal operations
func DecimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

func DecimalZero() decimal.Decimal {
	return decimal.Zero
}

// Helper functions for type conversions
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func StringToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
