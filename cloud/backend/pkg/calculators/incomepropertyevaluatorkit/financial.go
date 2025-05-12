package incomepropertyevaluatorkit

import (
	"github.com/shopspring/decimal"
)

// FinancialAnalysisCalculator handles property financial calculations
type FinancialAnalysisCalculator struct {
	Analysis *FinancialAnalysis
}

// NewFinancialAnalysisCalculator creates a new financial analysis calculator
func NewFinancialAnalysisCalculator(analysis *FinancialAnalysis) *FinancialAnalysisCalculator {
	return &FinancialAnalysisCalculator{
		Analysis: analysis,
	}
}

// TotalMonthlyRentalIncomeAmount calculates the total monthly rental income
func (calc *FinancialAnalysisCalculator) TotalMonthlyRentalIncomeAmount() decimal.Decimal {
	return calc.Analysis.MonthlyRentalIncome
}

// TotalAnnualRentalIncomeAmount calculates the total annual rental income
func (calc *FinancialAnalysisCalculator) TotalAnnualRentalIncomeAmount() decimal.Decimal {
	return calc.Analysis.AnnualRentalIncome
}

// TotalMonthlyFacilityIncomeAmount calculates the total monthly facility income
func (calc *FinancialAnalysisCalculator) TotalMonthlyFacilityIncomeAmount() decimal.Decimal {
	return calc.Analysis.MonthlyFacilityIncome
}

// TotalAnnualFacilityIncomeAmount calculates the total annual facility income
func (calc *FinancialAnalysisCalculator) TotalAnnualFacilityIncomeAmount() decimal.Decimal {
	return calc.Analysis.AnnualFacilityIncome
}

// TotalMonthlyGrossIncomeAmount calculates the total monthly gross income
func (calc *FinancialAnalysisCalculator) TotalMonthlyGrossIncomeAmount() decimal.Decimal {
	return calc.Analysis.MonthlyRentalIncome.Add(calc.Analysis.MonthlyFacilityIncome)
}

// TotalAnnualGrossIncomeAmount calculates the total annual gross income
func (calc *FinancialAnalysisCalculator) TotalAnnualGrossIncomeAmount() decimal.Decimal {
	return calc.Analysis.AnnualRentalIncome.Add(calc.Analysis.AnnualFacilityIncome)
}

// TotalPurchaseFeesAmount calculates the total amount of purchase fees
func (calc *FinancialAnalysisCalculator) TotalPurchaseFeesAmount() decimal.Decimal {
	return calc.Analysis.PurchaseFeesAmount
}

// TotalCapitalImprovementsAmount calculates the total amount of capital improvements
func (calc *FinancialAnalysisCalculator) TotalCapitalImprovementsAmount() decimal.Decimal {
	return calc.Analysis.CapitalImprovementsAmount
}

// TotalInitialInvestmentAmount calculates the total initial investment amount
func (calc *FinancialAnalysisCalculator) TotalInitialInvestmentAmount() decimal.Decimal {
	return calc.Analysis.PurchaseFeesAmount.Add(calc.Analysis.CapitalImprovementsAmount)
}

// TotalMonthlyExpensesAmount calculates the total monthly expenses
func (calc *FinancialAnalysisCalculator) TotalMonthlyExpensesAmount() decimal.Decimal {
	return calc.Analysis.MonthlyExpense
}

// TotalAnnualExpensesAmount calculates the total annual expenses
func (calc *FinancialAnalysisCalculator) TotalAnnualExpensesAmount() decimal.Decimal {
	return calc.Analysis.AnnualExpense
}

// MonthlyNetIncomeWithoutMortgage calculates the monthly net income without mortgage
func (calc *FinancialAnalysisCalculator) MonthlyNetIncomeWithoutMortgage() decimal.Decimal {
	grossIncome := calc.TotalMonthlyGrossIncomeAmount()
	expenses := calc.TotalMonthlyExpensesAmount()
	return grossIncome.Sub(expenses)
}

// AnnualNetIncomeWithoutMortgage calculates the annual net income without mortgage
func (calc *FinancialAnalysisCalculator) AnnualNetIncomeWithoutMortgage() decimal.Decimal {
	grossIncome := calc.TotalAnnualGrossIncomeAmount()
	expenses := calc.TotalAnnualExpensesAmount()
	return grossIncome.Sub(expenses)
}

// MonthlyNetIncomeWithMortgage calculates the monthly net income with mortgage
func (calc *FinancialAnalysisCalculator) MonthlyNetIncomeWithMortgage() decimal.Decimal {
	netIncome := calc.MonthlyNetIncomeWithoutMortgage()
	monthlyMortgagePayment := calc.Analysis.Mortgage.MortgagePayment

	// If payment frequency is not monthly, convert to monthly
	if calc.Analysis.Mortgage.PaymentFrequency != Monthly {
		paymentFreq := decimal.NewFromInt(int64(calc.Analysis.Mortgage.PaymentFrequency))
		annualPayment := monthlyMortgagePayment.Mul(paymentFreq)
		twelve := decimal.NewFromInt(12)
		monthlyMortgagePayment = annualPayment.Div(twelve)
	}

	return netIncome.Sub(monthlyMortgagePayment)
}

// AnnualNetIncomeWithMortgage calculates the annual net income with mortgage
func (calc *FinancialAnalysisCalculator) AnnualNetIncomeWithMortgage() decimal.Decimal {
	netIncome := calc.AnnualNetIncomeWithoutMortgage()
	paymentFreq := decimal.NewFromInt(int64(calc.Analysis.Mortgage.PaymentFrequency))
	annualMortgagePayment := calc.Analysis.Mortgage.MortgagePayment.Mul(paymentFreq)
	return netIncome.Sub(annualMortgagePayment)
}

// CapRateWithMortgageExpenseIncluded calculates the capitalization rate with mortgage included
func (calc *FinancialAnalysisCalculator) CapRateWithMortgageExpenseIncluded() decimal.Decimal {
	purchasePrice := calc.Analysis.PurchasePrice

	// Prevent division by zero
	if purchasePrice.IsZero() {
		return DecimalZero
	}

	netIncome := calc.AnnualNetIncomeWithMortgage()
	capRate := netIncome.Div(purchasePrice).Mul(DecimalHundred)

	return capRate.Round(2)
}

// CapRateWithMortgageExpenseExcluded calculates the capitalization rate without mortgage
func (calc *FinancialAnalysisCalculator) CapRateWithMortgageExpenseExcluded() decimal.Decimal {
	purchasePrice := calc.Analysis.PurchasePrice

	// Prevent division by zero
	if purchasePrice.IsZero() {
		return decimal.Zero
	}

	netIncome := calc.AnnualNetIncomeWithoutMortgage()
	hundred := decimal.NewFromInt(100)
	capRate := netIncome.Div(purchasePrice).Mul(hundred)

	return capRate.Round(2)
}
