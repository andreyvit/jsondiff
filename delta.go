package jsondiff

// A Delta represents an atomic difference between two JSON objects.
type Delta interface {
	// Similarity ranges from 0 (completely different) to 1 (completely equal).
	Similarity() float64

	PositionMatches(pos Position) bool
}

type Added struct {
	Position Position
	Value    any
}

func NewAdded(position Position, value any) *Added {
	return &Added{position, value}
}
func (*Added) Similarity() float64 {
	return 0
}
func (d *Added) PositionMatches(pos Position) bool {
	return d.Position == pos
}

type Deleted struct {
	Position Position
	Value    any
}

func NewDeleted(position Position, value any) *Deleted {
	return &Deleted{position, value}
}
func (Deleted) Similarity() float64 {
	return 0
}
func (d *Deleted) PositionMatches(pos Position) bool {
	return d.Position == pos
}

// A Modified represents a field whose value is changed.
type Modified struct {
	Position   Position
	OldValue   any
	NewValue   any
	similarity float64
}

func NewModified(position Position, oldValue, newValue any) *Modified {
	return &Modified{position, oldValue, newValue, modifiedSimilarity(oldValue, newValue)}
}
func (d *Modified) Similarity() float64 {
	return d.similarity
}
func (d *Modified) PositionMatches(pos Position) bool {
	return d.Position == pos
}

type Moved struct {
	OldPosition Position
	NewPosition Position
	Value       any
	similarity  float64
}

func NewMoved(oldPosition Position, newPosition Position, value any) *Moved {
	return &Moved{oldPosition, newPosition, value, moveSimilarity(oldPosition.(Index), newPosition.(Index))}
}
func (d *Moved) Similarity() float64 {
	return d.similarity
}
func (d *Moved) PositionMatches(pos Position) bool {
	return d.NewPosition == pos
}

type Object struct {
	Position   Position
	Deltas     []Delta
	similarity float64
}

func NewObject(position Position, deltas []Delta) *Object {
	return &Object{position, deltas, deltasSimilarity(deltas)}
}
func (d *Object) Similarity() (similarity float64) {
	return d.similarity
}
func (d *Object) PositionMatches(pos Position) bool {
	return d.Position == pos
}

type Array struct {
	Position   Position
	Deltas     []Delta
	similarity float64
}

func NewArray(position Position, deltas []Delta) *Array {
	return &Array{position, deltas, deltasSimilarity(deltas)}
}
func (d *Array) Similarity() float64 {
	return d.similarity
}
func (d *Array) PositionMatches(pos Position) bool {
	return d.Position == pos
}
