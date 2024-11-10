package assert

import (
	"fmt"

	"github.com/rprtr258/assert/internal/pp"
)

// T is the interface common to T, B, and F.
type T interface {
	Helper()
	Cleanup(func())
	Fail()
	FailNow()
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

var _ T = tT{}

type tT struct {
	T
	must bool
	kvs  []labeledContent
}

func (t tT) Cleanup(f func()) {
	t.T.Cleanup(f)
}

func (t tT) Fail() {
	t.T.Fail()
	if t.must {
		t.T.FailNow()
	}
	fail(t.T, t.kvs)
}

func (t tT) FailNow() {
	t.T.FailNow()
	fail(t.T, t.kvs)
}

func (t tT) Error(args ...any) {
	t.T.Error(args...)
	fail(t.T, t.kvs)
	if t.must {
		t.T.FailNow()
	}
}

func (t tT) Errorf(format string, args ...any) {
	t.T.Errorf(format, args...)
	fail(t.T, t.kvs)
	if t.must {
		t.T.FailNow()
	}
}

func (t tT) Fatal(args ...any) {
	t.T.Fatal(args...)
	fail(t, t.kvs)
}

func (t tT) Fatalf(format string, args ...any) {
	t.T.Fatalf(format, args...)
	fail(t, t.kvs)
}

// Must fails test immediately on fail.
func Must(t T) *tT {
	return &tT{
		T:    t,
		must: true,
		kvs:  nil,
	}
}

func Wrap(t T) *tT {
	return &tT{
		T:    t,
		must: false,
		kvs:  nil,
	}
}

func (t *tT) Msg(msg string) *tT {
	t.kvs = append(t.kvs, labeledContent{
		label:   "Message",
		content: msg,
	})
	return t
}

func (t *tT) Msgf(format string, args ...any) *tT {
	return t.Msg(fmt.Sprintf(format, args...))
}

func (t *tT) With(key string, value any) *tT {
	t.kvs = append(t.kvs, labeledContent{
		label:   key,
		content: pp.Sprint(value),
	})
	return t
}
