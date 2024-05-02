Usage:

Add to `go.mod`:
```
replace (
    github.com/stretchr/testify/assert => github.com/rprtr258/assert
)
```

## Why
- generics are used, to check types at compile time, instead of runtime
- meaningful diffs instead of bunch of lines(that is `git-diff(actual, expected)`) somewhere from values string representations
- tree diffs, so you can see exactly where values do not match
- in unequal strings unprintable characters are escaped appropriately
- less dependencies
- fixtures api: use test resources and have them destroyed automatically