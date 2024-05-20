package jsondiff

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type FormatOption int

const (
	ShowArrayIndex FormatOption = iota
	Colored
)

func (diff Diff) Format(left any, opts ...FormatOption) string {
	f := asciiFormatter{left: left}
	for _, opt := range opts {
		switch opt {
		case ShowArrayIndex:
			f.config.ShowArrayIndex = true
		case Colored:
			f.config.Coloring = true
		}
	}
	if v, ok := f.left.(map[string]any); ok {
		f.formatObject(v, diff)
	} else if v, ok := f.left.([]any); ok {
		f.formatArray(v, diff)
	} else {
		panic(fmt.Errorf("expected map[string]any or []any, got %T",
			f.left))
	}
	return strings.TrimRight(f.buffer.String(), "\n")
}

type asciiFormatter struct {
	left    any
	config  asciiFormatterConfig
	buffer  bytes.Buffer
	path    []string
	size    []int
	inArray []bool
	line    *asciiLine
}

type asciiFormatterConfig struct {
	ShowArrayIndex bool
	Coloring       bool
}

type asciiLine struct {
	marker string
	indent int
	buffer *bytes.Buffer
}

func (f *asciiFormatter) formatObject(left map[string]any, df Diff) {
	f.addLineWith(AsciiSame, "{")
	f.push("ROOT", len(left), false)
	f.processObject(left, df)
	f.pop()
	f.addLineWith(AsciiSame, "}")
}

func (f *asciiFormatter) formatArray(left []any, df Diff) {
	f.addLineWith(AsciiSame, "[")
	f.push("ROOT", len(left), true)
	f.processArray(left, df)
	f.pop()
	f.addLineWith(AsciiSame, "]")
}

func (f *asciiFormatter) processArray(array []any, deltas []Delta) error {
	patchedIndex := 0
	for index, value := range array {
		f.processItem(value, deltas, Index(index))
		patchedIndex++
	}

	// additional Added
	for _, delta := range deltas {
		switch delta.(type) {
		case *Added:
			d := delta.(*Added)
			// skip items already processed
			if int(d.Position.(Index)) < len(array) {
				continue
			}
			f.printRecursive(d.Position.String(), d.Value, AsciiAdded)
		}
	}

	return nil
}

func (f *asciiFormatter) processObject(object map[string]any, deltas []Delta) error {
	names := sortedKeys(object)
	for _, name := range names {
		value := object[name]
		f.processItem(value, deltas, Name(name))
	}

	// Added
	for _, delta := range deltas {
		switch delta.(type) {
		case *Added:
			d := delta.(*Added)
			f.printRecursive(d.Position.String(), d.Value, AsciiAdded)
		}
	}

	return nil
}

func (f *asciiFormatter) processItem(value any, deltas []Delta, position Position) error {
	matchedDeltas := filterDeltasByPosition(deltas, position)
	positionStr := position.String()
	if len(matchedDeltas) > 0 {
		for _, matchedDelta := range matchedDeltas {

			switch matchedDelta.(type) {
			case *Object:
				d := matchedDelta.(*Object)
				switch value.(type) {
				case map[string]any:
					//ok
				default:
					return errors.New("Type mismatch")
				}
				o := value.(map[string]any)

				f.newLine(AsciiSame)
				f.printKey(positionStr)
				f.print("{")
				f.closeLine()
				f.push(positionStr, len(o), false)
				f.processObject(o, d.Deltas)
				f.pop()
				f.newLine(AsciiSame)
				f.print("}")
				f.printComma()
				f.closeLine()

			case *Array:
				d := matchedDelta.(*Array)
				switch value.(type) {
				case []any:
					//ok
				default:
					return errors.New("Type mismatch")
				}
				a := value.([]any)

				f.newLine(AsciiSame)
				f.printKey(positionStr)
				f.print("[")
				f.closeLine()
				f.push(positionStr, len(a), true)
				f.processArray(a, d.Deltas)
				f.pop()
				f.newLine(AsciiSame)
				f.print("]")
				f.printComma()
				f.closeLine()

			case *Added:
				d := matchedDelta.(*Added)
				f.printRecursive(positionStr, d.Value, AsciiAdded)
				f.size[len(f.size)-1]++

			case *Modified:
				d := matchedDelta.(*Modified)
				savedSize := f.size[len(f.size)-1]
				f.printRecursive(positionStr, d.OldValue, AsciiDeleted)
				f.size[len(f.size)-1] = savedSize
				f.printRecursive(positionStr, d.NewValue, AsciiAdded)

			case *Deleted:
				d := matchedDelta.(*Deleted)
				f.printRecursive(positionStr, d.Value, AsciiDeleted)

			default:
				return errors.New("Unknown Delta type detected")
			}

		}
	} else {
		f.printRecursive(positionStr, value, AsciiSame)
	}

	return nil
}

func filterDeltasByPosition(deltas []Delta, position Position) (results []Delta) {
	results = make([]Delta, 0)
	for _, delta := range deltas {
		if delta.PositionMatches(position) {
			results = append(results, delta)
		}
	}
	return
}

const (
	AsciiSame    = " "
	AsciiAdded   = "+"
	AsciiDeleted = "-"
)

var AsciiStyles = map[string]string{
	AsciiAdded:   "30;42",
	AsciiDeleted: "30;41",
}

func (f *asciiFormatter) push(name string, size int, array bool) {
	f.path = append(f.path, name)
	f.size = append(f.size, size)
	f.inArray = append(f.inArray, array)
}

func (f *asciiFormatter) pop() {
	f.path = f.path[0 : len(f.path)-1]
	f.size = f.size[0 : len(f.size)-1]
	f.inArray = f.inArray[0 : len(f.inArray)-1]
}

func (f *asciiFormatter) addLineWith(marker string, value string) {
	f.line = &asciiLine{
		marker: marker,
		indent: len(f.path),
		buffer: bytes.NewBufferString(value),
	}
	f.closeLine()
}

func (f *asciiFormatter) newLine(marker string) {
	f.line = &asciiLine{
		marker: marker,
		indent: len(f.path),
		buffer: bytes.NewBuffer([]byte{}),
	}
}

func (f *asciiFormatter) closeLine() {
	style, ok := AsciiStyles[f.line.marker]
	if f.config.Coloring && ok {
		f.buffer.WriteString("\x1b[" + style + "m")
	}

	f.buffer.WriteString(f.line.marker)
	for n := 0; n < f.line.indent; n++ {
		f.buffer.WriteString("  ")
	}
	f.buffer.Write(f.line.buffer.Bytes())

	if f.config.Coloring && ok {
		f.buffer.WriteString("\x1b[0m")
	}

	f.buffer.WriteRune('\n')
}

func (f *asciiFormatter) printKey(name string) {
	if !f.inArray[len(f.inArray)-1] {
		fmt.Fprintf(f.line.buffer, `"%s": `, name)
	} else if f.config.ShowArrayIndex {
		fmt.Fprintf(f.line.buffer, `%s: `, name)
	}
}

func (f *asciiFormatter) printComma() {
	f.size[len(f.size)-1]--
	if f.size[len(f.size)-1] > 0 {
		f.line.buffer.WriteRune(',')
	}
}

func (f *asciiFormatter) printValue(value any) {
	switch value.(type) {
	case string:
		fmt.Fprintf(f.line.buffer, `"%s"`, value)
	case nil:
		f.line.buffer.WriteString("null")
	default:
		fmt.Fprintf(f.line.buffer, `%#v`, value)
	}
}

func (f *asciiFormatter) print(a string) {
	f.line.buffer.WriteString(a)
}

func (f *asciiFormatter) printRecursive(name string, value any, marker string) {
	switch value.(type) {
	case map[string]any:
		f.newLine(marker)
		f.printKey(name)
		f.print("{")
		f.closeLine()

		m := value.(map[string]any)
		size := len(m)
		f.push(name, size, false)

		keys := sortedKeys(m)
		for _, key := range keys {
			f.printRecursive(key, m[key], marker)
		}
		f.pop()

		f.newLine(marker)
		f.print("}")
		f.printComma()
		f.closeLine()

	case []any:
		f.newLine(marker)
		f.printKey(name)
		f.print("[")
		f.closeLine()

		s := value.([]any)
		size := len(s)
		f.push("", size, true)
		for _, item := range s {
			f.printRecursive("", item, marker)
		}
		f.pop()

		f.newLine(marker)
		f.print("]")
		f.printComma()
		f.closeLine()

	default:
		f.newLine(marker)
		f.printKey(name)
		f.printValue(value)
		f.printComma()
		f.closeLine()
	}
}

func sortedKeys(m map[string]any) (keys []string) {
	keys = make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}
