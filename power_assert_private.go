package assert

// BEHOLD: api for generated code, do not use

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/rprtr258/assert/internal/pp"
)

var (
	pkg     = ast.NewIdent("assert")
	pkgRoot = &ast.SelectorExpr{
		X:   pkg,
		Sel: ast.NewIdent("SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__"),
	}
)

type shit struct {
	// Snapshot mode: instead of failing the test, Assert/Require collect their
	// failure diagrams into ZZZCapturedSnapshots. Used by tests that exercise the power-assert
	// output without failing the suite.
	ZZZSnapshot          bool
	ZZZCapturedSnapshots []string

	ZZZNew     func(exprStr string) *assertData
}

var SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__ = shit{
	ZZZNew: func(exprStr string) *assertData {
		return &assertData{exprStr: exprStr}
	},
}

func (shit) ZZZAssert(tb testing.TB, assertData *assertData, cond bool) {
	tb.Helper()
	assert(tb, assertData, cond, "assert", tb.Fail)
}

func (shit) ZZZRequire(tb testing.TB, assertData *assertData, cond bool) {
	tb.Helper()
	assert(tb, assertData, cond, "require", tb.FailNow)
}

type expr struct {
	valueStr string
	position int
}

type assertData struct {
	exprs   []expr
	exprStr string
}

func (shit) ZZZAdd[T any](a *assertData, position int, value T) T {
	// TODO: pretty print in one line, no trailing commas
	s := pp.Sprint(value)
	s = strings.ReplaceAll(s, "\n    ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, ",}", "}")

	a.exprs = append(a.exprs, expr{s, position})
	return value
}

func assert(tb testing.TB, assertData *assertData, cond bool, fn string, onfail func()) {
	tb.Helper()
	if cond {
		return
	}

	slices.SortFunc(assertData.exprs, func(a, b expr) int {
		return a.position - b.position
	})

	var s strings.Builder
	s.WriteString(assertData.exprStr)
	s.WriteString("\n")
	for i, expr := range assertData.exprs {
		n := expr.position
		if i > 0 {
			n -= assertData.exprs[i-1].position + 1
		}
		s.WriteString(strings.Repeat(" ", n))
		s.WriteString("^")
	}
	for i, e := range slices.Backward(assertData.exprs) {
		s.WriteString("\n")
		for j := 0; j <= i; j++ {
			n := assertData.exprs[j].position
			if j > 0 {
				n -= assertData.exprs[j-1].position + 1
			}
			s.WriteString(strings.Repeat(" ", n))
			if j < i {
				s.WriteString("|")
			} else {
				s.WriteString(e.valueStr)
			}
		}
	}

	out := fn + " failed:\n" + s.String()
	if SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZSnapshot {
		SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZCapturedSnapshots = append(SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZCapturedSnapshots, out)
		return
	}
	defer onfail()

	// TODO: reports wrong line number, fix it
	tb.Errorf("%s", out)
}

const debug = false // TODO: make configurable, default to false

func debugf(format string, args ...any) {
	if !debug {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

func sprintCode(n ast.Node) string {
	buf := &bytes.Buffer{}
	_ = printer.Fprint(buf, token.NewFileSet(), n)
	return buf.String()
}

func dumpExpr(n ast.Expr, pos token.Pos) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   pkgRoot,
			Sel: ast.NewIdent("ZZZAdd"),
		},
		Args: []ast.Expr{
			ast.NewIdent("zzz"),
			&ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.Itoa(int(pos)),
			},
			n,
		},
	}
}

func rewriteExpr(n ast.Expr, offset token.Pos) ast.Expr {
	switch n := n.(type) {
	case nil:
		return nil
	case *ast.BasicLit:
		return n
	case *ast.ArrayType, *ast.StructType, *ast.FuncType, *ast.InterfaceType, *ast.MapType, *ast.ChanType:
		return n
	case *ast.Ident:
		switch n.Name {
		case "false", "true",
			"nil", "string", "byte", "uintptr",
			"int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64",
			"complex64", "complex128":
			return n
		default:
			return dumpExpr(n, n.Pos()-offset)
		}
	case *ast.CompositeLit:
		for i, e := range n.Elts {
			n.Elts[i] = rewriteExpr(e, offset)
		}
		return dumpExpr(n, n.Pos()-offset)
	case *ast.SelectorExpr:
		n.X = rewriteExpr(n.X, offset)
		return dumpExpr(n, n.Sel.NamePos-offset)
	case *ast.ParenExpr: // TODO: breaks position if optional, see (1+1) == 1 example
		n.X = rewriteExpr(n.X, offset)
		return dumpExpr(n, n.Lparen-offset)
	case *ast.SliceExpr:
		n.Low = rewriteExpr(n.Low, offset)
		n.High = rewriteExpr(n.High, offset)
		n.Max = rewriteExpr(n.Max, offset)
		n.X = rewriteExpr(n.X, offset)
		return dumpExpr(n, n.Lbrack-offset)
	case *ast.IndexExpr:
		n.Index = rewriteExpr(n.Index, offset)
		n.X = rewriteExpr(n.X, offset)
		return dumpExpr(n, n.Lbrack-offset)
	case *ast.UnaryExpr:
		if n.Op == token.AND {
			// TODO: actually invalid, see &s == &s example, which fails due to taking address of copies of s
			// needs testing
			return dumpExpr(n, n.OpPos-offset)
		}
		n.X = rewriteExpr(n.X, offset)
		return dumpExpr(n, n.OpPos-offset)
	case *ast.BinaryExpr:
		n.X = rewriteExpr(n.X, offset)
		n.Y = rewriteExpr(n.Y, offset)
		return dumpExpr(n, n.OpPos-offset)
	case *ast.CallExpr:
		for i, e := range n.Args {
			n.Args[i] = rewriteExpr(e, offset)
		}
		return dumpExpr(n, n.Pos()-offset)
	case *ast.StarExpr:
		n.X = rewriteExpr(n.X, offset)
		return dumpExpr(n, n.Star-offset)
	case *ast.BadExpr, *ast.FuncLit:
		return n
	case *ast.KeyValueExpr:
		n.Key = rewriteExpr(n.Key, offset)
		n.Value = rewriteExpr(n.Value, offset)
		return n
	default:
		log.Fatalf("unsupported expr type %T", n)
	}
	panic("unreachable")
}

func getModuleDir() (string, string, error) {
	// On the second (rewritten) pass we run from a temp copy of the module, so
	// runtime.Caller would point at the copy. The first pass records the real
	// module dir in ASSERT_MODULE_DIR; prefer it when set.
	if dir := os.Getenv("ASSERT_MODULE_DIR"); dir != "" {
		return dir, "", nil
	}
	// os.Getwd does not cut it, since tests are being run from a temp dir using temporary executable
	// so we have to do caller getting trickery and extract module path the hard way
	_, file, _, ok := runtime.Caller(4) // Assert/Require -> fuse -> run -> getModuleDir
	if !ok {
		return "", "", errors.New("could not get caller, check sources are available")
	}
	callerDir := filepath.Dir(file)

	dir := callerDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}

		if dir == "/" {
			return "", "", errors.New("module directory not found")
		}

		dir = filepath.Dir(dir)
	}
	pkgRelDir, err := filepath.Rel(dir, callerDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve package dir: %w", err)
	}
	return dir, pkgRelDir, nil
}

func run() error {
	moduleDir, pkgRelDir, err := getModuleDir()
	debugf("module dir %s, package dir %s", moduleDir, pkgRelDir)

	tmpDir, err := os.MkdirTemp("", "assert.*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	if !debug {
		defer os.RemoveAll(tmpDir)
	}
	debugf("temp dir %s created", tmpDir)

	// TODO: copy _test.go files, link everything besides
	if err := os.CopyFS(tmpDir, os.DirFS(moduleDir)); err != nil {
		return fmt.Errorf("copy project to temp dir: %w", err)
	}

	testfiles := []string{}
	if err := filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}

		testfiles = append(testfiles, path)

		return nil
	}); err != nil {
		return fmt.Errorf("collect test files: %w", err)
	}

	for _, fileRelPath := range testfiles {
		ff, err := os.Open(fileRelPath)
		if err != nil {
			return fmt.Errorf("open test file %s: %w", fileRelPath, err)
		}
		stat, err := ff.Stat()
		if err != nil {
			return fmt.Errorf("stat test file %s: %w", fileRelPath, err)
		}
		f, err := io.ReadAll(ff)
		if err != nil {
			return fmt.Errorf("read test file %s: %w", fileRelPath, err)
		}
		ff.Close()

		root, err := parser.ParseFile(token.NewFileSet(), fileRelPath, f, 0)
		if err != nil {
			return fmt.Errorf("parse test file %s: %w", fileRelPath, err)
		}

		found := false
		// TODO: survive package aliasing
		// TODO: detect Assert symbol usage which is not call, refuse it
		// TODO: detect Fuse symbol usage which is not call/outside of TestMain, refuse it
		for _, decl := range root.Decls {
			// first, find Test<Name>(*testing.T) functions
			fun, ok := decl.(*ast.FuncDecl)
			if !ok || len(fun.Type.Params.List) != 1 {
				continue
			}

			argTypeStr := sprintCode(fun.Type.Params.List[0].Type)
			if argTypeStr == "*testing.M" {
				// remove pa.Fuse() call
				astutil.Apply(fun.Body, nil, func(c *astutil.Cursor) bool {
					n := c.Node()

					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}

					selector, ok := call.Fun.(*ast.SelectorExpr)
					if !ok || sprintCode(selector) != pkg.Name+".Fuse" {
						return true
					}

					found = true

					c.Replace(ast.Expr(&ast.Ident{}))

					return true
				})

				continue
			}
			if argTypeStr != "*testing.T" {
				continue
			}

			astutil.Apply(fun.Body, nil, func(c *astutil.Cursor) bool {
				n := c.Node()

				nes, ok := n.(*ast.ExprStmt)
				if !ok {
					return true
				}

				// second, find Assert(t, <predicate>) calls
				call, ok := nes.X.(*ast.CallExpr)
				if !ok {
					return true
				}

				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				var finalCall *ast.SelectorExpr
				switch sprintCode(selector) {
				case pkg.Name + ".Assert":
					finalCall = &ast.SelectorExpr{
						X:   pkgRoot,
						Sel: ast.NewIdent("ZZZAssert"),
					}
				case pkg.Name + ".Require":
					finalCall = &ast.SelectorExpr{
						X:   pkgRoot,
						Sel: ast.NewIdent("ZZZRequire"),
					}
				default:
					return true
				}

				found = true

				c.Replace(&ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{ // zzz := assert.SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED__.ZZZNew("2+2 == 5")
							Tok: token.DEFINE,
							Lhs: []ast.Expr{&ast.Ident{Name: "zzz"}},
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   pkgRoot,
										Sel: ast.NewIdent("ZZZNew"),
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: strconv.Quote(sprintCode(call.Args[1])),
										},
									},
								},
							},
						},
						&ast.ExprStmt{ // assert.ZZZAssert(t, zzz, <fused predicate expression>)
							X: &ast.CallExpr{
								Fun: finalCall,
								Args: []ast.Expr{
									ast.NewIdent("t"),
									ast.NewIdent("zzz"),
									rewriteExpr(call.Args[1], call.Args[1].Pos()),
								},
							},
						},
					},
				})

				return true
			})
		}
		if !found {
			continue
		}

		debugf("rewriting %s", fileRelPath)
		if err := os.WriteFile(fileRelPath, []byte(sprintCode(root)), stat.Mode()); err != nil {
			return fmt.Errorf("write rewritten file %s: %w", fileRelPath, err)
		}
	}

	// Re-exec only the current package instead of the whole module.
	pkgArg := "./" + filepath.ToSlash(pkgRelDir)
	if pkgRelDir == "." {
		pkgArg = "."
	}
	cmd := exec.Command("go", "test", pkgArg)
	cmd.Env = append(os.Environ(), "ASSERT_MODULE_DIR="+moduleDir)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run tests: %w", err)
	}

	return nil
}
