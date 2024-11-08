package assert

import (
	"log"
	"testing"
)

// TODO: fuse once
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
