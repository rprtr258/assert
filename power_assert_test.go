package assert_test

import (
	"cmp"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/rprtr258/assert"
	"github.com/rprtr258/assert/internal/ass"

	"golang.org/x/tools/txtar"
)

// TestExample exercises power-assert diagram rendering. Every assertion below
// is intentionally false so its diagram is produced; instead of failing the
// suite, the diagrams are captured and compared against a golden txtar
// snapshot in testdata/power_assert.txtar.
//
// Regenerate the snapshot with:
//
//	ASSERT_UPDATE_SNAPSHOT=1 go test ./...
func TestExample(t *testing.T) {
	assert.SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZSnapshot = true
	assert.SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZCapturedSnapshots = nil

	assert.Assert(t, 2+2 == 5)

	two := 1 + 1
	assert.Assert(t, two+2 == 5)

	xs := []int{1, 2}
	assert.Assert(t, reflect.DeepEqual(append(xs, 3), []int{1, 2, 4}))
	assert.Assert(t, reflect.DeepEqual(append(xs, 3)[1:], []int{2, 4}))

	factorial := func(n int) int {
		res := 1
		for i := 2; i <= n; i++ {
			res *= i
		}
		return res
	}
	assert.Assert(t, factorial(5) == 60)

	assert.Assert(t, xs[1] == 1)
	assert.Assert(t, -1 == 1)
	assert.Assert(t, (1+1) == 1)

	var s struct{ x struct{ y int } }
	assert.Assert(t, &s == nil)
	assert.Assert(t, s.x.y != 0)

	assert.Assert(t, (*int)(nil) != nil)
	assert.Assert(t, len(make([]int, 0)) == 1)
	assert.Assert(t, new(struct{}) == nil)
	assert.Assert(t, *new(int) == 1)
	assert.Assert(t, len(map[int]int{1: two}) == 0)
	assert.Assert(t, &s == &s)

	t.Run("require", func(t *testing.T) {
		assert.Require(t, two != 1+1)
	})

	// The rewritten second pass runs from a temp copy; ASSERT_MODULE_DIR points
	// at the real module tree where the golden file lives.
	moduleDir := cmp.Or(os.Getenv("ASSERT_MODULE_DIR"), ".")
	goldenPath := filepath.Join(moduleDir, "testdata", "power_assert.txtar")

	got := assert.SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZCapturedSnapshots

	if os.Getenv("ASSERT_UPDATE_SNAPSHOT") == "1" {
		files := make([]txtar.File, len(got))
		for i, diagram := range got {
			files[i] = txtar.File{
				Name: strconv.Itoa(i),
				Data: []byte(diagram + "\n"),
			}
		}
		ass.NoError(t, os.WriteFile(goldenPath, txtar.Format(&txtar.Archive{Files: files}), 0o644))
		return
	}

	archive, err := txtar.ParseFile(goldenPath)
	ass.NoError(t, err)
	ass.Equal(t, len(archive.Files), len(got))

	for i, diagram := range got {
		golden := strings.TrimSpace(string(archive.Files[i].Data))
		ass.Equal(t, golden, diagram)
	}
}
