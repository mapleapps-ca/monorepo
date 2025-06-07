// repo/federateduser/helpers.go
package federateduser

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"

	dom "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
)

// Serialization helpers
func (r *federatedUserRepository) serializeProfileData(data *dom.FederatedUserProfileData) (string, error) {
	if data == nil {
		return "", nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *federatedUserRepository) serializeSecurityData(data *dom.FederatedUserSecurityData) (string, error) {
	if data == nil {
		return "", nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *federatedUserRepository) serializeMetadata(data *dom.FederatedUserMetadata) (string, error) {
	if data == nil {
		return "", nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Deserialization helper
func (r *federatedUserRepository) deserializeUserData(profileJSON, securityJSON, metadataJSON string, user *dom.FederatedUser) error {
	// Deserialize profile data
	if profileJSON != "" {
		var profileData dom.FederatedUserProfileData
		if err := json.Unmarshal([]byte(profileJSON), &profileData); err != nil {
			return fmt.Errorf("failed to unmarshal profile data: %w", err)
		}
		user.ProfileData = &profileData
	}

	// Deserialize security data
	if securityJSON != "" {
		var securityData dom.FederatedUserSecurityData
		if err := json.Unmarshal([]byte(securityJSON), &securityData); err != nil {
			return fmt.Errorf("failed to unmarshal security data: %w", err)
		}
		user.SecurityData = &securityData
	}

	// Deserialize metadata
	if metadataJSON != "" {
		var metadata dom.FederatedUserMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		user.Metadata = &metadata
	}

	return nil
}

// Search helpers
func (r *federatedUserRepository) generateSearchTerms(user *dom.FederatedUser) []string {
	terms := make([]string, 0)

	// Add lowercase versions of searchable fields
	if user.Email != "" {
		terms = append(terms, strings.ToLower(user.Email))
		// Also add email prefix for partial matching
		parts := strings.Split(user.Email, "@")
		if len(parts) > 0 {
			terms = append(terms, strings.ToLower(parts[0]))
		}
	}

	if user.Name != "" {
		terms = append(terms, strings.ToLower(user.Name))
		// Add individual words from name
		words := strings.Fields(strings.ToLower(user.Name))
		terms = append(terms, words...)
	}

	if user.FirstName != "" {
		terms = append(terms, strings.ToLower(user.FirstName))
	}

	if user.LastName != "" {
		terms = append(terms, strings.ToLower(user.LastName))
	}

	return terms
}

func (r *federatedUserRepository) calculateSearchBucket(term string) int {
	h := fnv.New32a()
	h.Write([]byte(term))
	return int(h.Sum32() % 100) // Distribute across 100 buckets
}
