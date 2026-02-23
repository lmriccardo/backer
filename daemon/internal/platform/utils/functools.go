package utils

// Map applies the given function `fn` to each element of the input slice `in`,
// returning a new slice of results.
//
// It is a generic equivalent of the functional "map" operation found in many
// languages. The order of elements in the result matches the input.
//
// Type Parameters:
//   - T: the type of elements in the input slice
//   - U: the type of elements in the output slice
func Map[T, U any](fn func(T) U, in []T) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}
