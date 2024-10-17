package helper

func ToPointer[T any](value T) *T {
	return &value
}
