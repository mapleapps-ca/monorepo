// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/service/federateduser/service.go
package me

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	uc_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/usecase/federateduser"
)

type MeResponseDTO struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Email       string             `bson:"email" json:"email"`
	FirstName   string             `bson:"first_name" json:"first_name"`
	LastName    string             `bson:"last_name" json:"last_name"`
	Name        string             `bson:"name" json:"name"`
	LexicalName string             `bson:"lexical_name" json:"lexical_name"`
	Role        int8               `bson:"role" json:"role"`
	// WasEmailVerified        bool               `bson:"was_email_verified" json:"was_email_verified,omitempty"`
	// EmailVerificationCode   string             `bson:"email_verification_code,omitempty" json:"email_verification_code,omitempty"`
	// EmailVerificationExpiry time.Time          `bson:"email_verification_expiry,omitempty" json:"email_verification_expiry,omitempty"`
	Phone        string `bson:"phone" json:"phone,omitempty"`
	Country      string `bson:"country" json:"country,omitempty"`
	Timezone     string `bson:"timezone" json:"timezone"`
	Region       string `bson:"region" json:"region,omitempty"`
	City         string `bson:"city" json:"city,omitempty"`
	PostalCode   string `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1 string `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2 string `bson:"address_line2" json:"address_line2,omitempty"`
	// HasShippingAddress                              bool               `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	// ShippingName                                    string             `bson:"shipping_name" json:"shipping_name,omitempty"`
	// ShippingPhone                                   string             `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	// ShippingCountry                                 string             `bson:"shipping_country" json:"shipping_country,omitempty"`
	// ShippingRegion                                  string             `bson:"shipping_region" json:"shipping_region,omitempty"`
	// ShippingCity                                    string             `bson:"shipping_city" json:"shipping_city,omitempty"`
	// ShippingPostalCode                              string             `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	// ShippingAddressLine1                            string             `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	// ShippingAddressLine2                            string             `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	// HowDidYouHearAboutUs                            int8               `bson:"how_did_you_hear_about_us" json:"how_did_you_hear_about_us,omitempty"`
	// HowDidYouHearAboutUsOther                       string             `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	// AgreeTermsOfService                            bool               `bson:"agree_terms_of_service" json:"agree_terms_of_service,omitempty"`
	AgreePromotions                                bool `bson:"agree_promotions" json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool `bson:"agree_to_tracking_across_third_party_apps_and_services" json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
	// CreatedFromIPAddress                            string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	// CreatedByFederatedIdentityID                    primitive.ObjectID `bson:"created_by_federatedidentity_id" json:"created_by_federatedidentity_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at,omitempty"`
	// CreatedByName                                   string             `bson:"created_by_name" json:"created_by_name"`
	// ModifiedFromIPAddress                           string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	// ModifiedByFederatedIdentityID                   primitive.ObjectID `bson:"modified_by_federatedidentity_id" json:"modified_by_federatedidentity_id"`
	// ModifiedAt                                      time.Time          `bson:"modified_at" json:"modified_at,omitempty"`
	// ModifiedByName                                  string             `bson:"modified_by_name" json:"modified_by_name"`
	Status int8 `bson:"status" json:"status"`
	// PaymentProcessorName                            string             `bson:"payment_processor_name" json:"payment_processor_name"`
	// PaymentProcessorCustomerID                      string             `bson:"payment_processor_customer_id" json:"payment_processor_customer_id"`
	// OTPEnabled                                      bool               `bson:"otp_enabled" json:"otp_enabled"`
	// OTPVerified                                     bool               `bson:"otp_verified" json:"otp_verified"`
	// OTPValidated                                    bool               `bson:"otp_validated" json:"otp_validated"`
	// OTPSecret                                       string             `bson:"otp_secret" json:"-"`
	// OTPAuthURL                                      string             `bson:"otp_auth_url" json:"-"`
	// OTPBackupCodeHash                               string             `bson:"otp_backup_code_hash" json:"-"`
	// OTPBackupCodeHashAlgorithm                      string             `bson:"otp_backup_code_hash_algorithm" json:"-"`
	// HowLongCollectingComicBooksForGrading           int8               `bson:"how_long_collecting_comic_books_for_grading" json:"how_long_collecting_comic_books_for_grading"`
	// HasPreviouslySubmittedComicBookForGrading       int8               `bson:"has_previously_submitted_comic_book_for_grading" json:"has_previously_submitted_comic_book_for_grading"`
	// HasOwnedGradedComicBooks                        int8               `bson:"has_owned_graded_comic_books" json:"has_owned_graded_comic_books"`
	// HasRegularComicBookShop                         int8               `bson:"has_regular_comic_book_shop" json:"has_regular_comic_book_shop"`
	// HasPreviouslyPurchasedFromAuctionSite           int8               `bson:"has_previously_purchased_from_auction_site" json:"has_previously_purchased_from_auction_site"`
	// HasPreviouslyPurchasedFromFacebookMarketplace   int8               `bson:"has_previously_purchased_from_facebook_marketplace" json:"has_previously_purchased_from_facebook_marketplace"`
	// HasRegularlyAttendedComicConsOrCollectibleShows int8               `bson:"has_regularly_attended_comic_cons_or_collectible_shows" json:"has_regularly_attended_comic_cons_or_collectible_shows"`
	ProfileVerificationStatus int8   `bson:"profile_verification_status" json:"profile_verification_status,omitempty"`
	WebsiteURL                string `bson:"website_url" json:"website_url"`
	Description               string `bson:"description" json:"description"`
	ComicBookStoreName        string `bson:"comic_book_store_name" json:"comic_book_store_name,omitempty"`
}

type GetMeService interface {
	Execute(sessCtx context.Context) (*MeResponseDTO, error)
}

type getMeServiceImpl struct {
	config             *config.Configuration
	logger             *zap.Logger
	userGetByIDUseCase uc_user.FederatedUserGetByIDUseCase
	userCreateUseCase  uc_user.FederatedUserCreateUseCase
	userUpdateUseCase  uc_user.FederatedUserUpdateUseCase
}

func NewGetMeService(
	config *config.Configuration,
	logger *zap.Logger,
	userGetByIDUseCase uc_user.FederatedUserGetByIDUseCase,
	userCreateUseCase uc_user.FederatedUserCreateUseCase,
	userUpdateUseCase uc_user.FederatedUserUpdateUseCase,
) GetMeService {
	return &getMeServiceImpl{
		config:             config,
		logger:             logger,
		userGetByIDUseCase: userGetByIDUseCase,
		userCreateUseCase:  userCreateUseCase,
		userUpdateUseCase:  userUpdateUseCase,
	}
}

func (svc *getMeServiceImpl) Execute(sessCtx context.Context) (*MeResponseDTO, error) {
	//
	// Get required from context.
	//

	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting local federateduser id",
			zap.Any("error", "Not found in context: user_id"))
		return nil, errors.New("federateduser id not found in context")
	}

	// Get the federateduser account (aka "Me") and if it doesn't exist then we will
	// create it immediately here and now.
	federateduser, err := svc.userGetByIDUseCase.Execute(sessCtx, userID)
	if err != nil {
		svc.logger.Error("Failed getting me", zap.Any("error", err))
		return nil, err
	}
	if federateduser == nil {
		err := fmt.Errorf("FederatedUser does not exist for federated user id: %v", userID.Hex())
		svc.logger.Error("Failed getting me", zap.Any("error", err))
		return nil, err
	}

	return &MeResponseDTO{
		ID:              federateduser.ID,
		Email:           federateduser.Email,
		FirstName:       federateduser.FirstName,
		LastName:        federateduser.LastName,
		Name:            federateduser.Name,
		LexicalName:     federateduser.LexicalName,
		Role:            federateduser.Role,
		Phone:           federateduser.Phone,
		Country:         federateduser.Country,
		Timezone:        federateduser.Timezone,
		Region:          federateduser.Region,
		City:            federateduser.City,
		PostalCode:      federateduser.PostalCode,
		AddressLine1:    federateduser.AddressLine1,
		AddressLine2:    federateduser.AddressLine2,
		AgreePromotions: federateduser.AgreePromotions,
		AgreeToTrackingAcrossThirdPartyAppsAndServices: federateduser.AgreeToTrackingAcrossThirdPartyAppsAndServices,
		CreatedAt: federateduser.CreatedAt,
		Status:    federateduser.Status,
	}, nil
}
