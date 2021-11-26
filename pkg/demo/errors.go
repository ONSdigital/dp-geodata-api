package demo

type Sentinel string

const (
	ErrMissingParams  = Sentinel("rows or cols is missing")
	ErrTooManyMetrics = Sentinel("too many metrics")
	ErrInvalidTable   = Sentinel("invalid table")
)

func (e Sentinel) Error() string {
	return string(e)
}

func (e Sentinel) Is(err error) bool {
	sentinel, ok := err.(Sentinel)
	if !ok {
		return false
	}
	return sentinel == e
}
