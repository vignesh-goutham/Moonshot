package bot

import (
	"fmt"
	"log"
	"time"

	"moonshot/services"
	"moonshot/types"

	"github.com/shopspring/decimal"
)

// DCABot represents the main DCA bot
type DCABot struct {
	config          *types.BotConfig
	coinbaseService *services.CoinbaseService
	fngService      *services.FNGService
	portfolio       *types.Portfolio
}

// NewDCABot creates a new DCA bot instance
func NewDCABot(config *types.BotConfig, coinbaseService *services.CoinbaseService, fngService *services.FNGService) *DCABot {
	return &DCABot{
		config:          config,
		coinbaseService: coinbaseService,
		fngService:      fngService,
	}
}

// Execute runs the main DCA bot logic
func (b *DCABot) Execute() (*types.ExecutionResult, error) {
	log.Println("Starting Moonshot DCA bot execution...")

	// Get current portfolio
	portfolio, err := b.coinbaseService.GetPortfolio()
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	b.portfolio = portfolio

	// Get Fear & Greed Index
	fngIndex, err := b.fngService.GetFearGreedIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get FNG index: %w", err)
	}

	log.Printf("Current F&G Index: %d (%s), Multiplier: %s",
		fngIndex.Value, fngIndex.Classification, fngIndex.Multiplier.String())

	// Calculate investment decisions (buying only)
	decisions, err := b.calculateBuyDecisions(fngIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate investment decisions: %w", err)
	}

	// Execute decisions
	executionResult := &types.ExecutionResult{
		Success:   true,
		Decisions: decisions,
		Portfolio: portfolio,
		FNGIndex:  fngIndex,
		Timestamp: time.Now(),
	}

	totalInvested := decimal.Zero
	successfulOrders := 0
	failedOrders := 0

	for _, decision := range decisions {
		if decision.Action == "buy" {
			if err := b.executeBuyOrder(decision); err != nil {
				log.Printf("❌ Failed to execute buy order: %v", err)
				executionResult.Success = false
				executionResult.Error = err.Error()
				failedOrders++
			} else {
				totalInvested = totalInvested.Add(decision.Amount)
				successfulOrders++
			}
		}
	}

	// Log execution summary
	if successfulOrders > 0 {
		log.Printf("✅ Successfully placed %d orders, total invested: $%s USD",
			successfulOrders, totalInvested.String())
	}
	if failedOrders > 0 {
		log.Printf("❌ Failed to place %d orders", failedOrders)
	}

	executionResult.TotalInvested = totalInvested
	executionResult.TotalSold = decimal.Zero // No selling

	log.Printf("Execution completed. Invested: %s USDC", totalInvested.String())

	return executionResult, nil
}

// calculateBuyDecisions calculates buy orders based on F&G index
func (b *DCABot) calculateBuyDecisions(fngIndex *types.FearGreedIndex) ([]types.InvestmentDecision, error) {
	var decisions []types.InvestmentDecision

	// Calculate available investment amount
	availableUSDC := b.portfolio.USDCBalance
	baseInvestment := b.config.WeeklyBaseInvestment

	// Apply F&G multiplier to determine investment amount
	investmentAmount := baseInvestment.Mul(fngIndex.Multiplier)

	// Calculate dynamic buffer based on market sentiment
	dynamicBuffer := b.calculateDynamicBuffer(fngIndex.Value)

	// Ensure we don't exceed available USDC (leave dynamic buffer for dip buying)
	maxInvestment := availableUSDC.Mul(decimal.NewFromFloat(1.0).Sub(dynamicBuffer))
	if investmentAmount.GreaterThan(maxInvestment) {
		investmentAmount = maxInvestment
	}

	if investmentAmount.LessThanOrEqual(decimal.Zero) {
		log.Println("No USDC available for investment")
		return decisions, nil
	}

	// Calculate allocations for BTC and ETH
	btcInvestment := investmentAmount.Mul(b.config.BTCAllocation).Div(decimal.NewFromInt(100))
	ethInvestment := investmentAmount.Mul(b.config.ETHAllocation).Div(decimal.NewFromInt(100))

	// Get current prices and create buy decisions
	btcPrice, err := b.getAssetPrice("BTC")
	if err == nil && btcInvestment.GreaterThan(decimal.Zero) {
		decisions = append(decisions, types.InvestmentDecision{
			Asset:     "BTC",
			Action:    "buy",
			Amount:    btcInvestment,
			Price:     btcPrice,
			Reason:    fmt.Sprintf("DCA with F&G multiplier %s (Index: %d)", fngIndex.Multiplier.String(), fngIndex.Value),
			Timestamp: time.Now(),
		})
	}

	ethPrice, err := b.getAssetPrice("ETH")
	if err == nil && ethInvestment.GreaterThan(decimal.Zero) {
		decisions = append(decisions, types.InvestmentDecision{
			Asset:     "ETH",
			Action:    "buy",
			Amount:    ethInvestment,
			Price:     ethPrice,
			Reason:    fmt.Sprintf("DCA with F&G multiplier %s (Index: %d)", fngIndex.Multiplier.String(), fngIndex.Value),
			Timestamp: time.Now(),
		})
	}

	log.Printf("Investment decisions: BTC: %s USDC, ETH: %s USDC", btcInvestment.String(), ethInvestment.String())
	log.Printf("Dynamic buffer: %s%% (F&G: %d)", dynamicBuffer.Mul(decimal.NewFromInt(100)).String(), fngIndex.Value)

	return decisions, nil
}

// calculateDynamicBuffer calculates buffer percentage based on F&G index
func (b *DCABot) calculateDynamicBuffer(fngValue int) decimal.Decimal {
	// F&G ranges from 0-100
	switch {
	case fngValue <= 20: // Extreme Fear
		// No buffer - go all in when markets are fearful
		return decimal.Zero

	case fngValue <= 40: // Fear
		// Small buffer - 5% to 10%
		ratio := decimal.NewFromInt(int64(40 - fngValue)).Div(decimal.NewFromInt(20))
		return decimal.NewFromFloat(0.05).Add(ratio.Mul(decimal.NewFromFloat(0.05)))

	case fngValue <= 60: // Neutral
		// Normal buffer - 15% to 20%
		ratio := decimal.NewFromInt(int64(60 - fngValue)).Div(decimal.NewFromInt(20))
		return decimal.NewFromFloat(0.15).Add(ratio.Mul(decimal.NewFromFloat(0.05)))

	case fngValue <= 80: // Greed
		// Higher buffer - 20% (capped)
		return decimal.NewFromFloat(0.20)

	default: // Extreme Greed (81-100)
		// Maximum buffer - 20% (capped)
		return decimal.NewFromFloat(0.20)
	}
}

// executeBuyOrder executes a buy order
func (b *DCABot) executeBuyOrder(decision types.InvestmentDecision) error {
	productID := decision.Asset + "-USDC"

	// Check if we have sufficient USDC balance
	if b.portfolio.USDCBalance.LessThan(decision.Amount) {
		return fmt.Errorf("insufficient USDC balance: have %s, need %s",
			b.portfolio.USDCBalance.String(), decision.Amount.String())
	}

	// Calculate size in asset units
	size := decision.Amount.Div(decision.Price)

	// Log the investment details
	log.Printf("💰 Investing $%s USD in %s (%.6f %s at $%s per %s)",
		decision.Amount.String(),
		decision.Asset,
		size.InexactFloat64(),
		decision.Asset,
		decision.Price.String(),
		decision.Asset)

	log.Printf("📊 Placing BUY order: %.6f %s at market price", size.InexactFloat64(), decision.Asset)

	orderResp, err := b.coinbaseService.PlaceOrder(productID, "BUY", "market", size.String(), "")
	if err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}

	// Log order result
	if orderResp.Success {
		log.Printf("✅ Order placed successfully! Order ID: %s", orderResp.OrderId)
	} else {
		log.Printf("❌ Order failed: %s", orderResp.FailureReason)
		return fmt.Errorf("order failed: %s", orderResp.FailureReason)
	}

	return nil
}

// getAssetPrice gets the current price of an asset
func (b *DCABot) getAssetPrice(symbol string) (decimal.Decimal, error) {
	productID := symbol + "-USDC"
	product, err := b.coinbaseService.GetProduct(productID)
	if err != nil {
		return decimal.Zero, err
	}

	price, err := decimal.NewFromString(product.Price)
	if err != nil {
		return decimal.Zero, err
	}

	return price, nil
}

// GetPortfolio returns the current portfolio
func (b *DCABot) GetPortfolio() *types.Portfolio {
	return b.portfolio
}

// GetConfig returns the bot configuration
func (b *DCABot) GetConfig() *types.BotConfig {
	return b.config
}
