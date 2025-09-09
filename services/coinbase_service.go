package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"moonshot/types"

	"github.com/coinbase-samples/advanced-trade-sdk-go/accounts"
	"github.com/coinbase-samples/advanced-trade-sdk-go/client"
	"github.com/coinbase-samples/advanced-trade-sdk-go/credentials"
	"github.com/coinbase-samples/advanced-trade-sdk-go/model"
	"github.com/coinbase-samples/advanced-trade-sdk-go/orders"
	"github.com/coinbase-samples/advanced-trade-sdk-go/products"
	"github.com/shopspring/decimal"
)

// CoinbaseService handles Coinbase Advanced API operations using the official SDK
type CoinbaseService struct {
	config     *types.CoinbaseConfig
	restClient client.RestClient
}

// NewCoinbaseService creates a new Coinbase service instance using the official SDK
func NewCoinbaseService(config *types.CoinbaseConfig) *CoinbaseService {
	// Create credentials struct for the official SDK
	creds := &credentials.Credentials{
		AccessKey:     config.APIKey,
		PrivatePemKey: config.APISecret,
	}

	// Create HTTP client
	httpClient, err := client.DefaultHttpClient()
	if err != nil {
		panic(fmt.Sprintf("unable to load default http client: %v", err))
	}

	// Create REST client
	restClient := client.NewRestClient(creds, httpClient)

	return &CoinbaseService{
		config:     config,
		restClient: restClient,
	}
}

// GetAccounts fetches all accounts using the official SDK
func (c *CoinbaseService) GetAccounts() ([]*model.Account, error) {
	accountsService := accounts.NewAccountsService(c.restClient)

	resp, err := accountsService.ListAccounts(context.Background(), &accounts.ListAccountsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return resp.Accounts, nil
}

// GetProduct fetches product information using the official SDK
func (c *CoinbaseService) GetProduct(productID string) (*products.GetProductResponse, error) {
	productsService := products.NewProductsService(c.restClient)

	resp, err := productsService.GetProduct(context.Background(), &products.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return resp, nil
}

// GetProductBook fetches product book information for pricing
func (c *CoinbaseService) GetProductBook(productID string) (*products.GetProductBookResponse, error) {
	productsService := products.NewProductsService(c.restClient)

	resp, err := productsService.GetProductBook(context.Background(), &products.GetProductBookRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get product book: %w", err)
	}

	return resp, nil
}

// PlaceOrder places a new order using the official SDK
func (c *CoinbaseService) PlaceOrder(productID, side, orderType, size, price string) (*orders.CreateOrderResponse, error) {
	ordersService := orders.NewOrdersService(c.restClient)

	// Create order configuration based on order type
	var orderConfig model.OrderConfiguration

	if orderType == "market" {
		orderConfig = model.OrderConfiguration{
			MarketMarketIoc: &model.MarketIoc{
				QuoteSize: size,
			},
		}
	} else if orderType == "limit" && price != "" {
		orderConfig = model.OrderConfiguration{
			LimitLimitGtc: &model.LimitGtc{
				BaseSize:   size,
				LimitPrice: price,
			},
		}
	} else {
		return nil, fmt.Errorf("unsupported order type: %s", orderType)
	}

	resp, err := ordersService.CreateOrder(context.Background(), &orders.CreateOrderRequest{
		ProductId:          productID,
		Side:               side,
		OrderConfiguration: orderConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return resp, nil
}

// GetPortfolio fetches current portfolio information using the official SDK
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
		balance, _ := decimal.NewFromString(account.AvailableBalance.Value)
		if balance.IsZero() {
			continue
		}

		if account.Currency == "USDC" {
			portfolio.USDCBalance = balance
			totalValue = totalValue.Add(balance)
		} else if account.Currency == "BTC" || account.Currency == "ETH" {
			// Get current price from product book
			productID := account.Currency + "-USDC"
			productBookResp, err := c.GetProductBook(productID)
			if err != nil {
				continue
			}

			// Use the best bid price as current price
			var price decimal.Decimal
			if productBookResp.PriceBook != nil && len(productBookResp.PriceBook.Bids) > 0 {
				price, _ = decimal.NewFromString(productBookResp.PriceBook.Bids[0].Price)
			} else if productBookResp.PriceBook != nil && len(productBookResp.PriceBook.Asks) > 0 {
				price, _ = decimal.NewFromString(productBookResp.PriceBook.Asks[0].Price)
			} else {
				continue
			}

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

// LoadCredentialsFromEnv loads credentials from environment variables in the format expected by the official SDK
func LoadCredentialsFromEnv() (*credentials.Credentials, error) {
	// Check if credentials are provided as a JSON string
	credentialsJSON := os.Getenv("COINBASE_CREDENTIALS_JSON")
	if credentialsJSON != "" {
		var creds credentials.Credentials
		if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials JSON: %w", err)
		}

		// Validate that we have the required fields
		if creds.AccessKey == "" || creds.PrivatePemKey == "" {
			return nil, fmt.Errorf("credentials JSON must contain 'accessKey' and 'privatePemKey' fields")
		}

		return &creds, nil
	}

	// Fallback to individual environment variables
	accessKey := os.Getenv("COINBASE_API_KEY")
	privatePemKey := os.Getenv("COINBASE_API_SECRET")

	if accessKey == "" || privatePemKey == "" {
		return nil, fmt.Errorf(`COINBASE_API_KEY and COINBASE_API_SECRET environment variables are required.

For Coinbase Advanced Trade API, you need to:
1. Go to https://portal.cdp.coinbase.com/
2. Create a new application
3. Generate API credentials
4. The private key should be in PEM format (starts with "-----BEGIN EC PRIVATE KEY-----")

Alternatively, you can use COINBASE_CREDENTIALS_JSON with the format:
{"accessKey":"your_access_key","privatePemKey":"your_pem_key"}`)
	}

	// Validate that the private key looks like a PEM key
	if !isValidPEMKey(privatePemKey) {
		return nil, fmt.Errorf(`invalid PEM key format. The COINBASE_API_SECRET should be a PEM-formatted private key.

Expected format:
-----BEGIN EC PRIVATE KEY-----
[base64 encoded key data]
-----END EC PRIVATE KEY-----

For environment variables with newlines, use the JSON format instead:
COINBASE_CREDENTIALS_JSON={"accessKey":"your_key","privatePemKey":"-----BEGIN EC PRIVATE KEY-----\\nyour_key_data\\n-----END EC PRIVATE KEY-----"}

Please generate new credentials from https://portal.cdp.coinbase.com/`)
	}

	return &credentials.Credentials{
		AccessKey:     accessKey,
		PrivatePemKey: privatePemKey,
	}, nil
}

// isValidPEMKey checks if the provided string looks like a valid PEM key
func isValidPEMKey(key string) bool {
	return len(key) > 50 &&
		(strings.Contains(key, "-----BEGIN") && strings.Contains(key, "-----END"))
}
