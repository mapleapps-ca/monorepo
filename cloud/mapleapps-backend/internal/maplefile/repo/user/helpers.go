package user

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"

	dom "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/user"
)

// Serialization helpers
func (r *userStorerImpl) serializeProfileData(data *dom.UserProfileData) (string, error) {
	if data == nil {
		return "", nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *userStorerImpl) serializeSecurityData(data *dom.UserSecurityData) (string, error) {
	if data == nil {
		return "", nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *userStorerImpl) serializeMetadata(data *dom.UserMetadata) (string, error) {
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
func (r *userStorerImpl) deserializeUserData(profileJSON, securityJSON, metadataJSON string, user *dom.User) error {
	// Deserialize profile data
	if profileJSON != "" {
		var profileData dom.UserProfileData
		if err := json.Unmarshal([]byte(profileJSON), &profileData); err != nil {
			return fmt.Errorf("failed to unmarshal profile data: %w", err)
		}
		user.ProfileData = &profileData
	}

	// Deserialize security data
	if securityJSON != "" {
		var securityData dom.UserSecurityData
		if err := json.Unmarshal([]byte(securityJSON), &securityData); err != nil {
			return fmt.Errorf("failed to unmarshal security data: %w", err)
		}
		user.SecurityData = &securityData
	}

	// Deserialize metadata
	if metadataJSON != "" {
		var metadata dom.UserMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		user.Metadata = &metadata
	}

	return nil
}

// Search helpers
func (r *userStorerImpl) generateSearchTerms(user *dom.User) []string {
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

func (r *userStorerImpl) calculateSearchBucket(term string) int {
	h := fnv.New32a()
	h.Write([]byte(term))
	return int(h.Sum32() % 100) // Distribute across 100 buckets
}
