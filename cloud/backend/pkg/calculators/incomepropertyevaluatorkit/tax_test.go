package incomepropertyevaluatorkit

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTaxCalculator_CalculateLandTransferTax(t *testing.T) {
	taxCalc := TaxCalculator{}

	// Test case 1: Purchase price from main.go ($250,000)
	purchasePrice1 := decimal.NewFromFloat(250000.00)
	expectedTax1 := decimal.NewFromFloat(2225.00) // From main.go output
	actualTax1 := taxCalc.CalculateLandTransferTax(purchasePrice1)

	assert.True(t, expectedTax1.Equal(actualTax1), "Land transfer tax for $250,000 should be $2,225.00")

	// Test case 2: Purchase price below $55,000
	purchasePrice2 := decimal.NewFromFloat(50000.00)
	expectedTax2 := decimal.NewFromFloat(250.00) // 50000 * 0.005
	actualTax2 := taxCalc.CalculateLandTransferTax(purchasePrice2)

	assert.True(t, expectedTax2.Equal(actualTax2), "Land transfer tax for $50,000 should be $250.00")

	// Test case 3: Purchase price between $55,000 and $250,000
	purchasePrice3 := decimal.NewFromFloat(100000.00)
	expectedTax3 := decimal.NewFromFloat(725.00) // 100000 * 0.01 - 275
	actualTax3 := taxCalc.CalculateLandTransferTax(purchasePrice3)

	assert.True(t, expectedTax3.Equal(actualTax3), "Land transfer tax for $100,000 should be $725.00")

	// Test case 4: Purchase price between $250,000 and $400,000
	purchasePrice4 := decimal.NewFromFloat(300000.00)
	expectedTax4 := decimal.NewFromFloat(2975.00) // 300000 * 0.015 - 1525
	actualTax4 := taxCalc.CalculateLandTransferTax(purchasePrice4)

	assert.True(t, expectedTax4.Equal(actualTax4), "Land transfer tax for $300,000 should be $2,975.00")

	// Test case 5: Purchase price above $400,000
	purchasePrice5 := decimal.NewFromFloat(500000.00)
	expectedTax5 := decimal.NewFromFloat(6475.00) // 500000 * 0.02 - 3525
	actualTax5 := taxCalc.CalculateLandTransferTax(purchasePrice5)

	assert.True(t, expectedTax5.Equal(actualTax5), "Land transfer tax for $500,000 should be $6,475.00")
}

func TestTaxCalculator_CalculateLandTransferTaxEdgeCases(t *testing.T) {
	taxCalc := TaxCalculator{}

	// Test edge case: $0 purchase price
	purchasePrice0 := decimal.Zero
	expectedTax0 := decimal.Zero // 0 * 0.005 = 0
	actualTax0 := taxCalc.CalculateLandTransferTax(purchasePrice0)

	assert.True(t, expectedTax0.Equal(actualTax0), "Land transfer tax for $0 should be $0.00")

	// Test edge case: Exactly $55,000 purchase price
	purchasePrice55k := decimal.NewFromFloat(55000.00)
	expectedTax55k := decimal.NewFromFloat(275.00) // 55000 * 0.01 - 275 = 275
	actualTax55k := taxCalc.CalculateLandTransferTax(purchasePrice55k)

	assert.True(t, expectedTax55k.Equal(actualTax55k), "Land transfer tax for $55,000 should be $275.00")

	// Test edge case: Exactly $250,000 purchase price
	purchasePrice250k := decimal.NewFromFloat(250000.00)
	expectedTax250k := decimal.NewFromFloat(2225.00) // 250000 * 0.015 - 1525 = 2225
	actualTax250k := taxCalc.CalculateLandTransferTax(purchasePrice250k)

	assert.True(t, expectedTax250k.Equal(actualTax250k), "Land transfer tax for $250,000 should be $2,225.00")

	// Test edge case: Exactly $400,000 purchase price
	purchasePrice400k := decimal.NewFromFloat(400000.00)
	expectedTax400k := decimal.NewFromFloat(4475.00) // 400000 * 0.02 - 3525 = 4475
	actualTax400k := taxCalc.CalculateLandTransferTax(purchasePrice400k)

	assert.True(t, expectedTax400k.Equal(actualTax400k), "Land transfer tax for $400,000 should be $4,475.00")
}
