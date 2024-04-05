package util

// #nosec

// TODO: Make this configurable later.

func ModifyFunc[S, T any](a []S, b []T, fn func(S, T) S) []S {
	var result []S
	for _, item := range a {
		for _, other := range b {
			result = append(result, fn(item, other))
		}
	}

	return result
}

func FilterFunc[S, T any](a []S, b []T, fn func(S, T) bool) []S {
	var result []S
	for _, i := range a {
		for _, j := range b {
			if fn(i, j) {
				result = append(result, i)
			}
		}
	}

	return result
}

func FilterMapFunc[S, T comparable](a map[S]T, fn func(S, T) bool) map[S]T {
	result := make(map[S]T)

	for k, v := range a {
		if fn(k, v) {
			result[k] = v
		}
	}

	return result
}
