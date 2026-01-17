package schema

import "fmt"

func convertStingerSliceToStringSlice[T fmt.Stringer](items []T) []string {
	var result []string
	for _, item := range items {
		result = append(result, item.String())
	}
	return result
}
