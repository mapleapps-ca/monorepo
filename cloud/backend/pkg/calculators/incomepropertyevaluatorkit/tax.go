package incomepropertyevaluatorkit

import (
	"github.com/shopspring/decimal"
)

// TaxCalculator provides tax-related calculations
type TaxCalculator struct{}

// CalculateLandTransferTax calculates the land transfer tax based on purchase price
func (t *TaxCalculator) CalculateLandTransferTax(purchasePrice decimal.Decimal) decimal.Decimal {
	var landTransferTax decimal.Decimal

	switch {
	case purchasePrice.LessThan(LTTLowerThreshold):
		landTransferTax = purchasePrice.Mul(LTTRateLowerTier)
	case purchasePrice.GreaterThanOrEqual(LTTLowerThreshold) && purchasePrice.LessThan(LTTMiddleThreshold):
		landTransferTax = purchasePrice.Mul(LTTRateMiddleTier).Sub(LTTAdjustmentMiddleTier)
	case purchasePrice.GreaterThanOrEqual(LTTMiddleThreshold) && purchasePrice.LessThan(LTTUpperThreshold):
		landTransferTax = purchasePrice.Mul(LTTRateUpperTier).Sub(LTTAdjustmentUpperTier)
	default: // >= LTTUpperThreshold
		landTransferTax = purchasePrice.Mul(LTTRateHighestTier).Sub(LTTAdjustmentHighestTier)
	}

	return landTransferTax.Round(2)
}
