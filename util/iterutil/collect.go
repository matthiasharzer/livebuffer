package iterutil

import "iter"

func Collect[T any](iterator iter.Seq[T]) []T {
	var result []T
	for value := range iterator {
		result = append(result, value)
	}
	return result
}

func Collect2[T any](iterator iter.Seq2[T, error]) ([]T, error) {
	var result []T
	for value, err := range iterator {
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}
