package assert

import (
	"log"
	"testing"
)

// Fuse - call in TestMain before using Assert
func fuse(tb testing.TB) {
	tb.Helper()
	if err := run(); err != nil {
		log.Fatalln(err.Error())
	}
	tb.SkipNow()
}

func Assert(tb testing.TB, cond bool) {
	tb.Helper()
	fuse(tb)
}

func Require(tb testing.TB, cond bool) {
	tb.Helper()
	fuse(tb)
}
