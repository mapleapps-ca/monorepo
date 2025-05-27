// internal/domain/collection/validation.go
package collection

import (
	"fmt"
)

// ValidateState validates that the state is one of the allowed values
func ValidateState(state string) error {
	switch state {
	case CollectionStateActive, CollectionStateDeleted, CollectionStateArchived:
		return nil
	default:
		return fmt.Errorf("invalid collection state: %s (must be one of: %s, %s, %s)",
			state, CollectionStateActive, CollectionStateDeleted, CollectionStateArchived)
	}
}

// IsValidStateTransition validates if a state transition is allowed
func IsValidStateTransition(fromState, toState string) error {
	// Validate both states first
	if err := ValidateState(fromState); err != nil {
		return fmt.Errorf("invalid from state: %w", err)
	}
	if err := ValidateState(toState); err != nil {
		return fmt.Errorf("invalid to state: %w", err)
	}

	// Define allowed transitions
	allowedTransitions := map[string][]string{
		CollectionStateActive:   {CollectionStateDeleted, CollectionStateArchived},
		CollectionStateDeleted:  {CollectionStateActive, CollectionStateArchived},
		CollectionStateArchived: {CollectionStateActive},
	}

	allowed, exists := allowedTransitions[fromState]
	if !exists {
		return fmt.Errorf("no transitions defined for state: %s", fromState)
	}

	for _, allowedState := range allowed {
		if allowedState == toState {
			return nil
		}
	}

	return fmt.Errorf("transition from %s to %s is not allowed", fromState, toState)
}

// GetDefaultState returns the default state for new collections
func GetDefaultState() string {
	return CollectionStateActive
}
