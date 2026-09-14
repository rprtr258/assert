# pp [![Go](https://github.com/rprtr258/assert/actions/workflows/go.yml/badge.svg)](https://github.com/rprtr258/assert/actions/workflows/go.yml)
Colored pretty printer for Go language — internal fork of [k0kubun/pp](https://github.com/k0kubun/pp), used by `assert` to render failure output.

![](http://i.gyazo.com/d3253ae839913b7239a7229caa4af551.png)

## Usage
This is an internal package, importable only within this module:
`github.com/rprtr258/assert/internal/pp`.

```go
m := map[string]string{"foo": "bar", "hello": "world"}
s := pp.Sprint(m) // formatted string

mypp := pp.New()
mypp.Print(m) // or print straight to stdout
```

![](http://i.gyazo.com/0d08376ed2656257627f79626d5e0cde.png)

### API
Package-level `pp.Sprint()` uses the shared `pp.Default` printer. Printers created
with `pp.New()` expose `Print` and `Sprint`.

### Custom colors
Syntax highlighting is done with ANSI codes from the internal `scuf` package; the
default color scheme is fixed (`scheme` in `printer.go`).

## Demo

### Timeline
![](http://i.gyazo.com/a8adaeec965db943486e35083cf707f2.png)

### UserStream event
![](http://i.gyazo.com/1e88915b3a6a9129f69fb5d961c4f079.png)

### Works on windows
![](http://i.gyazo.com/ab791997a980f1ab3ee2a01586efdce6.png)
