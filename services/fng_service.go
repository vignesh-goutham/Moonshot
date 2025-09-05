package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"moonshot/types"

	"github.com/shopspring/decimal"
)

// FNGService handles Fear & Greed Index operations
type FNGService struct {
	apiURL string
	client *http.Client
}

// FNGResponse represents the API response from alternative.me
type FNGResponse struct {
	Name string `json:"name"`
	Data []struct {
		Value          string `json:"value"`
		Classification string `json:"value_classification"`
		Timestamp      string `json:"timestamp"`
	} `json:"data"`
}

// NewFNGService creates a new FNG service instance
func NewFNGService(apiURL string) *FNGService {
	return &FNGService{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetFearGreedIndex fetches the current Fear & Greed Index
func (f *FNGService) GetFearGreedIndex() (*types.FearGreedIndex, error) {
	resp, err := f.client.Get(f.apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch FNG index: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var fngResp FNGResponse
	if err := json.Unmarshal(body, &fngResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FNG response: %w", err)
	}

	if len(fngResp.Data) == 0 {
		return nil, fmt.Errorf("no FNG data available")
	}

	data := fngResp.Data[0]
	value := 0
	if _, err := fmt.Sscanf(data.Value, "%d", &value); err != nil {
		return nil, fmt.Errorf("failed to parse FNG value: %w", err)
	}

	multiplier := f.calculateMultiplier(value)

	return &types.FearGreedIndex{
		Value:          value,
		Classification: data.Classification,
		Timestamp:      time.Now(),
		Multiplier:     multiplier,
	}, nil
}

// calculateMultiplier calculates the investment multiplier based on F&G value
// Lower F&G values (fear) = higher multiplier (buy more)
// Higher F&G values (greed) = lower multiplier (buy less)
func (f *FNGService) calculateMultiplier(value int) decimal.Decimal {
	// F&G ranges from 0-100
	// 0-25: Extreme Fear (multiplier: 1.5-2.0)
	// 26-45: Fear (multiplier: 1.2-1.5)
	// 46-55: Neutral (multiplier: 1.0)
	// 56-75: Greed (multiplier: 0.7-1.0)
	// 76-100: Extreme Greed (multiplier: 0.5-0.7)

	switch {
	case value <= 20: // Extreme Fear
		// Linear interpolation: 0->2.0, 25->1.5
		ratio := decimal.NewFromInt(int64(25 - value)).Div(decimal.NewFromInt(25))
		return decimal.NewFromFloat(1.5).Add(ratio.Mul(decimal.NewFromFloat(0.5)))

	case value <= 40: // Fear
		// Linear interpolation: 25->1.5, 45->1.2
		ratio := decimal.NewFromInt(int64(45 - value)).Div(decimal.NewFromInt(20))
		return decimal.NewFromFloat(1.2).Add(ratio.Mul(decimal.NewFromFloat(0.3)))

	case value <= 60: // Neutral
		return decimal.NewFromFloat(1.0)

	case value <= 80: // Greed
		// Linear interpolation: 55->1.0, 75->0.7
		ratio := decimal.NewFromInt(int64(75 - value)).Div(decimal.NewFromInt(20))
		return decimal.NewFromFloat(0.7).Add(ratio.Mul(decimal.NewFromFloat(0.3)))

	default: // Extreme Greed (80-100)
		// Linear interpolation: 75->0.7, 100->0.5
		ratio := decimal.NewFromInt(int64(100 - value)).Div(decimal.NewFromInt(25))
		return decimal.NewFromFloat(0.5).Add(ratio.Mul(decimal.NewFromFloat(0.2)))
	}
}
