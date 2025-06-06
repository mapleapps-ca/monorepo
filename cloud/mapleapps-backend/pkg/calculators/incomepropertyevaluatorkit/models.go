package incomepropertyevaluatorkit

import (
	"time"

	"github.com/shopspring/decimal"
)

// Constants for payment frequency
const (
	Monthly    = 12
	BiMonthly  = 6
	Quarterly  = 4
	SemiAnnual = 2
	Annual     = 1
	BiWeekly   = 26
	Weekly     = 52
)

// Constants for compounding period
const (
	MonthlyCompounding    = 12
	SemiAnnualCompounding = 2
	AnnualCompounding     = 1
)

// Mortgage represents a mortgage loan
type Mortgage struct {
	LoanPurchaseAmount     decimal.Decimal // Total property purchase price
	LoanAmount             decimal.Decimal // Amount of the loan
	DownPayment            decimal.Decimal // Down payment amount
	AmortizationYears      decimal.Decimal // Years to amortize the loan
	AnnualInterestRate     decimal.Decimal // Annual interest rate (as a decimal, e.g., 0.04 for 4%)
	PaymentFrequency       int             // How often payments are made
	CompoundingPeriod      int             // How often interest is compounded
	FirstPaymentDate       time.Time       // Date of first payment
	MortgagePayment        decimal.Decimal // Calculated mortgage payment per period
	InterestRatePerPayment decimal.Decimal // Interest rate per payment period
	TotalNumberOfPayments  decimal.Decimal // Total number of payments
	PercentFinanced        decimal.Decimal // Percentage of purchase price that is financed
	Insurance              string          // Type of mortgage insurance (e.g., "CMHC", "FHA")
	InsuranceAmount        decimal.Decimal // Amount of mortgage insurance
}

// MortgageInterval represents a period in the mortgage payment schedule
type MortgageInterval struct {
	Year                int             // Year number
	Interval            int             // Interval within the year
	PaymentAmount       decimal.Decimal // Payment amount
	InterestAmount      decimal.Decimal // Portion going to interest
	PrincipleAmount     decimal.Decimal // Portion going to principal
	LoanBalance         decimal.Decimal // Remaining loan balance
	TotalPaidToInterest decimal.Decimal // Cumulative interest paid
	TotalPaidToBank     decimal.Decimal // Cumulative total paid
	PaymentDate         time.Time       // Date of this payment
}

// FinancialAnalysis holds financial data for property analysis
type FinancialAnalysis struct {
	PurchasePrice             decimal.Decimal // Purchase price of the property
	InflationRate             decimal.Decimal // Annual inflation rate as a decimal (e.g., 0.025 for 2.5%)
	BuyingFeeRate             decimal.Decimal // Rate for buying fees as a decimal
	SellingFeeRate            decimal.Decimal // Rate for selling fees as a decimal
	AnnualRentalIncome        decimal.Decimal // Annual rental income
	MonthlyRentalIncome       decimal.Decimal // Monthly rental income
	AnnualFacilityIncome      decimal.Decimal // Annual income from facilities
	MonthlyFacilityIncome     decimal.Decimal // Monthly income from facilities
	AnnualGrossIncome         decimal.Decimal // Total annual gross income
	MonthlyGrossIncome        decimal.Decimal // Total monthly gross income
	AnnualExpense             decimal.Decimal // Annual expenses
	MonthlyExpense            decimal.Decimal // Monthly expenses
	AnnualNetIncome           decimal.Decimal // Annual net income without mortgage
	MonthlyNetIncome          decimal.Decimal // Monthly net income without mortgage
	AnnualCashFlow            decimal.Decimal // Annual cash flow with mortgage
	MonthlyCashFlow           decimal.Decimal // Monthly cash flow with mortgage
	CapRateWithMortgage       decimal.Decimal // Cap rate with mortgage included
	CapRateWithoutMortgage    decimal.Decimal // Cap rate without mortgage
	PurchaseFeesAmount        decimal.Decimal // Amount of purchase fees
	CapitalImprovementsAmount decimal.Decimal // Amount spent on capital improvements
	InitialInvestmentAmount   decimal.Decimal // Total initial investment
	Mortgage                  *Mortgage       // Associated mortgage
}

// AnnualProjection represents financial projections for a specific year
type AnnualProjection struct {
	Year                      int             // Year number
	SalesPrice                decimal.Decimal // Projected sales price
	DebtRemaining             decimal.Decimal // Remaining debt
	LegalFees                 decimal.Decimal // Legal fees for selling
	ProceedsOfSale            decimal.Decimal // Net proceeds from sale
	CashFlow                  decimal.Decimal // Annual cash flow
	InitialInvestment         decimal.Decimal // Initial investment amount
	TotalReturn               decimal.Decimal // Total return
	ReturnOnInvestmentRate    decimal.Decimal // ROI as a rate
	ReturnOnInvestmentPercent decimal.Decimal // ROI as a percentage
	AnnualizedROIRate         decimal.Decimal // Annualized ROI as a rate
	AnnualizedROIPercent      decimal.Decimal // Annualized ROI as a percentage
}

// RentalIncome represents rental income for a property
type RentalIncome struct {
	AnnualAmount         decimal.Decimal // Total annual amount
	AnnualAmountPerUnit  decimal.Decimal // Annual amount per unit
	MonthlyAmount        decimal.Decimal // Total monthly amount
	MonthlyAmountPerUnit decimal.Decimal // Monthly amount per unit
	NumberOfUnits        decimal.Decimal // Number of rental units
	Name                 string          // Name/description
}

// Expense represents an expense for a property
type Expense struct {
	AnnualAmount  decimal.Decimal // Annual expense amount
	MonthlyAmount decimal.Decimal // Monthly expense amount
	Name          string          // Name/description
}

// PurchaseFee represents a fee associated with purchasing a property
type PurchaseFee struct {
	Amount decimal.Decimal // Fee amount
	Name   string          // Name/description
}
