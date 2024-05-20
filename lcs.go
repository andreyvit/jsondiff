package jsondiff

type (
	// lcsIndexPair represents an pair of indices in the left and right arrays.
	lcsIndexPair struct {
		Left  int
		Right int
	}
)

func lcsBuildTable[T any](left, right []T, eq func(a, b T) bool) [][]int {
	sizeX := len(left) + 1
	sizeY := len(right) + 1

	table := make([][]int, sizeX)
	for x := 0; x < sizeX; x++ {
		table[x] = make([]int, sizeY)
	}

	for y := 1; y < sizeY; y++ {
		for x := 1; x < sizeX; x++ {
			increment := 0
			if eq(left[x-1], right[y-1]) {
				increment = 1
			}
			table[x][y] = max(table[x-1][y-1]+increment, table[x-1][y], table[x][y-1])
		}
	}
	return table
}

func lcsLength[T any](left, right []T, eq func(a, b T) bool) int {
	table := lcsBuildTable(left, right, eq)
	return table[len(left)][len(right)]
}

func lcsIndexPairs[T any](left, right []T, eq func(a, b T) bool) []lcsIndexPair {
	table := lcsBuildTable(left, right, eq)
	pairs := make([]lcsIndexPair, table[len(table)-1][len(table[0])-1])

	for x, y := len(left), len(right); x > 0 && y > 0; {
		if eq(left[x-1], right[y-1]) {
			pairs[table[x][y]-1] = lcsIndexPair{Left: x - 1, Right: y - 1}
			x--
			y--
		} else {
			if table[x-1][y] >= table[x][y-1] {
				x--
			} else {
				y--
			}
		}
	}

	return pairs
}
