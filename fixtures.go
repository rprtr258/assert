package assert

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func UseFileContent(tb testing.TB, filename string) []byte {
	tb.Helper()
	content, err := os.ReadFile(filename)
	NoError(tb, err)
	return content
}

func UseJSON[T any](tb testing.TB, data []byte) T {
	tb.Helper()
	var res T
	NoError(tb, json.Unmarshal(data, &res))
	return res
}

func UseTempDir(tb testing.TB) string {
	tb.Helper()
	res, err := os.MkdirTemp(os.TempDir(), "")
	NoError(tb, err)
	tb.Cleanup(func() {
		NoError(tb, os.RemoveAll(res))
	})
	return res
}

func UseFile(tb testing.TB, filename string) *os.File {
	tb.Helper()
	file, err := os.Open(filename)
	NoError(tb, err)
	tb.Cleanup(func() {
		NoError(tb, file.Close())
	})
	return file
}

func UseTempFile(tb testing.TB, content []byte) *os.File {
	tb.Helper()
	// create and fill file
	file, err := os.CreateTemp(os.TempDir(), "")
	NoError(tb, err)
	_, err = file.Write(content)
	NoError(tb, err)
	_, err = file.Seek(0, io.SeekStart)
	NoError(tb, err)
	tb.Cleanup(func() {
		NoError(tb, file.Close())
	})
	return file
}

// NOTE: envs are set for whole processes, so all parallel tests will see them
func UseEnv(tb testing.TB, key, value string) {
	tb.Helper()
	// cleanup is done inside
	tb.Setenv(key, value)
}
