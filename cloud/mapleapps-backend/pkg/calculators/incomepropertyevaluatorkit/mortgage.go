package incomepropertyevaluatorkit

import (
	"time"

	"github.com/shopspring/decimal"
)

// MortgageCalculator handles mortgage-related calculations
type MortgageCalculator struct {
	Mortgage *Mortgage
}

// NewMortgageCalculator creates a new mortgage calculator
func NewMortgageCalculator(mortgage *Mortgage) *MortgageCalculator {
	return &MortgageCalculator{
		Mortgage: mortgage,
	}
}

// CalculateMortgagePayment calculates the mortgage payment per payment period
func (calc *MortgageCalculator) CalculateMortgagePayment() decimal.Decimal {
	r := calc.InterestRatePerPaymentFrequency()
	n := calc.TotalNumberOfPayments()
	p := calc.Mortgage.LoanAmount

	// If no payments or interest rate is zero, handle as edge case
	if n.IsZero() || r.IsZero() {
		return decimal.Zero
	}

	// Formula: P = (r * PV * (1 + r)^n) / ((1 + r)^n - 1)
	// Where:
	// P = payment
	// r = interest rate per period
	// PV = present value (loan amount)
	// n = total number of payments

	// Calculate (1 + r)^n
	one := decimal.NewFromInt(1)
	onePlusR := one.Add(r)
	onePlusRPowN := onePlusR.Pow(n)

	// Calculate top: r * PV * (1 + r)^n
	top := r.Mul(p).Mul(onePlusRPowN)

	// Calculate bottom: (1 + r)^n - 1
	bottom := onePlusRPowN.Sub(one)

	// Prevent division by zero
	if bottom.IsZero() {
		return decimal.Zero
	}

	// Calculate payment: top / bottom
	payment := top.Div(bottom)

	// Round to 2 decimal places
	return payment.Round(2)
}

// TotalNumberOfPayments calculates the total number of payments over the life of the mortgage
func (calc *MortgageCalculator) TotalNumberOfPayments() decimal.Decimal {
	paymentFreq := decimal.NewFromInt(int64(calc.Mortgage.PaymentFrequency))
	return calc.Mortgage.AmortizationYears.Mul(paymentFreq)
}

// InterestRatePerPaymentFrequency calculates the interest rate per payment period
func (calc *MortgageCalculator) InterestRatePerPaymentFrequency() decimal.Decimal {
	compoundingPeriod := decimal.NewFromInt(int64(calc.Mortgage.CompoundingPeriod))
	annualInterestRate := calc.Mortgage.AnnualInterestRate
	paymentFrequency := decimal.NewFromInt(int64(calc.Mortgage.PaymentFrequency))

	// y = compounding periods per payment period
	y := compoundingPeriod.Div(paymentFrequency)

	// Calculate interest rate per payment: (1 + r/m)^y - 1
	// where r is annual rate, m is compounding periods per year, y is above
	one := decimal.NewFromInt(1)
	ratePerCompoundingPeriod := annualInterestRate.Div(compoundingPeriod)
	onePlusRatePerCompound := one.Add(ratePerCompoundingPeriod)

	// Calculating (1 + r/m)^y
	interestFactor := onePlusRatePerCompound.Pow(y)

	// Return (1 + r/m)^y - 1
	return interestFactor.Sub(one)
}

// PercentOfLoanFinanced calculates the percentage of the purchase price that is financed
func (calc *MortgageCalculator) PercentOfLoanFinanced() decimal.Decimal {
	loanPurchaseAmount := calc.Mortgage.LoanPurchaseAmount

	// Prevent division by zero
	if loanPurchaseAmount.IsZero() {
		return decimal.Zero
	}

	// Calculate loan amount minus down payment
	loanAmount := calc.Mortgage.LoanAmount

	// Calculate percent financed: (loanAmount / loanPurchaseAmount) * 100
	hundred := decimal.NewFromInt(100)
	percentFinanced := loanAmount.Div(loanPurchaseAmount).Mul(hundred)

	return percentFinanced.Round(2)
}

// CalculateMortgageInsurance calculates mortgage insurance premium
func (calc *MortgageCalculator) CalculateMortgageInsurance() decimal.Decimal {
	percentFinanced := calc.PercentOfLoanFinanced()
	loanPurchaseAmount := calc.Mortgage.LoanPurchaseAmount

	// If zero percent financed, no insurance needed
	if percentFinanced.IsZero() {
		return DecimalZero
	}

	// CMHC insurance rates (Canadian Mortgage and Housing Corporation)
	var rate decimal.Decimal

	switch {
	case percentFinanced.GreaterThanOrEqual(LTVNinetyPercent):
		rate = CMHCRateOver90Percent
	case percentFinanced.GreaterThanOrEqual(LTVEightyFivePercent) && percentFinanced.LessThan(LTVNinetyPercent):
		rate = CMHCRateBetween85And90Percent
	case percentFinanced.GreaterThanOrEqual(LTVEightyPercent) && percentFinanced.LessThan(LTVEightyFivePercent):
		rate = CMHCRateBetween80And85Percent
	default: // < 80%
		rate = CMHCRateUnder80Percent
	}

	premium := loanPurchaseAmount.Mul(rate)
	return premium.Round(2)
}

// FHAPremium calculates FHA mortgage insurance premium (US)
func (calc *MortgageCalculator) FHAPremium() decimal.Decimal {
	return calc.Mortgage.LoanAmount.Mul(FHAMortgageInsuranceRate).Round(2)
}

// MortgageInsurancePremium returns the appropriate mortgage insurance premium
func (calc *MortgageCalculator) MortgageInsurancePremium() decimal.Decimal {
	switch calc.Mortgage.Insurance {
	case "CMHC":
		return calc.CalculateMortgageInsurance()
	case "FHA":
		return calc.FHAPremium()
	default:
		return decimal.Zero
	}
}

// GeneratePaymentSchedule generates the complete mortgage payment schedule
func (calc *MortgageCalculator) GeneratePaymentSchedule() []MortgageInterval {
	mortgagePayment := calc.CalculateMortgagePayment()
	interestRatePerPayment := calc.InterestRatePerPaymentFrequency()
	loanBalance := calc.Mortgage.LoanAmount
	totalPaidToInterest := decimal.Zero
	totalPaidToBank := decimal.Zero

	schedule := []MortgageInterval{}

	amortYears := int(calc.Mortgage.AmortizationYears.IntPart())

	// Create a payment for each period
	for year := 1; year <= amortYears; year++ {
		for payment := 1; payment <= calc.Mortgage.PaymentFrequency; payment++ {
			// Calculate interest for this payment
			interestAmount := loanBalance.Mul(interestRatePerPayment).Round(2)

			// Calculate principal for this payment
			principalAmount := mortgagePayment.Sub(interestAmount).Round(2)

			// Update loan balance
			loanBalance = loanBalance.Sub(principalAmount).Round(2)

			// Update running totals
			totalPaidToInterest = totalPaidToInterest.Add(interestAmount).Round(2)
			totalPaidToBank = totalPaidToBank.Add(mortgagePayment).Round(2)

			// Calculate payment date
			paymentDate := calculatePaymentDate(calc.Mortgage.FirstPaymentDate, calc.Mortgage.PaymentFrequency, year, payment)

			// Create the interval
			interval := MortgageInterval{
				Year:                year,
				Interval:            payment,
				PaymentAmount:       mortgagePayment,
				InterestAmount:      interestAmount,
				PrincipleAmount:     principalAmount,
				LoanBalance:         loanBalance,
				TotalPaidToInterest: totalPaidToInterest,
				TotalPaidToBank:     totalPaidToBank,
				PaymentDate:         paymentDate,
			}

			schedule = append(schedule, interval)
		}
	}

	return schedule
}

// DebtRemainingAtEndOfYear calculates the remaining debt at the end of a specific year
func DebtRemainingAtEndOfYear(year int, schedule []MortgageInterval, mortgage *Mortgage) decimal.Decimal {
	// Find the last payment of the specified year
	index := (year * mortgage.PaymentFrequency) - 1

	// Return 0 if beyond the schedule
	if index >= len(schedule) {
		return decimal.Zero
	}

	return schedule[index].LoanBalance
}

// Helper function to calculate payment date
func calculatePaymentDate(firstPaymentDate time.Time, frequency int, year, payment int) time.Time {
	yearInterval := year - 1
	paymentInterval := payment - 1

	date := firstPaymentDate

	// Add years
	date = date.AddDate(yearInterval, 0, 0)

	// Add the appropriate interval based on payment frequency
	switch frequency {
	case Annual:
		// No additional adjustment needed
	case SemiAnnual:
		date = date.AddDate(0, paymentInterval*6, 0)
	case Quarterly:
		date = date.AddDate(0, paymentInterval*3, 0)
	case BiMonthly:
		date = date.AddDate(0, paymentInterval*2, 0)
	case Monthly:
		date = date.AddDate(0, paymentInterval, 0)
	case BiWeekly:
		date = date.AddDate(0, 0, paymentInterval*14)
	case Weekly:
		date = date.AddDate(0, 0, paymentInterval*7)
	}

	return date
}
