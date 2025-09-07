package nullable

// Nullable represents a value that may or may not be set.
type Nullable[T any] struct {
	value    T
	hasValue bool
}

// New creates a Nullable with a value.
func New[T any](v T) Nullable[T] {
	return Nullable[T]{value: v, hasValue: true}
}

// None creates an empty Nullable.
func None[T any]() Nullable[T] {
	return Nullable[T]{hasValue: false}
}

// HasValue returns true if the Nullable contains a value.
func (n Nullable[T]) HasValue() bool {
	return n.hasValue
}

// Get returns the value and whether it is set.
func (n Nullable[T]) Get() (T, bool) {
	return n.value, n.hasValue
}

// Set assigns a value and marks the Nullable as set.
func (n *Nullable[T]) Set(v T) {
	n.value = v
	n.hasValue = true
}

// Clear removes the value and marks the Nullable as unset.
func (n *Nullable[T]) Clear() {
	var zero T
	n.value = zero
	n.hasValue = false
}
