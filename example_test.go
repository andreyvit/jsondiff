package jsondiff_test

import (
	"fmt"

	"github.com/andreyvit/jsondiff"
)

func Example() {
	before := map[string]any{
		"foo": 10,
		"bar": 20,
		"boz": 30,
	}
	after := map[string]any{
		"foo": 10,
		"bar": 42,
	}
	diff := jsondiff.CompareObjects(before, after)
	fmt.Println(diff.Format(before))
	// Output: {
	// -  "bar": 20,
	// +  "bar": 42,
	// -  "boz": 30,
	//    "foo": 10
	//  }
}

func Example_arrays() {
	before := []any{
		"foo",
		"bar",
		"boz",
	}
	after := []any{
		"foo",
		"boz",
	}
	diff := jsondiff.CompareArrays(before, after)
	fmt.Println(diff.Format(before))
	// Output: [
	//    "foo",
	// -  "bar",
	//    "boz"
	//  ]
}

func ExampleHideUnchangedProperties() {
	before := map[string]any{
		"foo": 10,
		"bar": 20,
		"boz": 30,
	}
	after := map[string]any{
		"foo": 10,
		"bar": 42,
	}
	diff := jsondiff.CompareObjects(before, after)
	fmt.Println(diff.Format(before, jsondiff.HideUnchangedProperties))
	// Output: {
	// -  "bar": 20,
	// +  "bar": 42,
	// -  "boz": 30,
	//  }
}
