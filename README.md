# jsondiff

Zero-dependencies simple JSON diffing and formatting library for Go

This is a derivative of [yudai/gojsondiff](https://github.com/yudai/gojsondiff/tree/master) and [yudai/golcs](https://github.com/yudai/golcs/tree/master), removing all the complexity, dependencies and fancy options. Also, notably, this library removes the ability to apply diffs, focusing entirely on diffing for presentation purposes.


## Usage

```go
import (
    "github.com/andreyvit/jsondiff"
)

func main() {
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
}
```

Outputs:

```
 {
-  "bar": 20,
+  "bar": 42,
-  "boz": 30,
   "foo": 10
 }
```

Use `diff.Format(before, jsondiff.Colored)` to add some ANSI colors for printing.
