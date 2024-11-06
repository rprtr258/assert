package pa

import (
	"log"
	"testing"
)

// Fuse - call in TestMain before using Assert
func Fuse() {
	if err := run(); err != nil {
		log.Fatalln(err.Error())
	}
}

func Assert(tb testing.TB, cond bool) {
	tb.Helper()
	if !cond {
		tb.Fail()
	}
}

func Require(tb testing.TB, cond bool) {
	tb.Helper()
	if !cond {
		tb.FailNow()
	}
}
