// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser/constants.go
package federateduser

const (
	FederatedUserStatusActive   = 1   // User is active and can log in.
	FederatedUserStatusLocked   = 50  // User account is locked, typically due to too many failed login attempts.
	FederatedUserStatusArchived = 100 // User account is archived and cannot log in.
)

const (
	FederatedUserRoleRoot       = 1 // Root user, has all permissions
	FederatedUserRoleCompany    = 2 // Company user, has permissions for company-related operations
	FederatedUserRoleIndividual = 3 // Individual user, has permissions for individual-related operations
)

// FederatedUserCodeType defines the types of codes that can be generated for a federated user.
// These are typically used for actions like email verification or password resets.
const (
	// FederatedUserCodeTypeEmailVerification represents a code sent to verify a user's email address.
	FederatedUserCodeTypeEmailVerification = "email_verification"
	// FederatedUserCodeTypePasswordReset represents a code sent to allow a user to reset their password.
	FederatedUserCodeTypePasswordReset = "password_reset"
)

// FederatedUserPlan defines the available subscription plans for a federated user.
const (
	// FederatedUserPlanFree is the basic, free-tier plan.
	FederatedUserPlanFree = "free"
	// FederatedUserPlanPro is the professional-tier plan with more features and resources.
	FederatedUserPlanPro = "pro"
	// FederatedUserPlanBusiness is the business-tier plan for teams and organizations.
	FederatedUserPlanBusiness = "business"
)

// FederatedUserPlanStorageLimits maps each federated user plan to its corresponding storage limit in bytes.
var FederatedUserPlanStorageLimits = map[string]int64{
	FederatedUserPlanFree:     10 * 1024 * 1024 * 1024,   // 10 GB
	FederatedUserPlanPro:      100 * 1024 * 1024 * 1024,  // 100 GB
	FederatedUserPlanBusiness: 1024 * 1024 * 1024 * 1024, // 1 TB
}

func GetStorageLimitForFederatedUserPlan(plan string) int64 {
	limit, exists := FederatedUserPlanStorageLimits[plan]
	if !exists {
		return FederatedUserPlanStorageLimits[FederatedUserPlanFree] // fallback to free if unknown
	}
	return limit
}
