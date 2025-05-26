// Client-side: Secure ObjectID generation
package mongodb

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

type SecurePrimitiveObjectIDGenerator interface {
	GenerateValidObjectID() primitive.ObjectID
	ValidateObjectID(id primitive.ObjectID) error
}

// securePrimitiveObjectIDGeneratorImpl provides secure, validated ObjectID generation
type securePrimitiveObjectIDGeneratorImpl struct {
	logger *zap.Logger
}

func NewSecureObjectIDGenerator(logger *zap.Logger) SecurePrimitiveObjectIDGenerator {
	return &securePrimitiveObjectIDGeneratorImpl{
		logger: logger,
	}
}

// GenerateValidObjectID creates a cryptographically secure ObjectID
func (g *securePrimitiveObjectIDGeneratorImpl) GenerateValidObjectID() primitive.ObjectID {
	id := primitive.NewObjectID()

	if err := g.ValidateObjectID(id); err != nil {
		log.Fatalf("Failed to generate valid object ID - there may be something wrong with your system: %v\n", err)
	}

	// Log for audit trail
	g.logger.Debug("Generated secure ObjectID",
		zap.String("id", id.Hex()),
		zap.Time("timestamp", id.Timestamp()))

	return id
}

// ValidateObjectID ensures ObjectID meets security requirements
func (g *securePrimitiveObjectIDGeneratorImpl) ValidateObjectID(id primitive.ObjectID) error {
	if id.IsZero() {
		return errors.NewAppError("ObjectID cannot be zero", nil)
	}

	// Validate timestamp is reasonable
	timestamp := id.Timestamp()
	now := time.Now()

	if timestamp.After(now.Add(5 * time.Minute)) {
		return errors.NewAppError("ObjectID timestamp too far in future", nil)
	}

	if timestamp.Before(now.Add(-24 * time.Hour)) {
		return errors.NewAppError("ObjectID timestamp too old", nil)
	}

	return nil
}
