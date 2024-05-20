package jsondiff

import "strconv"

// A Position represents the position of a Delta in an object or an array.
type Position interface {
	// String returns the position as a string
	String() (name string)

	Equal(another Position) bool

	// CompareTo returns a true if the Position is smaller than another Position.
	// This function is used to sort Positions by the sort package.
	CompareTo(another Position) bool
}

// A Name is a Postition with a string, which means the delta is in an object.
type Name string

func (n Name) String() (name string) {
	return string(n)
}
func (n Name) Equal(another Position) bool {
	return n == another.(Name)
}
func (n Name) CompareTo(another Position) bool {
	return n < another.(Name)
}

// A Index is a Position with an int value, which means the Delta is in an Array.
type Index int

func (i Index) String() string {
	return strconv.Itoa(int(i))
}
func (i Index) Equal(another Position) bool {
	return i == another.(Index)
}
func (i Index) CompareTo(another Position) bool {
	return i < another.(Index)
}
