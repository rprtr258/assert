# UNDER DEVELOPMENT REQUIRES HEAVY **TESTING**

## Why
- generics are used where possible
- meaningful diffs instead of bunch of lines(that is `git-diff(actual, expected)`) somewhere from values string representations, tree diffs are used instead, so you can see exactly where values do not match
- in unequal strings unprintable characters are escaped appropriately
- less dependencies than `stretchr/testify`
- fixtures api: use test resources and have them destroyed automatically:
  - env vars
  - file content
  - parsed json
  - temp file
  - temp dir
  - `io.Reader` contents
- power assert
- no api mirroring for `assert`ing or `require`-ing and failing immediately after check
- golden files support
- pretty and colourful test output
- no `Expect(ACTUAL).To(Equal(EXPECTED))` [nonsense](https://github.com/onsi/gomega) rewriting of simple `ACTUAL == EXPECTED`, just use `assert.Equal(t, ACTUAL, EXPECTED)` or `assert.Assert(t, ACTUAL == EXPECTED)` and see values used in case of failure (dark magic inside)

## Comparison with other libraries
|features|[rprtr258/assert](https://github.com/rprtr258/assert)|[stretchr/testify](https://github.com/stretchr/testify)|[shoenig/test](https://github.com/shoenig/test)|[alecthomas/assert](https://github.com/alecthomas/assert)|
|-|-|-|-|-|
|minimal dependencies|:white_check_mark:|:x:|:white_check_mark:|:white_check_mark:|
|generics|:white_check_mark:|:x:|:white_check_mark:|:white_check_mark:|
|meaningful diffs|:white_check_mark:|:x:|:white_check_mark:|:question:|
|power assert|:white_check_mark:|:x:|:x:|:x:|
|fixtures|:white_check_mark:|:x:|:white_check_mark:[^1]|:x:|
|no api mirroring|:white_check_mark:|:x:|:x:|:white_check_mark:|
|golden files support|:white_check_mark:|:x:|:x:|:x:|

[^1]: temp file and port only
