package helper

import "strings"

func SliceCombine[T1 comparable, T2 any](slice1 []T1, slice2 []T2) map[T1]T2 {
	result := make(map[T1]T2)
	for i := 0; i < len(slice1); i++ {
		result[slice1[i]] = slice2[i]
	}

	return result
}

func Keys[T1 comparable, T2 any](object map[T1]T2) []T1 {
	result := make([]T1, 0, len(object))
	for key := range object {
		result = append(result, key)
	}

	return result
}

func Values[T1 comparable, T2 any](object map[T1]T2) []T2 {
	result := make([]T2, 0, len(object))
	for _, value := range object {
		result = append(result, value)
	}

	return result
}

func EndsWithAny(value string, slice []string) bool {
	for _, item := range slice {
		if strings.HasSuffix(value, item) {
			return true
		}
	}
	return false
}
