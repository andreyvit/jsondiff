package jsondiff

import (
	"encoding/json"
	"testing"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected string
	}{
		{
			name:  "base",
			left:  `{"str": "abcde", "num_int": 13, "num_float": 39.39, "bool": true, "arr": ["arr0", 21, {"str": "pek3f", "num": 1},   [0, "1"]],       "obj": {"str": "bcded", "num": 19,      "arr": [17, "str", {"str": "eafeb"}],   "obj": {"str": "efj3", "num": 14} }, "null": null}`,
			right: `{"str": "abcde", "num_int": 13, "num_float": 39.39, "bool": true, "arr": ["arr0", 21, {"str": "changed", "num": 1}, [0, "changed"]], "obj": {"str": "bcded", "new": "added", "arr": [17, "str", {"str": "changed"}], "obj": {"str": "changed", "num": 9999}}}`,
			expected: ` {
   "arr": [
     "arr0",
     21,
     {
       "num": 1,
-      "str": "pek3f"
+      "str": "changed"
     },
     [
       0,
-      "1"
+      "changed"
     ]
   ],
   "bool": true,
-  "null": null,
   "num_float": 39.39,
   "num_int": 13,
   "obj": {
     "arr": [
       17,
       "str",
       {
-        "str": "eafeb"
+        "str": "changed"
       }
     ],
-    "num": 19,
     "obj": {
-      "num": 14,
+      "num": 9999,
-      "str": "efj3"
+      "str": "changed"
     },
     "str": "bcded"
+    "new": "added"
   },
   "str": "abcde"
 }`,
		},
		{
			name:  "add_delete",
			left:  `{"delete": {"l0o": {"l1s": "abcd", "l1o": {"l2s": "efed" } }, "l0a": ["abcd", ["efcg" ] ] } }`,
			right: `{"add": {"l0o": {"l1s": "abcd", "l1o": {"l2s": "efed" } }, "l0a": ["abcd", ["efcg" ] ] } }`,
			expected: ` {
-  "delete": {
-    "l0a": [
-      "abcd",
-      [
-        "efcg"
-      ]
-    ],
-    "l0o": {
-      "l1o": {
-        "l2s": "efed"
-      },
-      "l1s": "abcd"
-    }
-  }
+  "add": {
+    "l0a": [
+      "abcd",
+      [
+        "efcg"
+      ]
+    ],
+    "l0o": {
+      "l1o": {
+        "l2s": "efed"
+      },
+      "l1s": "abcd"
+    }
+  }
 }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := diff(tt.left, tt.right)
			if actual != tt.expected {
				t.Errorf("** DIFF:\n%q\n\nEXPECTED:\n%q", actual, tt.expected)
				t.Errorf("** DIFF:\n%s\n\nEXPECTED:\n%s", actual, tt.expected)
			}
		})
	}
}

func diff(left, right string) string {
	var v1, v2 map[string]any
	ensure(json.Unmarshal([]byte(left), &v1))
	ensure(json.Unmarshal([]byte(right), &v2))
	return CompareObjects(v1, v2).Format(v1)
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}
