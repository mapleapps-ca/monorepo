package incomepropertyevaluatorkit

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFinancialAnalysisCalculator_GenerateAnnualProjections(t *testing.T) {
	// Setup the test financial analysis
	analysis := CreateFinancialAnalysisForTests()

	// Setup the mortgage calculator and calculate mortgage payment
	mortgageCalc := NewMortgageCalculator(analysis.Mortgage)
	analysis.Mortgage.MortgagePayment = mortgageCalc.CalculateMortgagePayment()

	// Create the financial calculator
	calculator := NewFinancialAnalysisCalculator(analysis)

	// Generate projections
	projections := calculator.GenerateAnnualProjections()

	// Verify we have 30 years of projections
	assert.Equal(t, 30, len(projections), "Should have 30 years of projections")

	// Test Year 1 projection values with tolerance
	year1 := projections[0]
	assert.Equal(t, 1, year1.Year, "First projection should be year 1")

	expected1SalesPrice := decimal.NewFromFloat(256250.00)
	AppreciatedValuesAlmostEqual(t, expected1SalesPrice, year1.SalesPrice,
		"Year 1 sales price should be close to 256250.00")

	expected1DebtRemaining := decimal.NewFromFloat(196203.59)
	BalanceValuesAlmostEqual(t, expected1DebtRemaining, year1.DebtRemaining,
		"Year 1 debt remaining should be close to 196203.59")

	expected1ProceedsOfSale := decimal.NewFromFloat(45046.41)
	ProceedsOfSaleValuesAlmostEqual(t, expected1ProceedsOfSale, year1.ProceedsOfSale,
		"Year 1 proceeds of sale should be close to 45046.41")

	// Test Year 10 projection values with tolerance
	year10 := projections[9]
	assert.Equal(t, 10, year10.Year, "Tenth projection should be year 10")

	expected10SalesPrice := decimal.NewFromFloat(320119.57)
	AppreciatedValuesAlmostEqual(t, expected10SalesPrice, year10.SalesPrice,
		"Year 10 sales price should be close to 320119.57")

	expected10DebtRemaining := decimal.NewFromFloat(141481.42)
	BalanceValuesAlmostEqual(t, expected10DebtRemaining, year10.DebtRemaining,
		"Year 10 debt remaining should be close to 141481.42")
}

func TestAppreciatedDecimalNumber(t *testing.T) {
	// Test a sample value with inflation over various years
	value := decimal.NewFromFloat(100.00)
	inflationRate := decimal.NewFromFloat(0.025) // 2.5%

	// Test 25 year appreciation
	expected25 := decimal.NewFromFloat(185.06) // 100 * (1.025^25)
	actual25 := appreciatedDecimalNumber(value, 25, inflationRate)

	AppreciatedValuesAlmostEqual(t, expected25, actual25,
		"25 year appreciation should be close to 185.06")
}
