package golden

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	"golang.org/x/tools/txtar"
)

func Load[T any](b []byte) map[string]T {
	files := txtar.Parse(b).Files

	tests := make(map[string]T, len(files))
	for _, file := range files {
		// check no object is not saved multiple times
		if _, ok := tests[file.Name]; ok {
			panic(fmt.Sprintf("duplicate file name: %s", file.Name))
		}

		// remove trailing \n
		data := file.Data[:len(file.Data)-1]
		var testcase T
		if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&testcase); err != nil {
			panic(fmt.Sprintf("failed to decode %q: %s", file.Name, err.Error()))
		}
		tests[file.Name] = testcase
	}
	return tests
}

func Save[T any](
	filename string,
	testcases map[string]T,
) {
	files := make([]txtar.File, 0, len(testcases))
	for name, testcase := range testcases {
		var bb bytes.Buffer
		if err := gob.NewEncoder(&bb).Encode(testcase); err != nil {
			panic(fmt.Sprintf("failed to encode %q: %s", name, err.Error()))
		}
		files = append(files, txtar.File{
			Name: name,
			Data: bb.Bytes(),
		})
	}

	if err := os.WriteFile(filename, txtar.Format(&txtar.Archive{
		Comment: nil,
		Files:   files,
	}), 0644); err != nil {
		panic(fmt.Sprintf("failed to write file %q: %s", filename, err.Error()))
	}
}
