package cvars

type Option int

const (
	// OptNilIfEmpty makes the Ptr function return nil if the given value is empty
	OptNilIfEmpty Option = iota + 1
)

// Ptr returns a pointer to the given value
func Ptr[T comparable](v T, opts ...Option) *T {
	var ev T

	for _, opt := range opts {
		if opt == OptNilIfEmpty && ev == v {
			return nil
		}
	}

	return &v
}
