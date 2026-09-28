package compensate

import "context"

// When CompensateFunc returns nil error, any future calls will return
// ErrCompensationAlreadyUsed
type Func func(ctx context.Context) error
