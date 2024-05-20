package jsondiff

import "reflect"

func modifiedSimilarity(oldValue, newValue interface{}) float64 {
	similarity := 0.3 // at least, they are at the same position
	if reflect.TypeOf(oldValue) == reflect.TypeOf(newValue) {
		similarity += 0.3 // types are same

		switch oldValue.(type) {
		case string:
			similarity += 0.4 * stringSimilarity(oldValue.(string), newValue.(string))
		case float64:
			ratio := oldValue.(float64) / newValue.(float64)
			if ratio > 1 {
				ratio = 1 / ratio
			}
			similarity += 0.4 * ratio
		}
	}
	return similarity
}

func moveSimilarity(beforeIndex, afterIndex Index) float64 {
	similarity := 0.6 // as type and contents are same
	ratio := float64(beforeIndex) / float64(afterIndex)
	if ratio > 1 {
		ratio = 1 / ratio
	}
	similarity += 0.4 * ratio
	return similarity
}

func deltasSimilarity(deltas []Delta) float64 {
	var similarity float64
	for _, delta := range deltas {
		similarity += delta.Similarity()
	}
	return similarity / float64(len(deltas))
}

func stringSimilarity(left, right string) (similarity float64) {
	matchingLength := float64(lcsLength([]rune(left), []rune(right), func(a, b rune) bool { return a == b }))
	return (matchingLength / float64(len(left))) * (matchingLength / float64(len(right)))
}
