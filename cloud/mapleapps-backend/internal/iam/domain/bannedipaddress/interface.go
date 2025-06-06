package bannedipaddress

import (
	"context"
	"math/big"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository Interface for a BannedIPAddress model in the database.
type Repository interface {
	Create(ctx context.Context, m *BannedIPAddress) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*BannedIPAddress, error)
	GetByNonce(ctx context.Context, nonce *big.Int) (*BannedIPAddress, error)
	UpdateByID(ctx context.Context, m *BannedIPAddress) error
	CountByFilter(ctx context.Context, filter *BannedIPAddressFilter) (uint64, error)
	ListByFilter(ctx context.Context, filter *BannedIPAddressFilter) (*BannedIPAddressFilterResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	ListAllValues(ctx context.Context) ([]string, error)
}
