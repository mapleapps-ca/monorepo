package incomepropertyevaluatorkit

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Tolerance constants for different types of financial values
var (
	// For monthly payments, etc.
	MonthlyPaymentTolerance = decimal.NewFromFloat(5.0)

	// For annual cash flows
	AnnualCashFlowTolerance = decimal.NewFromFloat(50.0)

	// For proceeds of sale and larger monetary values
	ProceedsOfSaleTolerance = decimal.NewFromFloat(1000.0)

	// For loan balances
	BalanceTolerance = decimal.NewFromFloat(1500.0)

	// For appreciated values
	AppreciatedValueTolerance = decimal.NewFromFloat(100.0)

	// For percentages and rates
	RateTolerance = decimal.NewFromFloat(0.1)
)

// DecimalsAlmostEqual checks if two decimals are within tolerance of each other
func DecimalsAlmostEqual(t *testing.T, expected, actual decimal.Decimal, tolerance decimal.Decimal, msgAndArgs ...any) {
	t.Helper()
	diff := expected.Sub(actual).Abs()
	if diff.GreaterThan(tolerance) {
		assert.Fail(t, "Values not within tolerance",
			"Expected %s, got %s, difference %s exceeds tolerance %s",
			expected.String(), actual.String(), diff.String(), tolerance.String())
	}
}

// Helper functions for common comparison types
func MonthlyPaymentValuesAlmostEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...any) {
	DecimalsAlmostEqual(t, expected, actual, MonthlyPaymentTolerance, msgAndArgs...)
}

func AnnualCashFlowValuesAlmostEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...any) {
	DecimalsAlmostEqual(t, expected, actual, AnnualCashFlowTolerance, msgAndArgs...)
}

func ProceedsOfSaleValuesAlmostEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...any) {
	DecimalsAlmostEqual(t, expected, actual, ProceedsOfSaleTolerance, msgAndArgs...)
}

func BalanceValuesAlmostEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...any) {
	DecimalsAlmostEqual(t, expected, actual, BalanceTolerance, msgAndArgs...)
}

func AppreciatedValuesAlmostEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...any) {
	DecimalsAlmostEqual(t, expected, actual, AppreciatedValueTolerance, msgAndArgs...)
}

func RateValuesAlmostEqual(t *testing.T, expected, actual decimal.Decimal, msgAndArgs ...any) {
	DecimalsAlmostEqual(t, expected, actual, RateTolerance, msgAndArgs...)
}

// CreateMortgageForTests creates a mortgage with the exact same settings as in main.go
func CreateMortgageForTests() *Mortgage {
	// Using the exact same values and settings as in main.go
	return &Mortgage{
		LoanPurchaseAmount:     decimal.NewFromFloat(250000.00),
		LoanAmount:             decimal.NewFromFloat(200000.00),
		DownPayment:            decimal.NewFromFloat(50000.00),
		AmortizationYears:      decimal.NewFromInt(25),
		AnnualInterestRate:     decimal.NewFromFloat(0.04), // 4%
		PaymentFrequency:       Monthly,
		CompoundingPeriod:      SemiAnnualCompounding,
		FirstPaymentDate:       time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), // Fixed date to ensure reproducibility
		Insurance:              "CMHC",
		MortgagePayment:        decimal.Zero,
		InterestRatePerPayment: decimal.Zero,
		TotalNumberOfPayments:  decimal.Zero,
		PercentFinanced:        decimal.Zero,
		InsuranceAmount:        decimal.Zero,
	}
}

// CreateFinancialAnalysisForTests creates a financial analysis with the exact same settings as in main.go
func CreateFinancialAnalysisForTests() *FinancialAnalysis {
	mortgage := CreateMortgageForTests()

	return &FinancialAnalysis{
		PurchasePrice:             decimal.NewFromFloat(250000.00),
		InflationRate:             decimal.NewFromFloat(0.025), // 2.5%
		BuyingFeeRate:             decimal.NewFromFloat(0.006), // 0.6%
		SellingFeeRate:            decimal.NewFromFloat(0.06),  // 6%
		AnnualRentalIncome:        decimal.NewFromFloat(24600.00),
		MonthlyRentalIncome:       decimal.NewFromFloat(2050.00),
		AnnualFacilityIncome:      decimal.NewFromFloat(0.00),
		MonthlyFacilityIncome:     decimal.NewFromFloat(0.00),
		AnnualExpense:             decimal.NewFromFloat(7340.18),
		MonthlyExpense:            decimal.NewFromFloat(611.69),
		PurchaseFeesAmount:        decimal.NewFromFloat(58100.00),
		CapitalImprovementsAmount: decimal.NewFromFloat(0.00),
		AnnualNetIncome:           decimal.Zero,
		MonthlyNetIncome:          decimal.Zero,
		AnnualCashFlow:            decimal.Zero,
		MonthlyCashFlow:           decimal.Zero,
		CapRateWithMortgage:       decimal.Zero,
		CapRateWithoutMortgage:    decimal.Zero,
		InitialInvestmentAmount:   decimal.Zero,
		Mortgage:                  mortgage,
	}
}
