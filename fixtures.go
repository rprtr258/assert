package assert

import (
	"encoding/json"
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
