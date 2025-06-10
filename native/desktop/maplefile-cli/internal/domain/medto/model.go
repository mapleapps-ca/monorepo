// native/desktop/maplefile-cli/internal/domain/medto/model.go
package medto

import (
	"time"

	"github.com/gocql/gocql"
)

// MeResponseDTO represents the response from the cloud service when getting or updating user profile
type MeResponseDTO struct {
	ID                                             gocql.UUID `bson:"_id" json:"id"`
	Email                                          string     `bson:"email" json:"email"`
	FirstName                                      string     `bson:"first_name" json:"first_name"`
	LastName                                       string     `bson:"last_name" json:"last_name"`
	Name                                           string     `bson:"name" json:"name"`
	LexicalName                                    string     `bson:"lexical_name" json:"lexical_name"`
	Role                                           int8       `bson:"role" json:"role"`
	WasEmailVerified                               bool       `bson:"was_email_verified" json:"was_email_verified,omitempty"`
	Phone                                          string     `bson:"phone" json:"phone,omitempty"`
	Country                                        string     `bson:"country" json:"country,omitempty"`
	Timezone                                       string     `bson:"timezone" json:"timezone"`
	Region                                         string     `bson:"region" json:"region,omitempty"`
	City                                           string     `bson:"city" json:"city,omitempty"`
	PostalCode                                     string     `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                                   string     `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                                   string     `bson:"address_line2" json:"address_line2,omitempty"`
	AgreePromotions                                bool       `bson:"agree_promotions" json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool       `bson:"agree_to_tracking_across_third_party_apps_and_services" json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
	CreatedAt                                      time.Time  `bson:"created_at" json:"created_at,omitempty"`
	Status                                         int8       `bson:"status" json:"status"`
}

// UpdateMeRequestDTO represents the request payload for updating user profile
type UpdateMeRequestDTO struct {
	Email                                          string `bson:"email" json:"email"`
	FirstName                                      string `bson:"first_name" json:"first_name"`
	LastName                                       string `bson:"last_name" json:"last_name"`
	Phone                                          string `bson:"phone" json:"phone,omitempty"`
	Country                                        string `bson:"country" json:"country,omitempty"`
	Region                                         string `bson:"region" json:"region,omitempty"`
	Timezone                                       string `bson:"timezone" json:"timezone"`
	AgreePromotions                                bool   `bson:"agree_promotions" json:"agree_promotions,omitempty"`
	AgreeToTrackingAcrossThirdPartyAppsAndServices bool   `bson:"agree_to_tracking_across_third_party_apps_and_services" json:"agree_to_tracking_across_third_party_apps_and_services,omitempty"`
}
