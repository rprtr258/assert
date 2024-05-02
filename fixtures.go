package assert

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func UseFileContent(t testing.TB, filename string) []byte {
	content, err := os.ReadFile(filename)
	NoError(t, err)
	return content
}

func UseJSON[T any](t testing.TB, data []byte) T {
	var res T
	NoError(t, json.Unmarshal(data, &res))
	return res
}

func UseTempDir(t testing.TB) string {
	res, err := os.MkdirTemp(os.TempDir(), "")
	NoError(t, err)
	t.Cleanup(func() {
		NoError(t, os.RemoveAll(res))
	})
	return res
}

func UseFile(t testing.TB, filename string) *os.File {
	file, err := os.Open(filename)
	NoError(t, err)
	t.Cleanup(func() {
		NoError(t, file.Close())
	})
	return file
}

func UseTempFile(t testing.TB, content []byte) *os.File {
	// create and fill file
	file, err := os.CreateTemp(os.TempDir(), "")
	NoError(t, err)
	_, err = file.Write(content)
	NoError(t, err)
	_, err = file.Seek(0, io.SeekStart)
	NoError(t, err)
	t.Cleanup(func() {
		NoError(t, file.Close())
	})
	return file
}

// NOTE: envs are set for whole processes, so all parallel tests will see them
func UseEnv(t testing.TB, key, value string) {
	// cleanup is done inside
	t.Setenv(key, value)
}
