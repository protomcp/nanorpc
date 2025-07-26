package testutils

// GetField retrieves a typed value from a map[K]any
func GetField[K comparable, T any](m map[K]any, key K) (T, bool) {
	var zero T
	if v, ok := m[key]; ok {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	return zero, false
}
