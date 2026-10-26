package jsondiff_test

import (
	"strconv"
	"testing"

	"github.com/andreyvit/jsondiff"
)

var benchmarkResult jsondiff.Diff

func BenchmarkCompareArraysUnchanged(b *testing.B) {
	for _, size := range []int{0, 1, 64, 256, 1024} {
		b.Run("N="+strconv.Itoa(size), func(b *testing.B) {
			left, right := benchmarkArrayPair(size)
			if diff := jsondiff.CompareArrays(left, right); len(diff) != 0 {
				b.Fatalf("expected no deltas, got %d", len(diff))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkResult = jsondiff.CompareArrays(left, right)
			}
		})
	}
}

func BenchmarkCompareObjectsUnchangedArray(b *testing.B) {
	for _, size := range []int{0, 1, 64, 256, 1024} {
		b.Run("N="+strconv.Itoa(size), func(b *testing.B) {
			left, right := benchmarkArrayPair(size)
			before := map[string]any{"heading": "Before", "items": left}
			after := map[string]any{"heading": "After", "items": right}
			if diff := jsondiff.CompareObjects(before, after); len(diff) != 1 {
				b.Fatalf("expected one delta, got %d", len(diff))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkResult = jsondiff.CompareObjects(before, after)
			}
		})
	}
}

func BenchmarkCompareArraysLastChanged(b *testing.B) {
	for _, size := range []int{1, 64, 256, 1024} {
		b.Run("N="+strconv.Itoa(size), func(b *testing.B) {
			left, right := benchmarkArrayPair(size)
			right[size-1].(map[string]any)["title"] = "Changed"
			if diff := jsondiff.CompareArrays(left, right); len(diff) != 1 {
				b.Fatalf("expected one delta, got %d", len(diff))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkResult = jsondiff.CompareArrays(left, right)
			}
		})
	}
}

func benchmarkArrayPair(size int) (left, right []any) {
	left = make([]any, size)
	right = make([]any, size)
	for i := range left {
		left[i] = map[string]any{"id": i, "title": "Unchanged"}
		right[i] = map[string]any{"id": i, "title": "Unchanged"}
	}
	return left, right
}
