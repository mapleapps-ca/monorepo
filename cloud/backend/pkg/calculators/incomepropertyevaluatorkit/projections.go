package incomepropertyevaluatorkit

import (
	"github.com/shopspring/decimal"
)

// GenerateAnnualProjections generates financial projections for each year
func (calc *FinancialAnalysisCalculator) GenerateAnnualProjections() []AnnualProjection {
	// Create a slice to hold all projections
	projections := []AnnualProjection{}

	// Get required values
	mortgage := calc.Analysis.Mortgage
	paymentSchedule := NewMortgageCalculator(mortgage).GeneratePaymentSchedule()
	inflationRate := calc.Analysis.InflationRate
	annualNetIncomeWithMortgage := calc.AnnualNetIncomeWithMortgage()
	annualNetIncomeWithoutMortgage := calc.AnnualNetIncomeWithoutMortgage()
	salesPrice := calc.Analysis.PurchasePrice
	sellingFeeRate := calc.Analysis.SellingFeeRate
	initialInvestment := calc.TotalInitialInvestmentAmount()

	// For IRR calculation
	negInitialInvestment := initialInvestment.Neg() // Initial investment is negative
	cashFlowArray := []decimal.Decimal{negInitialInvestment}

	// Previous year's cash flow
	var previousYearsCashFlow decimal.Decimal
	previousYearsCashFlow = decimal.Zero

	zero := decimal.Zero
	hundred := decimal.NewFromInt(100)

	// Generate projections for 30 years
	for year := 1; year <= 30; year++ {
		// Calculate remaining debt at end of year
		loanBalance := DebtRemainingAtEndOfYear(year, paymentSchedule, mortgage)

		// Handle case where loan is paid off
		if loanBalance.LessThan(zero) {
			loanBalance = zero
		}

		// Calculate cash flow
		var cashFlow, appreciatedCashFlow decimal.Decimal
		if loanBalance.GreaterThan(zero) {
			cashFlow = annualNetIncomeWithMortgage
			appreciatedCashFlow = appreciatedDecimalNumber(annualNetIncomeWithMortgage, year, inflationRate)
		} else {
			cashFlow = annualNetIncomeWithoutMortgage
			appreciatedCashFlow = appreciatedDecimalNumber(annualNetIncomeWithoutMortgage, year, inflationRate)
		}

		// Calculate appreciated sales price
		appreciatedSalesPrice := appreciatedDecimalNumber(salesPrice, year, inflationRate)

		// Calculate legal & selling fees
		fees := salesPrice.Mul(sellingFeeRate)
		appreciatedFees := appreciatedDecimalNumber(fees, year, inflationRate)

		// Calculate proceeds of sale
		proceedsOfSale := appreciatedSalesPrice.Sub(appreciatedFees).Sub(loanBalance)

		// Calculate total return
		totalReturn := proceedsOfSale.Add(appreciatedCashFlow)

		// Calculate ROI
		roiRate := returnOnInvestmentRate(initialInvestment, totalReturn)
		roiPercent := roiRate.Mul(hundred)

		// IRR Calculation - Step 1
		if previousYearsCashFlow.IsZero() {
			previousYearsCashFlow = cashFlow
		}

		// IRR Calculation - Step 2
		netProceedsFromSales := previousYearsCashFlow.Add(proceedsOfSale)

		// IRR Calculation - Step 3
		cashFlowArray = append(cashFlowArray, netProceedsFromSales)

		// IRR Calculation - Step 4
		irr := calculateIRR(cashFlowArray)
		irrPercent := irr.Mul(hundred)

		// Create annual projection
		projection := AnnualProjection{
			Year:                      year,
			SalesPrice:                appreciatedSalesPrice,
			DebtRemaining:             loanBalance,
			LegalFees:                 appreciatedFees,
			ProceedsOfSale:            proceedsOfSale,
			CashFlow:                  appreciatedCashFlow,
			InitialInvestment:         initialInvestment,
			TotalReturn:               totalReturn,
			ReturnOnInvestmentRate:    roiRate,
			ReturnOnInvestmentPercent: roiPercent,
			AnnualizedROIRate:         irr,
			AnnualizedROIPercent:      irrPercent,
		}

		projections = append(projections, projection)

		// Update for next iteration - Step 1 (continued)
		// Remove the total return and just use cash flow
		cashFlowArray = cashFlowArray[:len(cashFlowArray)-1]
		cashFlowArray = append(cashFlowArray, previousYearsCashFlow)

		previousYearsCashFlow = appreciatedCashFlow
	}

	return projections
}

// appreciatedDecimalNumber calculates the appreciated value of a number over a number of years
func appreciatedDecimalNumber(value decimal.Decimal, year int, inflationRate decimal.Decimal) decimal.Decimal {
	one := decimal.NewFromInt(1)

	// appreciationRate = 1 + inflationRate
	appreciationRate := one.Add(inflationRate)

	// Convert year to decimal for exponentiation
	yearDecimal := decimal.NewFromInt(int64(year))

	// appreciationFactor = (1 + inflationRate)^year
	appreciationFactor := appreciationRate.Pow(yearDecimal)

	// appreciatedValue = value * appreciationFactor
	appreciatedValue := value.Mul(appreciationFactor)

	return appreciatedValue.Round(2)
}

// returnOnInvestmentRate calculates the ROI rate
func returnOnInvestmentRate(initialInvestment, totalReturn decimal.Decimal) decimal.Decimal {
	// Prevent division by zero
	if initialInvestment.IsZero() {
		return decimal.Zero
	}

	// ROI = (totalReturn - initialInvestment) / initialInvestment
	roi := totalReturn.Sub(initialInvestment).Div(initialInvestment)

	return roi.Round(4) // Round to 4 decimal places
}

// calculateIRR calculates the Internal Rate of Return for a series of cash flows
func calculateIRR(cashFlows []decimal.Decimal) decimal.Decimal {
	// Simple implementation - for a more robust solution, use Newton-Raphson method

	// For very simple cases, approximate IRR
	if len(cashFlows) <= 2 {
		if len(cashFlows) == 2 && cashFlows[0].IsNegative() && cashFlows[1].IsPositive() {
			// Simple one period return
			return cashFlows[1].Div(cashFlows[0].Neg()).Sub(DecimalOne)
		}
		return DecimalZero
	}

	// For more complex cases, iterate to find IRR
	guess := IRRInitialGuess
	tolerance := IRRTolerance
	increment := IRRIncrement

	// Try to converge to NPV = 0
	for i := 0; i < IRRMaxIterations; i++ {
		npv := calculateNPV(cashFlows, guess)

		if npv.Abs().LessThan(tolerance) {
			return guess
		}

		// Adjust guess based on NPV
		if npv.GreaterThan(DecimalZero) {
			guess = guess.Add(increment)
		} else {
			guess = guess.Sub(increment)
		}

		// Prevent infinite loop
		if guess.LessThan(IRRNegativeLimit) || guess.GreaterThan(DecimalOne) {
			break
		}
	}

	return guess
}

// calculateNPV calculates Net Present Value for a series of cash flows with a given discount rate
func calculateNPV(cashFlows []decimal.Decimal, rate decimal.Decimal) decimal.Decimal {
	npv := decimal.Zero
	one := decimal.NewFromInt(1)

	for i, flow := range cashFlows {
		// Calculate discount factor: 1/(1+rate)^i
		iDecimal := decimal.NewFromInt(int64(i))
		discountFactor := one.Add(rate).Pow(iDecimal)

		// Add discounted cash flow to NPV
		npv = npv.Add(flow.Div(discountFactor))
	}

	return npv
}
