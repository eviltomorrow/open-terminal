package preset

import (
	"time"
)

func SetNum[T int | int64](value, defaultValue T) T {
	if value == 0 {
		return defaultValue
	}
	return value
}

func SetString[T string](value, defaultValue T) T {
	if value == "" {
		return defaultValue
	}
	return value
}

func SetBool[T bool](value, _ T) T {
	return value
}

func SetDuration[T time.Duration](value, defaultValue T) T {
	if value == 0 {
		return defaultValue
	}
	return value
}
