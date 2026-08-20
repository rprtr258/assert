package assert

import (
	"log"
	"os"
	"testing"
)

// Fuse rewrites the current package's _test.go files once (assert.Assert/
// Require -> direct power-assert checks) and re-execs them as a subprocess.
// In the rewritten subprocess the assert.Fuse(m) call is stripped, so the
// real test suite runs once with the assertions already inlined.
//
// Fuse does not return in the parent process. Call it from TestMain:
//
//	func TestMain(m *testing.M) {
//	    assert.Fuse(m)
//	    os.Exit(m.Run())
//	}
func Fuse(m *testing.M) {
	code, err := run()
	if err != nil {
		log.Fatalln(err.Error())
	}
	os.Exit(code)
}

// Assert is rewritten to a direct power-assert check by assert.Fuse. Reaching
// it at runtime means Fuse was not wired into TestMain; fail loudly with an
// actionable message instead of silently degrading or re-exec'ing per call.
func Assert(tb testing.TB, cond bool) {
	tb.Helper()
	tb.Fatal("assert.Assert used without assert.Fuse: add `assert.Fuse(m)` to a TestMain(m *testing.M) in this package (see readme.md)")
}

// Require is rewritten to a direct power-assert check by assert.Fuse. See Assert.
func Require(tb testing.TB, cond bool) {
	tb.Helper()
	tb.Fatal("assert.Require used without assert.Fuse: add `assert.Fuse(m)` to a TestMain(m *testing.M) in this package (see readme.md)")
}

