package incomepropertyevaluatorkit

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMortgageCalculator_CalculateMortgagePayment(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)

	// The expected value from main.go
	expected := decimal.NewFromFloat(1055.67)
	actual := calculator.CalculateMortgagePayment()

	// Use the tolerance-based comparison for money values
	MonthlyPaymentValuesAlmostEqual(t, expected, actual, "Mortgage payment should be close to 1055.67")
}

func TestMortgageCalculator_TotalNumberOfPayments(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)

	// 25 years * 12 months
	expected := decimal.NewFromInt(300)
	actual := calculator.TotalNumberOfPayments()

	assert.True(t, expected.Equal(actual), "Total number of payments should be 300")
}

func TestMortgageCalculator_InterestRatePerPaymentFrequency(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)

	// Based on main.go output
	expected := decimal.NewFromFloat(0.0033)
	actual := calculator.InterestRatePerPaymentFrequency()

	// For rates, use a smaller tolerance
	RateValuesAlmostEqual(t, expected, actual, "Interest rate per payment should be close to 0.0033")
}

func TestMortgageCalculator_PercentOfLoanFinanced(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)

	// 200000 / 250000 * 100 = 80%
	expected := decimal.NewFromFloat(80.00)
	actual := calculator.PercentOfLoanFinanced()

	assert.True(t, expected.Equal(actual), "Percent financed should be 80.00%")
}

func TestMortgageCalculator_CalculateMortgageInsurance(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)

	expected := decimal.NewFromFloat(4375.00) // 250000 * 0.0175
	actual := calculator.CalculateMortgageInsurance()

	assert.True(t, expected.Equal(actual), "CMHC insurance should be 4375.00")
}

func TestMortgageCalculator_GeneratePaymentSchedule(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)

	schedule := calculator.GeneratePaymentSchedule()

	// Verify number of payments
	expectedPayments := 300 // 25 years * 12 months
	assert.Equal(t, expectedPayments, len(schedule), "Schedule should have 300 payments")

	// Verify first payment
	firstPayment := schedule[0]
	assert.Equal(t, 1, firstPayment.Year, "First payment should be in year 1")
	assert.Equal(t, 1, firstPayment.Interval, "First payment should be interval 1")

	// Use tolerance-based comparison for payment amount
	expectedPayment := decimal.NewFromFloat(1055.67)
	MonthlyPaymentValuesAlmostEqual(t, expectedPayment, firstPayment.PaymentAmount,
		"First payment amount should be close to 1055.67")

	// Verify year 1 debt remaining with tolerance
	year1End := schedule[11] // Index 11 is the 12th payment (end of year 1)
	expectedDebt1 := decimal.NewFromFloat(196203.59)
	BalanceValuesAlmostEqual(t, expectedDebt1, year1End.LoanBalance,
		"Year 1 debt remaining should be close to 196203.59")
}

func TestDebtRemainingAtEndOfYear(t *testing.T) {
	mortgage := CreateMortgageForTests()
	calculator := NewMortgageCalculator(mortgage)
	schedule := calculator.GeneratePaymentSchedule()

	// Test year 1 debt remaining
	expected1 := decimal.NewFromFloat(196203.59)
	actual1 := DebtRemainingAtEndOfYear(1, schedule, mortgage)
	BalanceValuesAlmostEqual(t, expected1, actual1,
		"Year 1 debt remaining should be close to 196203.59")

	// Test year 10 debt remaining
	expected10 := decimal.NewFromFloat(141481.42)
	actual10 := DebtRemainingAtEndOfYear(10, schedule, mortgage)
	BalanceValuesAlmostEqual(t, expected10, actual10,
		"Year 10 debt remaining should be close to 141481.42")
}

func TestCalculatePaymentDate(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Test Monthly
	monthlyDate := calculatePaymentDate(baseDate, Monthly, 1, 3) // Year 1, 3rd payment
	expectedMonthly := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedMonthly, monthlyDate, "Monthly payment date incorrect")

	// Test Quarterly
	quarterlyDate := calculatePaymentDate(baseDate, Quarterly, 2, 2) // Year 2, 2nd payment
	expectedQuarterly := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedQuarterly, quarterlyDate, "Quarterly payment date incorrect")
}
