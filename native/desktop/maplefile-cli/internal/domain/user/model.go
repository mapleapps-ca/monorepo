// monorepo/native/desktop/maplefile-cli/internal/domain/user/model.go
package user

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

const (
	UserStatusActive   = 1   // User is active and can log in.
	UserStatusLocked   = 50  // User account is locked, typically due to too many failed login attempts.
	UserStatusArchived = 100 // User account is archived and cannot log in.

	UserRoleRoot       = 1 // Root user, has all permissions
	UserRoleCompany    = 2 // Company user, has permissions for company-related operations
	UserRoleIndividual = 3 // Individual user, has permissions for individual-related operations
)

type User struct {
	ID                                             gocql.UUID `bson:"_id" json:"id"`
	Email                                          string     `bson:"email" json:"email"`
	FirstName                                      string     `bson:"first_name" json:"first_name"`
	LastName                                       string     `bson:"last_name" json:"last_name"`
	Name                                           string     `bson:"name" json:"name"`
	LexicalName                                    string     `bson:"lexical_name" json:"lexical_name"`
	Role                                           int8       `bson:"role" json:"role"`
	Status                                         int8       `bson:"status" json:"status"`
	WasEmailVerified                               bool       `bson:"was_email_verified" json:"was_email_verified,omitempty"`
	EmailVerificationCode                          string     `bson:"email_verification_code,omitempty" json:"email_verification_code,omitempty"`
	EmailVerificationExpiry                        time.Time  `bson:"email_verification_expiry,omitempty" json:"email_verification_expiry"`
	PasswordResetVerificationCode                  string     `bson:"password_reset_verification_code,omitempty" json:"password_reset_verification_code,omitempty"`
	PasswordResetVerificationExpiry                time.Time  `bson:"password_reset_verification_expiry,omitempty" json:"password_reset_verification_expiry"`
	Phone                                          string     `bson:"phone" json:"phone,omitempty"`
	Country                                        string     `bson:"country" json:"country,omitempty"`
	Timezone                                       string     `bson:"timezone" json:"timezone"`
	Region                                         string     `bson:"region" json:"region,omitempty"`
	City                                           string     `bson:"city" json:"city,omitempty"`
	PostalCode                                     string     `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                                   string     `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                                   string     `bson:"address_line2" json:"address_line2,omitempty"`
	HasShippingAddress                             bool       `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                                   string     `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone                                  string     `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry                                string     `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion                                 string     `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                                   string     `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode                             string     `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1                           string     `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2                           string     `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	AgreeTermsOfService                            bool       `bson:"agree_terms_of_service" json:"agree_terms_of_service,omitempty"`
	AgreePromotions                                bool       `bson:"agree_promotions" json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool       `bson:"agree_to_tracking_across_third_party_apps_and_services" json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`

	// --- E2EE Related ---
	PasswordSalt []byte `json:"password_salt" bson:"password_salt"`
	// KDFParams stores the key derivation function parameters used to derive the user's password hash.
	KDFParams                         keys.KDFParams                         `json:"kdf_params" bson:"kdf_params"`
	EncryptedMasterKey                keys.EncryptedMasterKey                `json:"encrypted_master_key" bson:"encrypted_master_key"`
	PublicKey                         keys.PublicKey                         `json:"public_key" bson:"public_key"`
	EncryptedPrivateKey               keys.EncryptedPrivateKey               `json:"encrypted_private_key" bson:"encrypted_private_key"`
	EncryptedRecoveryKey              keys.EncryptedRecoveryKey              `json:"encrypted_recovery_key" bson:"encrypted_recovery_key"`
	MasterKeyEncryptedWithRecoveryKey keys.MasterKeyEncryptedWithRecoveryKey `json:"master_key_encrypted_with_recovery_key" bson:"master_key_encrypted_with_recovery_key"`
	EncryptedChallenge                []byte                                 `json:"encrypted_challenge,omitempty" bson:"encrypted_challenge,omitempty"`
	VerificationID                    string                                 `json:"verificationID"`

	// Track KDF upgrade status
	LastPasswordChange   time.Time `json:"last_password_change" bson:"last_password_change"`
	KDFParamsNeedUpgrade bool      `json:"kdf_params_need_upgrade" bson:"kdf_params_need_upgrade"`

	// Key rotation tracking fields
	CurrentKeyVersion int                     `json:"current_key_version" bson:"current_key_version"`
	LastKeyRotation   *time.Time              `json:"last_key_rotation,omitempty" bson:"last_key_rotation,omitempty"`
	KeyRotationPolicy *keys.KeyRotationPolicy `json:"key_rotation_policy,omitempty" bson:"key_rotation_policy,omitempty"`

	// --- Metadata ---
	CreatedFromIPAddress  string     `bson:"created_from_ip_address" json:"created_from_ip_address"`
	CreatedByUserID       gocql.UUID `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedAt             time.Time  `bson:"created_at" json:"created_at"`
	CreatedByName         string     `bson:"created_by_name" json:"created_by_name"`
	ModifiedFromIPAddress string     `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	ModifiedByUserID      gocql.UUID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedAt            time.Time  `bson:"modified_at" json:"modified_at"`
	ModifiedByName        string     `bson:"modified_by_name" json:"modified_by_name"`
	LastLoginAt           time.Time  `json:"last_login_at" bson:"last_login_at"`

	// OTPEnabled controls whether we force 2FA or not during login.
	OTPEnabled bool `bson:"otp_enabled" json:"otp_enabled"`

	// OTPVerified indicates user has successfully validated their opt token afer enabling 2FA thus turning it on.
	OTPVerified bool `bson:"otp_verified" json:"otp_verified"`

	// OTPValidated automatically gets set as `false` on successful login and then sets `true` once successfully validated by 2FA.
	OTPValidated bool `bson:"otp_validated" json:"otp_validated"`

	// OTPSecret the unique one-time password secret to be shared between our
	// backend and 2FA authenticator sort of apps that support `TOPT`.
	OTPSecret string `bson:"otp_secret" json:"-"`

	// OTPAuthURL is the URL used to share.
	OTPAuthURL string `bson:"otp_auth_url" json:"-"`

	// OTPBackupCodeHash is the one-time use backup code which resets the 2FA settings and allow the user to setup 2FA from scratch for the user.
	OTPBackupCodeHash string `bson:"otp_backup_code_hash" json:"-"`

	// OTPBackupCodeHashAlgorithm tracks the hashing algorithm used.
	OTPBackupCodeHashAlgorithm string `bson:"otp_backup_code_hash_algorithm" json:"-"`
}

// UserFilter represents the filter criteria for listing users
type UserFilter struct {
	// Basic filters
	Name   *string `bson:"name,omitempty" json:"name,omitempty"`
	Email  *string `bson:"email,omitempty" json:"email,omitempty"`
	Role   int8    `bson:"role,omitempty" json:"role,omitempty"`
	Status int8    `bson:"status,omitempty" json:"status,omitempty"`

	// Date range filters
	CreatedAtStart *time.Time `bson:"created_at_start,omitempty" json:"created_at_start,omitempty"`
	CreatedAtEnd   *time.Time `bson:"created_at_end,omitempty" json:"created_at_end,omitempty"`

	// Pagination - cursor based
	LastID        *gocql.UUID `bson:"last_id,omitempty" json:"last_id,omitempty"`
	LastCreatedAt *time.Time  `bson:"last_created_at,omitempty" json:"last_created_at,omitempty"`
	Limit         int64       `bson:"limit,omitempty" json:"limit,omitempty"`

	// Search term for text search across multiple fields
	SearchTerm *string `bson:"search_term,omitempty" json:"search_term,omitempty"`
}

// UserFilterResult represents the result of a filtered list operation
type UserFilterResult struct {
	Users         []*User    `bson:"users" json:"users"`
	HasMore       bool       `bson:"has_more" json:"has_more"`
	LastID        gocql.UUID `bson:"last_id,omitempty" json:"last_id,omitempty"`
	LastCreatedAt time.Time  `bson:"last_created_at" json:"last_created_at"`
	TotalCount    uint64     `bson:"total_count,omitempty" json:"total_count,omitempty"`
}
