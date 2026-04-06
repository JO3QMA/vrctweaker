//go:build !windows

package sleepsuppress

type noopExecutionState struct{}

// NewExecutionState is a no-op outside Windows.
func NewExecutionState() ExecutionState {
	return noopExecutionState{}
}

func (noopExecutionState) SetSuppress(_ bool) error {
	return nil
}
