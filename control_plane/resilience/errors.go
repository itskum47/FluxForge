package resilience

import (
	"fmt"
)

// ReconciliationError represents reconciliation failures
type ReconciliationError struct {
	Total   int
	Success int
	Skipped int
	Failed  int
}

func (e *ReconciliationError) Error() string {
	return fmt.Sprintf("reconciliation partial failure: %d succeeded, %d skipped, %d failed (total: %d)",
		e.Success, e.Skipped, e.Failed, e.Total)
}
