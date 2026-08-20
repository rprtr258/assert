# assert library optimization TODO

Target: `github.com/rprtr258/assert` (v0.1.0)
Source reference: `~/go/pkg/mod/github.com/rprtr258/assert@v0.1.0/`

## Background / root cause

`Assert`/`Require` (`power_assert.go:16,22`) call `fuse()` → `run()` (`power_assert_private.go:244`).
`run()` executes **on every call**:

1. `runtime.Caller` walk to find `go.mod`
2. `os.CopyFS` the entire module to a random temp dir (including `.git`, vendor, all sources)
3. `os.Chdir(tmpDir)` — process-global mutation
4. Parse + AST-rewrite every `_test.go` in the module
5. `exec.Command("go", "test", "./...")` — re-run the entire module's test suite as a subprocess
6. `tb.SkipNow()` so the original test reports `SKIP`

So N assertions = N full copy→rewrite→recompile→whole-suite runs. This is why
`TestGenerate`/`TestFind` in `fun/iter` take seconds while all `assert.Equal`/`True` tests
are sub-millisecond.

The design intends a one-time `Fuse()` (the rewriter strips `pa.Fuse()` calls at
`power_assert_private.go:312`), but the shipped `Assert`/`Require` re-execute per call.

## PRs (keep separate)

### PR1 — Fix `os.Chdir` race with `t.Parallel` (correctness) — DONE

`os.Chdir(tmpDir)` (`power_assert_private.go:~262`) mutates process-global state while
parallel tests run. Latent race; also causes bogus interleaved SKIP output in `-v` logs
(subprocess stdout pipes into parent).

Replace parent-side chdir with `cmd.Dir = tmpDir` on the `exec.Command` at
`power_assert_private.go:~397`. No parent-side `os.Chdir`.

Verify: run `go test -v -count=1 -race ./...` on a package using `assert.Assert` under
`t.Parallel()`; confirm no `-race` failure and clean `-v` output.

### PR2 — Scope re-exec to the current package

`exec.Command("go", "test", "./...")` (`power_assert_private.go:~397`) re-runs the whole
module. The `// TODO: pass args` acknowledges this.

Change to `go test .` (the package being tested). One package re-exec instead of a
module-wide one.

Verify: timing of a single `assert.Assert` call drops from "whole module test" to
"one package test".

### PR3 — Link instead of copy; skip non-source dirs; pre-scan before parse

`os.CopyFS(tmpDir, os.DirFS(moduleDir))` duplicates the entire tree including `.git`,
`vendor`, build dirs. The existing TODO: "copy _test.go files, link everything besides".

- Use `os.Link` for non-test files; fall back to `os.CopyFS` copy across devices.
- Skip `.git`, `vendor`, build output dirs.
- Before `parser.ParseFile`, guard with `bytes.Contains(src, []byte("assert.Assert")) ||
  bytes.Contains(src, []byte("assert.Require"))` to skip files with no targets.

Verify: temp dir size and parse time drop; rewritten output unchanged for a sample
test file.

### PR4 — Fuse once per package, not per-call (structural, API change)

Make `Assert`/`Require` plain runtime checks against pre-rewritten code, driven by a
single `assert.Fuse(m)` in `TestMain(m *testing.M)`. `Fuse` does the copy+rewrite+reexec
**once per package**; the subprocess runs the real suite with `Assert` already rewritten
to `ZZZAssert`.

- The AST-stripping code for `Fuse` already exists (`power_assert_private.go:312`); make
  it the only entry point.
- If `Fuse` wasn't called, `Assert` should error with a clear "add assert.Fuse(m) to
  TestMain" message, or fall back to a non-power assertion.
- This is O(1) re-execs per package instead of O(assertions).

Verify: a package with K `assert.Assert` calls runs the subprocess once, not K times;
per-test timing approaches the non-power baseline.

### PR5 — Content-addressed cache

A random `assert.*` temp dir defeats Go's build/test cache. Use a stable dir keyed by
module path + hash of all `_test.go` contents.

- On cache hit: skip copy+parse+rewrite entirely; warm `go test` cache makes re-exec
  near-instant.
- On cache miss: populate it.

Verify: second `go test -count=1` run of an unchanged package skips rewrite and hits
the go test cache.

### RFC — Move rewriting out of runtime (v0.2 direction)

The runtime self-re-exec model is the unusual choice; power-assert in other languages
does compile-time/source rewriting. Propose a `cmd/assert-rewrite` or `go:generate`
directive that rewrites test files on disk once. `Assert` becomes a normal check with
zero runtime overhead and no subprocess. `Fuse` (PR4) is the v0.1.x bridge to this.

Or use toolexec, e.g. see https://opentelemetry.io/docs/zero-code/go/compile-time/getting-started/#keep-using-go-build

## Reference: which call sites are slow today

`assert.Assert`/`assert.Require` users (the slow ones):
- `fun/iter/iter_test.go:33,36,222` — `TestGenerate`, `TestFind`
- `fun/orderedmap/orderedmap_test.go:29,33,36`
- `fun/exp/json/json_test.go:50`

`assert.Equal`/`True`/`False` (`assert.go:132,296,313`) do NOT call `fuse()` — they
compare in-place and return immediately. They are not affected.
