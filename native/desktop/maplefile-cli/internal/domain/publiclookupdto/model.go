package publiclookupdto

import "go.mongodb.org/mongo-driver/bson/primitive"

type PublicLookupRequestDTO struct {
	Email string `json:"email"`
}

type PublicLookupResponseDTO struct {
	UserID            primitive.ObjectID `json:"user_id"`
	Email             string             `json:"email"`
	Name              string             `json:"name"`                 // Optional: for display
	PublicKeyInBase64 string             `json:"public_key_in_base64"` // Base64 encoded
	VerificationID    string             `json:"verification_id"`
}
