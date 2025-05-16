// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/user/model.go
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	UserStatusActive   = 1   // User is active and can log in.
	UserStatusLocked   = 50  // User account is locked, typically due to too many failed login attempts.
	UserStatusArchived = 100 // User account is archived and cannot log in.

	UserRoleRoot       = 1 // Root user, has all permissions
	UserRoleCompany    = 2 // Company user, has permissions for company-related operations
	UserRoleIndividual = 3 // Individual user, has permissions for individual-related operations

	UserProfileVerificationStatusUnverified         = 1 // The user's profile has not yet been submitted for verification.
	UserProfileVerificationStatusSubmittedForReview = 2 // The user's profile has been submitted and is awaiting review.
	UserProfileVerificationStatusApproved           = 3 // The user's profile has been approved.
	UserProfileVerificationStatusRejected           = 4 // The user's profile has been rejected.

	// StorePendingStatus indicates this store needs to be reviewed by CPS and approved / rejected.
	StorePendingStatus  = 1 // Store is pending review.
	StoreActiveStatus   = 2 // Store is active and can be used.
	StoreRejectedStatus = 3 // Store has been rejected.
	StoreErrorStatus    = 4 // Store has encountered an error.
	StoreArchivedStatus = 5 // Store has been archived.

	EstimatedSubmissionsPerMonth1To10   = 1 // Estimated submissions per month: 1 to 10
	EstimatedSubmissionsPerMonth10To25  = 2 // Estimated submissions per month: 10 to 25
	EstimatedSubmissionsPerMonth25To50  = 3 // Estimated submissions per month: 25 to 50
	EstimatedSubmissionsPerMonth50To10  = 4 // Estimated submissions per month: 50 to 100
	EstimatedSubmissionsPerMonth100Plus = 5 // Estimated submissions per month: 100+

	HasOtherGradingServiceYes = 1 // Has other grading service: Yes
	HasOtherGradingServiceNo  = 2 // Has other grading service: No
	RequestWelcomePackageYes  = 1 // Request welcome package: Yes
	RequestWelcomePackageNo   = 2 // Request welcome package: No

	SpecialCollection040001 = 1
)

type User struct {
	ID                                             primitive.ObjectID `bson:"_id" json:"id"`
	Email                                          string             `bson:"email" json:"email"`
	FirstName                                      string             `bson:"first_name" json:"first_name"`
	LastName                                       string             `bson:"last_name" json:"last_name"`
	Name                                           string             `bson:"name" json:"name"`
	LexicalName                                    string             `bson:"lexical_name" json:"lexical_name"`
	PasswordHashAlgorithm                          string             `bson:"password_hash_algorithm" json:"-"`
	PasswordHash                                   string             `bson:"password_hash" json:"-"`
	Role                                           int8               `bson:"role" json:"role"`
	WasEmailVerified                               bool               `bson:"was_email_verified" json:"was_email_verified,omitempty"`
	EmailVerificationCode                          string             `bson:"email_verification_code,omitempty" json:"email_verification_code,omitempty"`
	EmailVerificationExpiry                        time.Time          `bson:"email_verification_expiry,omitempty" json:"email_verification_expiry,omitempty"`
	PasswordResetVerificationCode                  string             `bson:"password_reset_verification_code,omitempty" json:"password_reset_verification_code,omitempty"`
	PasswordResetVerificationExpiry                time.Time          `bson:"password_reset_verification_expiry,omitempty" json:"password_reset_verification_expiry,omitempty"`
	Phone                                          string             `bson:"phone" json:"phone,omitempty"`
	Country                                        string             `bson:"country" json:"country,omitempty"`
	Timezone                                       string             `bson:"timezone" json:"timezone"`
	Region                                         string             `bson:"region" json:"region,omitempty"`
	City                                           string             `bson:"city" json:"city,omitempty"`
	PostalCode                                     string             `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                                   string             `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                                   string             `bson:"address_line2" json:"address_line2,omitempty"`
	HasShippingAddress                             bool               `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                                   string             `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone                                  string             `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry                                string             `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion                                 string             `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                                   string             `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode                             string             `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1                           string             `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2                           string             `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	AgreeTermsOfService                            bool               `bson:"agree_terms_of_service" json:"agree_terms_of_service,omitempty"`
	AgreePromotions                                bool               `bson:"agree_promotions" json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool               `bson:"agree_to_tracking_across_third_party_apps_and_services" json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
	CreatedFromIPAddress                           string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	CreatedByUserID                                primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedAt                                      time.Time          `bson:"created_at" json:"created_at,omitempty"`
	CreatedByName                                  string             `bson:"created_by_name" json:"created_by_name"`
	ModifiedFromIPAddress                          string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	ModifiedByUserID                               primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedAt                                     time.Time          `bson:"modified_at" json:"modified_at,omitempty"`
	ModifiedByName                                 string             `bson:"modified_by_name" json:"modified_by_name"`
	Status                                         int8               `bson:"status" json:"status"`

	// The name of the payment processor we are using to handle payments with
	// this particular member.
	PaymentProcessorName string `bson:"payment_processor_name" json:"payment_processor_name"`
	// The unique identifier used by the payment processor which has a somesort of
	// copy of this member's details saved and we can reference that customer on
	// the payment processor using this `customer_id`.
	PaymentProcessorCustomerID string `bson:"payment_processor_customer_id" json:"payment_processor_customer_id"`

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

	// ProfileVerificationStatus indicates the profile verification status of this user account.
	ProfileVerificationStatus int8 `bson:"profile_verification_status" json:"profile_verification_status,omitempty"`

	HowDidYouHearAboutUs      int8   `bson:"how_did_you_hear_about_us" json:"how_did_you_hear_about_us,omitempty"`
	HowDidYouHearAboutUsOther string `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`

	StoreLogoS3Key         string    `bson:"store_logo_s3_key" json:"store_logo_s3_key,omitempty"`
	StoreLogoTitle         string    `bson:"store_logo_title" json:"store_logo_title,omitempty"`
	StoreLogoFileURL       string    `bson:"-" json:"store_logo_file_url,omitempty"` // (Optional, added by endpoint)
	StoreLogoFileURLExpiry time.Time `bson:"-" json:"store_logo_file_url_expiry"`    // (Optional, added by endpoint)

	ComicBookStoreName           string `bson:"comic_book_store_name" json:"comic_book_store_name,omitempty"`
	HowLongStoreOperating        int8   `bson:"how_long_store_operating" json:"how_long_store_operating,omitempty"`
	RetailPartnershipReason      string `bson:"retail_partnership_reason" json:"retail_partnership_reason,omitempty"`         // "Please describe how you could become a good retail partner for the ComicCoin Blockchain"
	ComicCoinPartnershipReason   string `bson:"comic_coin_partnership_reason" json:"comic_coin_partnership_reason,omitempty"` // "Please describe how the ComicCoin Blockchain could help you grow your business"
	EstimatedSubmissionsPerMonth int8   `bson:"estimated_submissions_per_month" json:"estimated_submissions_per_month"`
	HasOtherGradingService       int8   `bson:"has_other_grading_service" json:"has_other_grading_service"`
	OtherGradingServiceName      string `bson:"other_grading_service_name" json:"other_grading_service_name"`
	RequestWelcomePackage        int8   `bson:"request_welcome_package" json:"request_welcome_package"`

	HowLongCollectingComicBooksForGrading           int8 `bson:"how_long_collecting_comic_books_for_grading" json:"how_long_collecting_comic_books_for_grading"`
	HasPreviouslySubmittedComicBookForGrading       int8 `bson:"has_previously_submitted_comic_book_for_grading" json:"has_previously_submitted_comic_book_for_grading"`
	HasOwnedGradedComicBooks                        int8 `bson:"has_owned_graded_comic_books" json:"has_owned_graded_comic_books"`
	HasRegularComicBookShop                         int8 `bson:"has_regular_comic_book_shop" json:"has_regular_comic_book_shop"`
	HasPreviouslyPurchasedFromAuctionSite           int8 `bson:"has_previously_purchased_from_auction_site" json:"has_previously_purchased_from_auction_site"`
	HasPreviouslyPurchasedFromFacebookMarketplace   int8 `bson:"has_previously_purchased_from_facebook_marketplace" json:"has_previously_purchased_from_facebook_marketplace"`
	HasRegularlyAttendedComicConsOrCollectibleShows int8 `bson:"has_regularly_attended_comic_cons_or_collectible_shows" json:"has_regularly_attended_comic_cons_or_collectible_shows"`

	// Website URL of the individualuser's website/blog/etc or user's company website.
	WebsiteURL string `bson:"website_url" json:"website_url"`

	// Description of the individual user or user's company to be used in their profile.
	Description string `bson:"description" json:"description"`
}

type UserClaimedCoinTransaction struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Timestamp time.Time          `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
	Amount    uint64             `bson:"amount,omitempty" json:"amount,omitempty"`
}

// UserFilter represents the filter criteria for listing users
type UserFilter struct {
	// Basic filters
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Role   int8    `json:"role,omitempty"`
	Status int8    `json:"status,omitempty"`

	// Date range filters
	CreatedAtStart *time.Time `json:"created_at_start,omitempty"`
	CreatedAtEnd   *time.Time `json:"created_at_end,omitempty"`

	// Pagination - cursor based
	LastID        *primitive.ObjectID `json:"last_id,omitempty"`
	LastCreatedAt *time.Time          `json:"last_created_at,omitempty"`
	Limit         int64               `json:"limit,omitempty"`

	// Profile verification status filter
	ProfileVerificationStatus int8 `json:"profile_verification_status,omitempty"`

	// Search term for text search across multiple fields
	SearchTerm *string `json:"search_term,omitempty"`
}

// UserFilterResult represents the result of a filtered list operation
type UserFilterResult struct {
	Users         []*User            `json:"users"`
	HasMore       bool               `json:"has_more"`
	LastID        primitive.ObjectID `json:"last_id,omitempty"`
	LastCreatedAt time.Time          `json:"last_created_at,omitempty"`
	TotalCount    uint64             `json:"total_count,omitempty"`
}
