package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	incomepropertykit "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/calculators/incomepropertyevaluatorkit"
)

func main() {
	// Create a mortgage
	mortgage := &incomepropertykit.Mortgage{
		LoanPurchaseAmount:     decimal.NewFromFloat(250000.00),
		LoanAmount:             decimal.NewFromFloat(200000.00),
		DownPayment:            decimal.NewFromFloat(50000.00),
		AmortizationYears:      decimal.NewFromInt(25),
		AnnualInterestRate:     decimal.NewFromFloat(0.04), // 4%
		PaymentFrequency:       incomepropertykit.Monthly,
		CompoundingPeriod:      incomepropertykit.SemiAnnualCompounding,
		FirstPaymentDate:       time.Now(),
		Insurance:              "CMHC",
		MortgagePayment:        decimal.Zero,
		InterestRatePerPayment: decimal.Zero,
		TotalNumberOfPayments:  decimal.Zero,
		PercentFinanced:        decimal.Zero,
		InsuranceAmount:        decimal.Zero,
	}

	// Calculate mortgage details
	mortgageCalc := incomepropertykit.NewMortgageCalculator(mortgage)
	mortgage.MortgagePayment = mortgageCalc.CalculateMortgagePayment()
	mortgage.InterestRatePerPayment = mortgageCalc.InterestRatePerPaymentFrequency()
	mortgage.TotalNumberOfPayments = mortgageCalc.TotalNumberOfPayments()
	mortgage.PercentFinanced = mortgageCalc.PercentOfLoanFinanced()
	mortgage.InsuranceAmount = mortgageCalc.MortgageInsurancePremium()

	fmt.Printf("Mortgage Payment: $%s\n", mortgage.MortgagePayment.StringFixed(2))
	fmt.Printf("Interest Rate Per Payment: %s\n", mortgage.InterestRatePerPayment.StringFixed(4))
	fmt.Printf("Total Number of Payments: %s\n", mortgage.TotalNumberOfPayments.StringFixed(0))
	fmt.Printf("Percent Financed: %s%%\n", mortgage.PercentFinanced.StringFixed(2))
	fmt.Printf("Insurance Amount: $%s\n", mortgage.InsuranceAmount.StringFixed(2))

	// Create a financial analysis
	analysis := &incomepropertykit.FinancialAnalysis{
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

	// Calculate financial analysis
	financialCalc := incomepropertykit.NewFinancialAnalysisCalculator(analysis)

	// Calculate and update net income values
	analysis.AnnualNetIncome = financialCalc.AnnualNetIncomeWithoutMortgage()
	analysis.MonthlyNetIncome = financialCalc.MonthlyNetIncomeWithoutMortgage()
	analysis.AnnualCashFlow = financialCalc.AnnualNetIncomeWithMortgage()
	analysis.MonthlyCashFlow = financialCalc.MonthlyNetIncomeWithMortgage()
	analysis.CapRateWithMortgage = financialCalc.CapRateWithMortgageExpenseIncluded()
	analysis.CapRateWithoutMortgage = financialCalc.CapRateWithMortgageExpenseExcluded()
	analysis.InitialInvestmentAmount = financialCalc.TotalInitialInvestmentAmount()

	fmt.Printf("\nFinancial Analysis:\n")
	fmt.Printf("Annual Net Income (without mortgage): $%s\n", analysis.AnnualNetIncome.StringFixed(2))
	fmt.Printf("Monthly Net Income (without mortgage): $%s\n", analysis.MonthlyNetIncome.StringFixed(2))
	fmt.Printf("Annual Cash Flow (with mortgage): $%s\n", analysis.AnnualCashFlow.StringFixed(2))
	fmt.Printf("Monthly Cash Flow (with mortgage): $%s\n", analysis.MonthlyCashFlow.StringFixed(2))
	fmt.Printf("Cap Rate (with mortgage): %s%%\n", analysis.CapRateWithMortgage.StringFixed(2))
	fmt.Printf("Cap Rate (without mortgage): %s%%\n", analysis.CapRateWithoutMortgage.StringFixed(2))
	fmt.Printf("Initial Investment: $%s\n", analysis.InitialInvestmentAmount.StringFixed(2))

	// Generate annual projections
	projections := financialCalc.GenerateAnnualProjections()

	fmt.Printf("\nAnnual Projections:\n")
	fmt.Printf("Year 1:\n")
	fmt.Printf("  Sales Price: $%s\n", projections[0].SalesPrice.StringFixed(2))
	fmt.Printf("  Debt Remaining: $%s\n", projections[0].DebtRemaining.StringFixed(2))
	fmt.Printf("  Proceeds of Sale: $%s\n", projections[0].ProceedsOfSale.StringFixed(2))
	fmt.Printf("  Cash Flow: $%s\n", projections[0].CashFlow.StringFixed(2))
	fmt.Printf("  ROI: %s%%\n", projections[0].ReturnOnInvestmentPercent.StringFixed(2))

	fmt.Printf("\nYear 10:\n")
	fmt.Printf("  Sales Price: $%s\n", projections[9].SalesPrice.StringFixed(2))
	fmt.Printf("  Debt Remaining: $%s\n", projections[9].DebtRemaining.StringFixed(2))
	fmt.Printf("  Proceeds of Sale: $%s\n", projections[9].ProceedsOfSale.StringFixed(2))
	fmt.Printf("  Cash Flow: $%s\n", projections[9].CashFlow.StringFixed(2))
	fmt.Printf("  ROI: %s%%\n", projections[9].ReturnOnInvestmentPercent.StringFixed(2))

	fmt.Printf("\nYear 15:\n")
	fmt.Printf("  Sales Price: $%s\n", projections[14].SalesPrice.StringFixed(2))
	fmt.Printf("  Debt Remaining: $%s\n", projections[14].DebtRemaining.StringFixed(2))
	fmt.Printf("  Proceeds of Sale: $%s\n", projections[14].ProceedsOfSale.StringFixed(2))
	fmt.Printf("  Cash Flow: $%s\n", projections[14].CashFlow.StringFixed(2))
	fmt.Printf("  ROI: %s%%\n", projections[14].ReturnOnInvestmentPercent.StringFixed(2))

	fmt.Printf("\nYear 25:\n")
	fmt.Printf("  Sales Price: $%s\n", projections[24].SalesPrice.StringFixed(2))
	fmt.Printf("  Debt Remaining: $%s\n", projections[24].DebtRemaining.StringFixed(2))
	fmt.Printf("  Proceeds of Sale: $%s\n", projections[24].ProceedsOfSale.StringFixed(2))
	fmt.Printf("  Cash Flow: $%s\n", projections[24].CashFlow.StringFixed(2))
	fmt.Printf("  ROI: %s%%\n", projections[24].ReturnOnInvestmentPercent.StringFixed(2))

	// Calculate land transfer tax
	taxCalc := incomepropertykit.TaxCalculator{}
	landTransferTax := taxCalc.CalculateLandTransferTax(decimal.NewFromFloat(250000.00))
	fmt.Printf("\nLand Transfer Tax: $%s\n", landTransferTax.StringFixed(2))
}
