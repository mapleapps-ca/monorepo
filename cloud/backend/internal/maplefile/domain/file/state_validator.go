// cloud/backend/internal/maplefile/domain/file/state_validator.go
package file

import "errors"

// StateTransition validates file state transitions
type StateTransition struct {
	From string
	To   string
}

// IsValidStateTransition checks if a file state transition is allowed
func IsValidStateTransition(from, to string) error {
	validTransitions := map[StateTransition]bool{
		// From pending
		{FileStatePending, FileStateActive}:   true,
		{FileStatePending, FileStateDeleted}:  true,
		{FileStatePending, FileStateArchived}: false,

		// From active
		{FileStateActive, FileStatePending}:  false,
		{FileStateActive, FileStateDeleted}:  true,
		{FileStateActive, FileStateArchived}: true,

		// From deleted (cannot be restored nor archived)
		{FileStateDeleted, FileStatePending}:  false,
		{FileStateDeleted, FileStateActive}:   false,
		{FileStateDeleted, FileStateArchived}: false,

		// From archived (can only be restored to active)
		{FileStateArchived, FileStateActive}: true,

		// Same state transitions (no-op)
		{FileStatePending, FileStatePending}:   true,
		{FileStateActive, FileStateActive}:     true,
		{FileStateDeleted, FileStateDeleted}:   true,
		{FileStateArchived, FileStateArchived}: true,
	}

	if !validTransitions[StateTransition{from, to}] {
		return errors.New("invalid state transition from " + from + " to " + to)
	}

	return nil
}
