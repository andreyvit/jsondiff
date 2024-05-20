package jsondiff

import (
	"container/list"
	"reflect"
)

type Diff []Delta

func CompareObjects(left, right map[string]any) Diff {
	deltas := make([]Delta, 0)

	names := sortedKeys(left) // stabilize delta order
	for _, name := range names {
		if rightValue, ok := right[name]; ok {
			same, delta := compareValues(Name(name), left[name], rightValue)
			if !same {
				deltas = append(deltas, delta)
			}
		} else {
			deltas = append(deltas, NewDeleted(Name(name), left[name]))
		}
	}

	names = sortedKeys(right) // stabilize delta order
	for _, name := range names {
		if _, ok := left[name]; !ok {
			deltas = append(deltas, NewAdded(Name(name), right[name]))
		}
	}

	return deltas
}

type maybe struct {
	index    int
	lcsIndex int
	item     any
}

func compareArrays(left, right []any) []Delta {
	deltas := make([]Delta, 0)
	// LCS index pairs
	lcsPairs := lcsIndexPairs(left, right, reflect.DeepEqual)

	// list up items not in LCS, they are maybe deleted
	maybeDeleted := list.New() // but maybe moved or modified
	lcsI := 0
	for i, leftValue := range left {
		if lcsI < len(lcsPairs) && lcsPairs[lcsI].Left == i {
			lcsI++
		} else {
			maybeDeleted.PushBack(maybe{index: i, lcsIndex: lcsI, item: leftValue})
		}
	}

	// list up items not in LCS, they are maybe Added
	maybeAdded := list.New() // but maybe moved or modified
	lcsI = 0
	for i, rightValue := range right {
		if lcsI < len(lcsPairs) && lcsPairs[lcsI].Right == i {
			lcsI++
		} else {
			maybeAdded.PushBack(maybe{index: i, lcsIndex: lcsI, item: rightValue})
		}
	}

	// find moved items
	var delNext *list.Element // for prefetch to remove item in iteration
	for delCandidate := maybeDeleted.Front(); delCandidate != nil; delCandidate = delNext {
		delCan := delCandidate.Value.(maybe)
		delNext = delCandidate.Next()

		for addCandidate := maybeAdded.Front(); addCandidate != nil; addCandidate = addCandidate.Next() {
			addCan := addCandidate.Value.(maybe)
			if reflect.DeepEqual(delCan.item, addCan.item) {
				deltas = append(deltas, NewMoved(Index(delCan.index), Index(addCan.index), delCan.item))
				maybeAdded.Remove(addCandidate)
				maybeDeleted.Remove(delCandidate)
				break
			}
		}
	}

	// find modified or add+del
	prevIndexDel := 0
	prevIndexAdd := 0
	delElement := maybeDeleted.Front()
	addElement := maybeAdded.Front()
	for i := 0; i <= len(lcsPairs); i++ { // not "< len(lcsPairs)"
		var lcsPair lcsIndexPair
		var delSize, addSize int
		if i < len(lcsPairs) {
			lcsPair = lcsPairs[i]
			delSize = lcsPair.Left - prevIndexDel - 1
			addSize = lcsPair.Right - prevIndexAdd - 1
			prevIndexDel = lcsPair.Left
			prevIndexAdd = lcsPair.Right
		}

		var delSlice []maybe
		if delSize > 0 {
			delSlice = make([]maybe, 0, delSize)
		} else {
			delSlice = make([]maybe, 0, maybeDeleted.Len())
		}
		for ; delElement != nil; delElement = delElement.Next() {
			d := delElement.Value.(maybe)
			if d.lcsIndex != i {
				break
			}
			delSlice = append(delSlice, d)
		}

		var addSlice []maybe
		if addSize > 0 {
			addSlice = make([]maybe, 0, addSize)
		} else {
			addSlice = make([]maybe, 0, maybeAdded.Len())
		}
		for ; addElement != nil; addElement = addElement.Next() {
			a := addElement.Value.(maybe)
			if a.lcsIndex != i {
				break
			}
			addSlice = append(addSlice, a)
		}

		if len(delSlice) > 0 && len(addSlice) > 0 {
			var bestDeltas []Delta
			bestDeltas, delSlice, addSlice = maximizeSimilarities(delSlice, addSlice)
			for _, delta := range bestDeltas {
				deltas = append(deltas, delta)
			}
		}

		for _, del := range delSlice {
			deltas = append(deltas, NewDeleted(Index(del.index), del.item))
		}
		for _, add := range addSlice {
			deltas = append(deltas, NewAdded(Index(add.index), add.item))
		}
	}

	return deltas
}

func compareValues(position Position, left, right any) (same bool, delta Delta) {
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false, NewModified(position, left, right)
	}

	switch left.(type) {
	case map[string]any:
		l := left.(map[string]any)
		childDeltas := CompareObjects(l, right.(map[string]any))
		if len(childDeltas) > 0 {
			return false, NewObject(position, childDeltas)
		}

	case []any:
		l := left.([]any)
		childDeltas := compareArrays(l, right.([]any))

		if len(childDeltas) > 0 {
			return false, NewArray(position, childDeltas)
		}

	default:
		if !reflect.DeepEqual(left, right) {
			return false, NewModified(position, left, right)
		}
	}

	return true, nil
}

func maximizeSimilarities(left []maybe, right []maybe) (resultDeltas []Delta, freeLeft, freeRight []maybe) {
	deltaTable := make([][]Delta, len(left))
	for i := 0; i < len(left); i++ {
		deltaTable[i] = make([]Delta, len(right))
	}
	for i, leftValue := range left {
		for j, rightValue := range right {
			_, delta := compareValues(Index(rightValue.index), leftValue.item, rightValue.item)
			deltaTable[i][j] = delta
		}
	}

	sizeX := len(left) + 1 // margins for both sides
	sizeY := len(right) + 1

	// fill out with similarities
	dpTable := make([][]float64, sizeX)
	for i := 0; i < sizeX; i++ {
		dpTable[i] = make([]float64, sizeY)
	}
	for x := sizeX - 2; x >= 0; x-- {
		for y := sizeY - 2; y >= 0; y-- {
			prevX := dpTable[x+1][y]
			prevY := dpTable[x][y+1]
			score := deltaTable[x][y].Similarity() + dpTable[x+1][y+1]

			dpTable[x][y] = max(prevX, prevY, score)
		}
	}

	minLength := len(left)
	if minLength > len(right) {
		minLength = len(right)
	}
	maxInvalidLength := minLength - 1

	freeLeft = make([]maybe, 0, len(left)-minLength)
	freeRight = make([]maybe, 0, len(right)-minLength)

	resultDeltas = make([]Delta, 0, minLength)
	var x, y int
	for x, y = 0, 0; x <= sizeX-2 && y <= sizeY-2; {
		current := dpTable[x][y]
		nextX := dpTable[x+1][y]
		nextY := dpTable[x][y+1]

		xValidLength := len(left) - maxInvalidLength + y
		yValidLength := len(right) - maxInvalidLength + x

		if x+1 < xValidLength && current == nextX {
			freeLeft = append(freeLeft, left[x])
			x++
		} else if y+1 < yValidLength && current == nextY {
			freeRight = append(freeRight, right[y])
			y++
		} else {
			resultDeltas = append(resultDeltas, deltaTable[x][y])
			x++
			y++
		}
	}
	for ; x < sizeX-1; x++ {
		freeLeft = append(freeLeft, left[x-1])
	}
	for ; y < sizeY-1; y++ {
		freeRight = append(freeRight, right[y-1])
	}

	return resultDeltas, freeLeft, freeRight
}
