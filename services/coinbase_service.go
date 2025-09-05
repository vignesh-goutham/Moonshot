package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"moonshot/types"

	"github.com/shopspring/decimal"
)

// CoinbaseService handles Coinbase Advanced API operations
type CoinbaseService struct {
	config  *types.CoinbaseConfig
	client  *http.Client
	baseURL string
}

// CoinbaseAccount represents a Coinbase account
type CoinbaseAccount struct {
	ID        string `json:"id"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
	Available string `json:"available"`
	Hold      string `json:"hold"`
}

// CoinbaseProduct represents a trading product
type CoinbaseProduct struct {
	ID            string `json:"id"`
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	Price         string `json:"price"`
	Status        string `json:"status"`
}

// CoinbaseOrder represents an order
type CoinbaseOrder struct {
	ID         string `json:"order_id"`
	ProductID  string `json:"product_id"`
	Side       string `json:"side"`
	OrderType  string `json:"order_type"`
	Size       string `json:"size"`
	FilledSize string `json:"filled_size"`
	Price      string `json:"price"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// NewCoinbaseService creates a new Coinbase service instance
func NewCoinbaseService(config *types.CoinbaseConfig) *CoinbaseService {
	baseURL := "https://api.exchange.coinbase.com"
	if config.Sandbox {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
	}

	return &CoinbaseService{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

// GetAccounts fetches all accounts
func (c *CoinbaseService) GetAccounts() ([]CoinbaseAccount, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/accounts", nil)
	if err != nil {
		return nil, err
	}

	c.addAuthHeaders(req, "GET", "/accounts", "")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var accounts []CoinbaseAccount
	if err := json.Unmarshal(body, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	return accounts, nil
}

// GetProduct fetches product information
func (c *CoinbaseService) GetProduct(productID string) (*CoinbaseProduct, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/products/"+productID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var product CoinbaseProduct
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &product, nil
}

// PlaceOrder places a new order
func (c *CoinbaseService) PlaceOrder(productID, side, orderType, size, price string) (*CoinbaseOrder, error) {
	orderData := map[string]string{
		"product_id": productID,
		"side":       side,
		"order_type": orderType,
		"size":       size,
	}

	if price != "" {
		orderData["price"] = price
	}

	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/orders", nil)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	c.addAuthHeaders(req, "POST", "/orders", string(jsonData))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var order CoinbaseOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// GetPortfolio fetches current portfolio information
func (c *CoinbaseService) GetPortfolio() (*types.Portfolio, error) {
	accounts, err := c.GetAccounts()
	if err != nil {
		return nil, err
	}

	portfolio := &types.Portfolio{
		Assets:      make(map[string]*types.Asset),
		LastUpdated: time.Now(),
	}

	totalValue := decimal.Zero

	for _, account := range accounts {
		balance, _ := decimal.NewFromString(account.Balance)
		if balance.IsZero() {
			continue
		}

		if account.Currency == "USDC" {
			portfolio.USDCBalance = balance
			totalValue = totalValue.Add(balance)
		} else if account.Currency == "BTC" || account.Currency == "ETH" {
			// Get current price
			productID := account.Currency + "-USDC"
			product, err := c.GetProduct(productID)
			if err != nil {
				continue
			}

			price, _ := decimal.NewFromString(product.Price)
			assetValue := balance.Mul(price)
			totalValue = totalValue.Add(assetValue)

			portfolio.Assets[account.Currency] = &types.Asset{
				Symbol:  account.Currency,
				Name:    account.Currency,
				Balance: balance,
				Price:   price,
				Value:   assetValue,
			}
		}
	}

	portfolio.TotalValue = totalValue
	return portfolio, nil
}

// addAuthHeaders adds Coinbase authentication headers
func (c *CoinbaseService) addAuthHeaders(req *http.Request, method, path, body string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + method + path + body

	h := hmac.New(sha256.New, []byte(c.config.APISecret))
	h.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	req.Header.Set("CB-ACCESS-KEY", c.config.APIKey)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("CB-ACCESS-PASSPHRASE", c.config.Passphrase)
}
