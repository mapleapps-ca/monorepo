package incomepropertyevaluatorkit

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFinancialAnalysisCalculator_TotalRentalIncomeAmount(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expectedMonthly := decimal.NewFromFloat(2050.00)
	actualMonthly := calculator.TotalMonthlyRentalIncomeAmount()
	assert.True(t, expectedMonthly.Equal(actualMonthly), "Monthly rental income should be 2050.00")

	expectedAnnual := decimal.NewFromFloat(24600.00)
	actualAnnual := calculator.TotalAnnualRentalIncomeAmount()
	assert.True(t, expectedAnnual.Equal(actualAnnual), "Annual rental income should be 24600.00")
}

func TestFinancialAnalysisCalculator_TotalGrossIncomeAmount(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expectedMonthly := decimal.NewFromFloat(2050.00) // 2050 + 0
	actualMonthly := calculator.TotalMonthlyGrossIncomeAmount()
	assert.True(t, expectedMonthly.Equal(actualMonthly), "Monthly gross income should be 2050.00")

	expectedAnnual := decimal.NewFromFloat(24600.00) // 24600 + 0
	actualAnnual := calculator.TotalAnnualGrossIncomeAmount()
	assert.True(t, expectedAnnual.Equal(actualAnnual), "Annual gross income should be 24600.00")
}

func TestFinancialAnalysisCalculator_TotalInitialInvestmentAmount(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expected := decimal.NewFromFloat(58100.00) // 58100 + 0
	actual := calculator.TotalInitialInvestmentAmount()
	assert.True(t, expected.Equal(actual), "Initial investment should be 58100.00")
}

func TestFinancialAnalysisCalculator_TotalExpensesAmount(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expectedMonthly := decimal.NewFromFloat(611.69)
	actualMonthly := calculator.TotalMonthlyExpensesAmount()
	assert.True(t, expectedMonthly.Equal(actualMonthly), "Monthly expenses should be 611.69")

	expectedAnnual := decimal.NewFromFloat(7340.18)
	actualAnnual := calculator.TotalAnnualExpensesAmount()
	assert.True(t, expectedAnnual.Equal(actualAnnual), "Annual expenses should be 7340.18")
}

func TestFinancialAnalysisCalculator_NetIncomeWithoutMortgage(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expectedMonthly := decimal.NewFromFloat(1438.31) // 2050 - 611.69
	actualMonthly := calculator.MonthlyNetIncomeWithoutMortgage()
	assert.True(t, expectedMonthly.Equal(actualMonthly), "Monthly net income without mortgage should be 1438.31")

	expectedAnnual := decimal.NewFromFloat(17259.82) // 24600 - 7340.18
	actualAnnual := calculator.AnnualNetIncomeWithoutMortgage()
	assert.True(t, expectedAnnual.Equal(actualAnnual), "Annual net income without mortgage should be 17259.82")
}

func TestFinancialAnalysisCalculator_NetIncomeWithMortgage(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	mortgageCalc := NewMortgageCalculator(analysis.Mortgage)
	analysis.Mortgage.MortgagePayment = mortgageCalc.CalculateMortgagePayment()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expectedMonthly := decimal.NewFromFloat(382.64)
	actualMonthly := calculator.MonthlyNetIncomeWithMortgage()
	MonthlyPaymentValuesAlmostEqual(t, expectedMonthly, actualMonthly,
		"Monthly net income with mortgage should be close to 382.64")

	expectedAnnual := decimal.NewFromFloat(4591.78)
	actualAnnual := calculator.AnnualNetIncomeWithMortgage()
	AnnualCashFlowValuesAlmostEqual(t, expectedAnnual, actualAnnual,
		"Annual net income with mortgage should be close to 4591.78")
}

func TestFinancialAnalysisCalculator_CapRate(t *testing.T) {
	analysis := CreateFinancialAnalysisForTests()
	mortgageCalc := NewMortgageCalculator(analysis.Mortgage)
	analysis.Mortgage.MortgagePayment = mortgageCalc.CalculateMortgagePayment()
	calculator := NewFinancialAnalysisCalculator(analysis)

	expectedWithMortgage := decimal.NewFromFloat(1.84)
	actualWithMortgage := calculator.CapRateWithMortgageExpenseIncluded()
	RateValuesAlmostEqual(t, expectedWithMortgage, actualWithMortgage,
		"Cap rate with mortgage should be close to 1.84%")

	// For cap rates, use a smaller tolerance
	rateTolerance := decimal.NewFromFloat(0.05) // Allow 0.05% difference
	diff := expectedWithMortgage.Sub(actualWithMortgage).Abs()
	assert.True(t, diff.LessThan(rateTolerance),
		"Cap rate with mortgage should be close to 1.84%%, got %s%%, diff %s%%",
		actualWithMortgage.String(), diff.String())

	expectedWithoutMortgage := decimal.NewFromFloat(6.90)
	actualWithoutMortgage := calculator.CapRateWithMortgageExpenseExcluded()
	assert.True(t, expectedWithoutMortgage.Equal(actualWithoutMortgage),
		"Cap rate without mortgage should be 6.90%%")
}
