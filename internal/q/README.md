# q
Internal fork of [ryboe/q](https://github.com/ryboe/q).

Instead of print-debugging, `Q` extracts the source text of the arguments of a function call at the caller's location.

`assert` uses this to show the actual argument expressions in failure messages. For example, when
```go
assert.Equal(t, got, want)
```

fails, `assert.Equal` calls
```go
q.Q("assert", "Equal") // []string{"got", "want"}
```

and the failure output shows the values by their original names.

## How it works
1. `runtime.Caller` with `CallDepth = 2` resolves the file and line of the caller's call site.
2. That source file is parsed and the `<pkgName>.<funcName>(...)` (or bare `<funcName>(...)`) call on that line is located.
3. The source text of each argument expression is returned, e.g. `assert.Equal(t, got, 42)` yields `[]string{"t", "got", "42"}`.

## API
```go
// Q returns the source text of each argument of the <pkgName>.<funcName>(...)
// (or bare <funcName>(...)) call on the caller's line.
func Q(pkgName, funcName string) []string
```
