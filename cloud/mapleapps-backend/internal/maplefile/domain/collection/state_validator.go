// monorepo/cloud/backend/internal/maplefile/domain/collection/state_validator.go
package collection

import "errors"

// StateTransition validates collection state transitions
type StateTransition struct {
	From string
	To   string
}

// IsValidStateTransition checks if a state transition is allowed
func IsValidStateTransition(from, to string) error {
	validTransitions := map[StateTransition]bool{
		// From active
		{CollectionStateActive, CollectionStateDeleted}:  true,
		{CollectionStateActive, CollectionStateArchived}: true,

		// From deleted (cannot be restored nor archived)
		{CollectionStateDeleted, CollectionStateActive}:   false,
		{CollectionStateDeleted, CollectionStateArchived}: false,

		// From archived (can only be restored to active)
		{CollectionStateArchived, CollectionStateActive}: true,

		// Same state transitions (no-op)
		{CollectionStateActive, CollectionStateActive}:     true,
		{CollectionStateDeleted, CollectionStateDeleted}:   true,
		{CollectionStateArchived, CollectionStateArchived}: true,
	}

	if !validTransitions[StateTransition{from, to}] {
		return errors.New("invalid state transition from " + from + " to " + to)
	}

	return nil
}
