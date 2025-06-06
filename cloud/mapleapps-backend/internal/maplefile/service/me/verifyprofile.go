// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/me/verifyprofile.go
package me

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	domain "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/user"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type VerifyProfileRequestDTO struct {
	// Common fields
	Country                   string `json:"country,omitempty"`
	Region                    string `json:"region,omitempty"`
	City                      string `json:"city,omitempty"`
	PostalCode                string `json:"postal_code,omitempty"`
	AddressLine1              string `json:"address_line1,omitempty"`
	AddressLine2              string `json:"address_line2,omitempty"`
	HasShippingAddress        bool   `json:"has_shipping_address,omitempty"`
	ShippingName              string `json:"shipping_name,omitempty"`
	ShippingPhone             string `json:"shipping_phone,omitempty"`
	ShippingCountry           string `json:"shipping_country,omitempty"`
	ShippingRegion            string `json:"shipping_region,omitempty"`
	ShippingCity              string `json:"shipping_city,omitempty"`
	ShippingPostalCode        string `json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1      string `json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2      string `json:"shipping_address_line2,omitempty"`
	HowDidYouHearAboutUs      int8   `json:"how_did_you_hear_about_us,omitempty"`
	HowDidYouHearAboutUsOther string `json:"how_did_you_hear_about_us_other,omitempty"`
	WebsiteURL                string `json:"website_url,omitempty"`
	Description               string `bson:"description" json:"description"`

	// Customer specific fields
	HowLongCollectingComicBooksForGrading           int8 `json:"how_long_collecting_comic_books_for_grading,omitempty"`
	HasPreviouslySubmittedComicBookForGrading       int8 `json:"has_previously_submitted_comic_book_for_grading,omitempty"`
	HasOwnedGradedComicBooks                        int8 `json:"has_owned_graded_comic_books,omitempty"`
	HasRegularComicBookShop                         int8 `json:"has_regular_comic_book_shop,omitempty"`
	HasPreviouslyPurchasedFromAuctionSite           int8 `json:"has_previously_purchased_from_auction_site,omitempty"`
	HasPreviouslyPurchasedFromFacebookMarketplace   int8 `json:"has_previously_purchased_from_facebook_marketplace,omitempty"`
	HasRegularlyAttendedComicConsOrCollectibleShows int8 `json:"has_regularly_attended_comic_cons_or_collectible_shows,omitempty"`

	// Retailer specific fields
	ComicBookStoreName         string `json:"comic_book_store_name,omitempty"`
	StoreLogo                  string `json:"store_logo,omitempty"`
	HowLongStoreOperating      int8   `json:"how_long_store_operating,omitempty"`
	GradingComicsExperience    string `json:"grading_comics_experience,omitempty"`
	RetailPartnershipReason    string `json:"retail_partnership_reason,omitempty"`
	ComicCoinPartnershipReason string `json:"comic_coin_partnership_reason,omitempty"`

	EstimatedSubmissionsPerMonth int8   `json:"estimated_submissions_per_month,omitempty"`
	HasOtherGradingService       int8   `json:"has_other_grading_service,omitempty"`
	OtherGradingServiceName      string `json:"other_grading_service_name,omitempty"`
	RequestWelcomePackage        int8   `json:"request_welcome_package,omitempty"`

	// Explicitly specify user role if needed (overrides the user's current role)
	UserRole int8 `json:"user_role,omitempty"`
}

type VerifyProfileResponseDTO struct {
	Message  string `json:"message"`
	UserRole int8   `json:"user_role"`
	Status   int8   `json:"profile_verification_status"`
}

type VerifyProfileService interface {
	Execute(sessCtx context.Context, req *VerifyProfileRequestDTO) (*VerifyProfileResponseDTO, error)
}

type verifyProfileServiceImpl struct {
	config             *config.Configuration
	logger             *zap.Logger
	userGetByIDUseCase uc_user.UserGetByIDUseCase
	userUpdateUseCase  uc_user.UserUpdateUseCase
}

func NewVerifyProfileService(
	config *config.Configuration,
	logger *zap.Logger,
	userGetByIDUseCase uc_user.UserGetByIDUseCase,
	userUpdateUseCase uc_user.UserUpdateUseCase,
) VerifyProfileService {
	logger = logger.Named("VerifyProfileService")
	return &verifyProfileServiceImpl{
		config:             config,
		logger:             logger,
		userGetByIDUseCase: userGetByIDUseCase,
		userUpdateUseCase:  userUpdateUseCase,
	}
}

func (s *verifyProfileServiceImpl) Execute(
	sessCtx context.Context,
	req *VerifyProfileRequestDTO,
) (*VerifyProfileResponseDTO, error) {
	//
	// STEP 1: Get required from context.
	//
	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		s.logger.Error("Failed getting local user id",
			zap.Any("error", "Not found in context: user_id"))
		return nil, errors.New("user id not found in context")
	}

	//
	// STEP 2: Retrieve user from database
	//
	user, err := s.userGetByIDUseCase.Execute(sessCtx, userID)
	if err != nil {
		s.logger.Error("Failed retrieving user", zap.Any("error", err))
		return nil, err
	}
	if user == nil {
		s.logger.Error("User not found", zap.Any("userID", userID))
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "User not found")
	}

	// Check if we need to override the user role based on the request
	if req.UserRole != 0 && (req.UserRole == domain.UserRoleIndividual || req.UserRole == domain.UserRoleCompany) {
		s.logger.Info("Setting user role based on request",
			zap.Int("original_role", int(user.Role)),
			zap.Int("new_role", int(req.UserRole)))
		user.Role = req.UserRole
	}

	//
	// STEP 3: Validate request based on user role
	//
	e := make(map[string]string)

	// Validate common fields regardless of role
	s.validateCommonFields(req, e)

	// Role-specific validation
	if user.Role == domain.UserRoleIndividual {
		s.validateCustomerFields(req, e)
	} else if user.Role == domain.UserRoleCompany {
		s.validateRetailerFields(req, e)
	} else {
		s.logger.Warn("Unrecognized user role", zap.Int("role", int(user.Role)))
		e["user_role"] = "Invalid user role. Must be either customer or retailer."
	}

	// Return validation errors if any
	if len(e) != 0 {
		s.logger.Warn("Failed validation", zap.Any("errors", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 4: Update user profile based on role
	//

	// Update common fields
	s.updateCommonFields(user, req)

	// Update role-specific fields
	if user.Role == domain.UserRoleIndividual {
		s.updateCustomerFields(user, req)
	} else if user.Role == domain.UserRoleCompany {
		s.updateRetailerFields(user, req)
	} else {
		// Developers Note: No you cannot update a user with an unrecognized role nor the administrative role!
		s.logger.Error("Unrecognized user role", zap.Int("role", int(user.Role)))
		e["user_role"] = "Invalid user role. Must be either customer or retailer."
	}

	//
	// STEP 5: Update profile verification status and timestamps
	//
	user.ProfileVerificationStatus = domain.UserProfileVerificationStatusSubmittedForReview
	user.ModifiedAt = time.Now()
	user.ModifiedFromIPAddress, _ = sessCtx.Value(constants.SessionIPAddress).(string)

	//
	// STEP 6: Save updated user to database
	//
	if err := s.userUpdateUseCase.Execute(sessCtx, user); err != nil {
		s.logger.Error("Failed to update user", zap.Any("error", err))
		return nil, err
	}

	//
	// STEP 7: Generate appropriate response
	//
	var responseMessage string
	if user.Role == domain.UserRoleIndividual {
		responseMessage = "Your profile has been submitted for verification. You'll be notified once it's been reviewed."
	} else if user.Role == domain.UserRoleCompany {
		responseMessage = "Your retailer profile has been submitted for verification. Our team will review your application and contact you soon."
	} else {
		responseMessage = "Your profile has been submitted for verification."
	}

	return &VerifyProfileResponseDTO{
		Message:  responseMessage,
		UserRole: user.Role,
		Status:   user.ProfileVerificationStatus,
	}, nil
}

// validateCommonFields validates fields common to all user types
func (s *verifyProfileServiceImpl) validateCommonFields(req *VerifyProfileRequestDTO, e map[string]string) {
	if req.Country == "" {
		e["country"] = "Country is required"
	}
	if req.City == "" {
		e["city"] = "City is required"
	}
	if req.AddressLine1 == "" {
		e["address_line1"] = "Address is required"
	}
	if req.PostalCode == "" {
		e["postal_code"] = "Postal code is required"
	}
	if req.HowDidYouHearAboutUs == 0 {
		e["how_did_you_hear_about_us"] = "How did you hear about us is required"
	}
	if req.HowDidYouHearAboutUs == 7 && req.HowDidYouHearAboutUsOther == "" { // Assuming 7 is "Other"
		e["how_did_you_hear_about_us_other"] = "Please specify how you heard about us"
	}

	// Validate shipping address if it's enabled
	if req.HasShippingAddress {
		if req.ShippingName == "" {
			e["shipping_name"] = "Shipping name is required"
		}
		if req.ShippingPhone == "" {
			e["shipping_phone"] = "Shipping phone is required"
		}
		if req.ShippingCountry == "" {
			e["shipping_country"] = "Shipping country is required"
		}
		if req.ShippingCity == "" {
			e["shipping_city"] = "Shipping city is required"
		}
		if req.ShippingAddressLine1 == "" {
			e["shipping_address_line1"] = "Shipping address is required"
		}
		if req.ShippingPostalCode == "" {
			e["shipping_postal_code"] = "Shipping postal code is required"
		}
	}

	// More common fields...
	if req.WebsiteURL == "" {
		e["website_url"] = "Website URL is required"
	}
	if req.Description == "" {
		e["description"] = "Description is required"
	}
}

// validateCustomerFields validates fields specific to customers
func (s *verifyProfileServiceImpl) validateCustomerFields(req *VerifyProfileRequestDTO, e map[string]string) {
	if req.HowLongCollectingComicBooksForGrading == 0 {
		e["how_long_collecting_comic_books_for_grading"] = "How long you've been collecting comic books for grading is required"
	}
	if req.HasPreviouslySubmittedComicBookForGrading == 0 {
		e["has_previously_submitted_comic_book_for_grading"] = "Previous submission information is required"
	}
	if req.HasOwnedGradedComicBooks == 0 {
		e["has_owned_graded_comic_books"] = "Information about owning graded comic books is required"
	}
	if req.HasRegularComicBookShop == 0 {
		e["has_regular_comic_book_shop"] = "Regular comic book shop information is required"
	}
	if req.HasPreviouslyPurchasedFromAuctionSite == 0 {
		e["has_previously_purchased_from_auction_site"] = "Auction site purchase information is required"
	}
	if req.HasPreviouslyPurchasedFromFacebookMarketplace == 0 {
		e["has_previously_purchased_from_facebook_marketplace"] = "Facebook Marketplace purchase information is required"
	}
	if req.HasRegularlyAttendedComicConsOrCollectibleShows == 0 {
		e["has_regularly_attended_comic_cons_or_collectible_shows"] = "Comic convention attendance information is required"
	}
}

// validateRetailerFields validates fields specific to retailers
func (s *verifyProfileServiceImpl) validateRetailerFields(req *VerifyProfileRequestDTO, e map[string]string) {
	if req.ComicBookStoreName == "" {
		e["comic_book_store_name"] = "Store name is required"
	}
	if req.HowLongStoreOperating == 0 {
		e["how_long_store_operating"] = "Store operation duration is required"
	}
	if req.GradingComicsExperience == "" {
		e["grading_comics_experience"] = "Grading comics experience is required"
	}
	if req.RetailPartnershipReason == "" {
		e["retail_partnership_reason"] = "Retail partnership reason is required"
	}
	if req.ComicBookStoreName == "" {
		e["comic_book_store_name"] = "Comic book store name is required"
	}
	if req.EstimatedSubmissionsPerMonth == 0 {
		e["estimated_submissions_per_month"] = "Estimated submissions per month is required"
	}
	if req.HasOtherGradingService == 0 {
		e["has_other_grading_service"] = "Other grading service information is required"
	}
	if req.HasOtherGradingService == 1 && req.OtherGradingServiceName == "" {
		e["other_grading_service_name"] = "Please specify the grading service"
	}
	if req.RequestWelcomePackage == 0 {
		e["request_welcome_package"] = "Welcome package request information is required"
	}
}

// updateCommonFields updates common fields for all user types
func (s *verifyProfileServiceImpl) updateCommonFields(user *domain.User, req *VerifyProfileRequestDTO) {
	user.Country = req.Country
	user.Region = req.Region
	user.City = req.City
	user.PostalCode = req.PostalCode
	user.AddressLine1 = req.AddressLine1
	user.AddressLine2 = req.AddressLine2
	user.HasShippingAddress = req.HasShippingAddress
	user.ShippingName = req.ShippingName
	user.ShippingPhone = req.ShippingPhone
	user.ShippingCountry = req.ShippingCountry
	user.ShippingRegion = req.ShippingRegion
	user.ShippingCity = req.ShippingCity
	user.ShippingPostalCode = req.ShippingPostalCode
	user.ShippingAddressLine1 = req.ShippingAddressLine1
	user.ShippingAddressLine2 = req.ShippingAddressLine2
	user.HowDidYouHearAboutUs = req.HowDidYouHearAboutUs
	user.HowDidYouHearAboutUsOther = req.HowDidYouHearAboutUsOther
	user.WebsiteURL = req.WebsiteURL
	user.Description = req.Description
}

// updateCustomerFields updates fields specific to customers
func (s *verifyProfileServiceImpl) updateCustomerFields(user *domain.User, req *VerifyProfileRequestDTO) {
	user.HowLongCollectingComicBooksForGrading = req.HowLongCollectingComicBooksForGrading
	user.HasPreviouslySubmittedComicBookForGrading = req.HasPreviouslySubmittedComicBookForGrading
	user.HasOwnedGradedComicBooks = req.HasOwnedGradedComicBooks
	user.HasRegularComicBookShop = req.HasRegularComicBookShop
	user.HasPreviouslyPurchasedFromAuctionSite = req.HasPreviouslyPurchasedFromAuctionSite
	user.HasPreviouslyPurchasedFromFacebookMarketplace = req.HasPreviouslyPurchasedFromFacebookMarketplace
	user.HasRegularlyAttendedComicConsOrCollectibleShows = req.HasRegularlyAttendedComicConsOrCollectibleShows
}

// updateRetailerFields updates fields specific to retailers
func (s *verifyProfileServiceImpl) updateRetailerFields(user *domain.User, req *VerifyProfileRequestDTO) {
	user.ComicBookStoreName = req.ComicBookStoreName
	user.HowLongStoreOperating = req.HowLongStoreOperating
	user.RetailPartnershipReason = req.RetailPartnershipReason
	user.ComicCoinPartnershipReason = req.ComicCoinPartnershipReason
	user.EstimatedSubmissionsPerMonth = req.EstimatedSubmissionsPerMonth
	user.HasOtherGradingService = req.HasOtherGradingService
	user.OtherGradingServiceName = req.OtherGradingServiceName
	user.RequestWelcomePackage = req.RequestWelcomePackage
}
