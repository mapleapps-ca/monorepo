package incomepropertyevaluatorkit

import (
	"github.com/shopspring/decimal"
)

// Payment Frequency Constants (already defined in models.go)
// const (
//     Monthly    = 12
//     BiMonthly  = 6
//     ...
// )

// Compounding Period Constants (already defined in models.go)
// const (
//     MonthlyCompounding    = 12
//     ...
// )

// Common Decimal Values
var (
	DecimalZero    = decimal.Zero
	DecimalOne     = decimal.NewFromInt(1)
	DecimalHundred = decimal.NewFromInt(100)
)

// CMHC Insurance Rates
var (
	CMHCRateOver90Percent         = decimal.NewFromFloat(0.0275) // 2.75%
	CMHCRateBetween85And90Percent = decimal.NewFromFloat(0.0200) // 2.00%
	CMHCRateBetween80And85Percent = decimal.NewFromFloat(0.0175) // 1.75%
	CMHCRateUnder80Percent        = decimal.Zero                 // 0%
)

// FHA Insurance Rates
var (
	FHAMortgageInsuranceRate = decimal.NewFromFloat(0.0175) // 1.75%
)

// Land Transfer Tax Thresholds
var (
	LTTLowerThreshold  = decimal.NewFromFloat(55000.0)
	LTTMiddleThreshold = decimal.NewFromFloat(250000.0)
	LTTUpperThreshold  = decimal.NewFromFloat(400000.0)
)

// Land Transfer Tax Rates
var (
	LTTRateLowerTier   = decimal.NewFromFloat(0.005) // 0.5%
	LTTRateMiddleTier  = decimal.NewFromFloat(0.01)  // 1.0%
	LTTRateUpperTier   = decimal.NewFromFloat(0.015) // 1.5%
	LTTRateHighestTier = decimal.NewFromFloat(0.02)  // 2.0%
)

// Land Transfer Tax Adjustments
var (
	LTTAdjustmentMiddleTier  = decimal.NewFromFloat(275)
	LTTAdjustmentUpperTier   = decimal.NewFromFloat(1525)
	LTTAdjustmentHighestTier = decimal.NewFromFloat(3525)
)

// Loan-to-Value Thresholds
var (
	LTVNinetyPercent     = decimal.NewFromInt(90)
	LTVEightyFivePercent = decimal.NewFromInt(85)
	LTVEightyPercent     = decimal.NewFromInt(80)
)

// IRR Calculation Constants
var (
	IRRInitialGuess  = decimal.NewFromFloat(0.1)
	IRRTolerance     = decimal.NewFromFloat(0.0001)
	IRRIncrement     = decimal.NewFromFloat(0.01)
	IRRNegativeLimit = decimal.NewFromFloat(-0.99)
	IRRMaxIterations = 100
)
