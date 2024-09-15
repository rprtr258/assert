## Example

```go
//go:embed testdata.txtar
var _txtar []byte

// expected strings by test name
var _expectedByObject map[string]string = golden.Load[string](_txtar)

// sometimes in tests update golden file
golden.Save("testdata.txtar", _expectedByObject)
```