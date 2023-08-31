package q

import (
	"fmt"
	"go/ast"
	"testing"

	a "github.com/stretchr/testify/assert"
)

func TestOutput(t *testing.T) {
	for _, testcase := range []struct {
		args string
		want string
	}{
		{
			args: fmt.Sprintf(
				"%s=%s",
				colorize("a", _csiBold),
				colorize("int(1)", _csiCyan),
			),
			want: fmt.Sprintf(
				"%s=%s",
				colorize("a", _csiBold),
				colorize("int(1)", _csiCyan),
			),
		},
	} {
		a.Equal(t, testcase.want, output(testcase.args))
	}
}

// TestExtractingArgsFromSourceText verifies that exprToString() and argName()
// arg able to extract the text of the arguments passed to q.Q().
// For example, q.Q(myVar) should return "myVar".
func TestExtractingArgsFromSourceText(t *testing.T) {
	for _, test := range []struct {
		arg  ast.Expr
		want string
	}{
		{
			arg:  &ast.Ident{NamePos: 123, Obj: ast.NewObj(ast.Var, "myVar")},
			want: "myVar",
		},
		{
			arg:  &ast.Ident{NamePos: 234, Obj: ast.NewObj(ast.Var, "awesomeVar")},
			want: "awesomeVar",
		},
		{
			arg:  &ast.Ident{NamePos: 456, Obj: ast.NewObj(ast.Bad, "myVar")},
			want: "",
		},
		{
			arg:  &ast.Ident{NamePos: 789, Obj: ast.NewObj(ast.Con, "myVar")},
			want: "myVar",
		},
		{
			arg: &ast.BinaryExpr{
				X:     &ast.BasicLit{ValuePos: 49, Kind: 5, Value: "1"},
				OpPos: 51,
				Op:    12,
				Y:     &ast.BasicLit{ValuePos: 53, Kind: 5, Value: "2"},
			},
			want: "1 + 2",
		},
		{
			arg: &ast.BinaryExpr{
				X:     &ast.BasicLit{ValuePos: 89, Kind: 6, Value: "3.14"},
				OpPos: 94,
				Op:    15,
				Y:     &ast.BasicLit{ValuePos: 96, Kind: 6, Value: "1.59"},
			},
			want: "3.14 / 1.59",
		},
		{
			arg: &ast.BinaryExpr{
				X:     &ast.BasicLit{ValuePos: 73, Kind: 5, Value: "123"},
				OpPos: 77,
				Op:    14,
				Y:     &ast.BasicLit{ValuePos: 79, Kind: 5, Value: "234"},
			},
			want: "123 * 234",
		},
		{
			arg: &ast.CallExpr{
				Fun: &ast.Ident{
					NamePos: 30,
					Name:    "foo",
					Obj: &ast.Object{
						Kind: 5,
						Name: "foo",
						Decl: &ast.FuncDecl{
							Doc:  nil,
							Recv: nil,
							Name: &ast.Ident{
								NamePos: 44,
								Name:    "foo",
								Obj:     &ast.Object{},
							},
							Type: &ast.FuncType{
								Func: 39,
								Params: &ast.FieldList{
									Opening: 47,
									List:    nil,
									Closing: 48,
								},
								Results: &ast.FieldList{
									Opening: 0,
									List: []*ast.Field{
										{
											Doc:   nil,
											Names: nil,
											Type: &ast.Ident{
												NamePos: 50,
												Name:    "int",
												Obj:     nil,
											},
											Tag:     nil,
											Comment: nil,
										},
									},
									Closing: 0,
								},
							},
							Body: &ast.BlockStmt{
								Lbrace: 54,
								List: []ast.Stmt{
									&ast.ReturnStmt{
										Return: 57,
										Results: []ast.Expr{
											&ast.BasicLit{ValuePos: 64, Kind: 5, Value: "123"},
										},
									},
								},
								Rbrace: 68,
							},
						},
						Data: nil,
						Type: nil,
					},
				},
				Lparen:   33,
				Args:     nil,
				Ellipsis: 0,
				Rparen:   34,
			},
			want: "foo()",
		},
		{
			arg: &ast.IndexExpr{
				X: &ast.Ident{
					NamePos: 51,
					Name:    "a",
					Obj: &ast.Object{
						Kind: 4,
						Name: "a",
						Decl: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									NamePos: 30,
									Name:    "a",
									Obj:     &ast.Object{},
								},
							},
							TokPos: 32,
							Tok:    47,
							Rhs: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.ArrayType{
										Lbrack: 35,
										Len:    nil,
										Elt: &ast.Ident{
											NamePos: 37,
											Name:    "int",
											Obj:     nil,
										},
									},
									Lbrace: 40,
									Elts: []ast.Expr{
										&ast.BasicLit{ValuePos: 41, Kind: 5, Value: "1"},
										&ast.BasicLit{ValuePos: 44, Kind: 5, Value: "2"},
										&ast.BasicLit{ValuePos: 47, Kind: 5, Value: "3"},
									},
									Rbrace: 48,
								},
							},
						},
						Data: nil,
						Type: nil,
					},
				},
				Lbrack: 52,
				Index:  &ast.BasicLit{ValuePos: 53, Kind: 5, Value: "1"},
				Rbrack: 54,
			},
			want: "a[1]",
		},
		{
			arg: &ast.KeyValueExpr{
				Key: &ast.Ident{
					NamePos: 72,
					Name:    "Greeting",
					Obj:     nil,
				},
				Colon: 80,
				Value: &ast.BasicLit{ValuePos: 82, Kind: 9, Value: "\"Hello\""},
			},
			want: `Greeting: "Hello"`,
		},
		{
			arg: &ast.ParenExpr{
				Lparen: 35,
				X: &ast.BinaryExpr{
					X:     &ast.BasicLit{ValuePos: 36, Kind: 5, Value: "2"},
					OpPos: 38,
					Op:    14,
					Y:     &ast.BasicLit{ValuePos: 40, Kind: 5, Value: "3"},
				},
				Rparen: 41,
			},
			want: "(2 * 3)",
		},
		{
			arg: &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: 44,
					Name:    "fmt",
					Obj:     nil,
				},
				Sel: &ast.Ident{
					NamePos: 48,
					Name:    "Print",
					Obj:     nil,
				},
			},
			want: "fmt.Print",
		},
		{
			arg: &ast.SliceExpr{
				X: &ast.Ident{
					NamePos: 51,
					Name:    "a",
					Obj: &ast.Object{
						Kind: 4,
						Name: "a",
						Decl: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									NamePos: 30,
									Name:    "a",
									Obj:     &ast.Object{},
								},
							},
							TokPos: 32,
							Tok:    47,
							Rhs: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.ArrayType{
										Lbrack: 35,
										Len:    nil,
										Elt: &ast.Ident{
											NamePos: 37,
											Name:    "int",
											Obj:     (*ast.Object)(nil),
										},
									},
									Lbrace: 40,
									Elts: []ast.Expr{
										&ast.BasicLit{ValuePos: 41, Kind: 5, Value: "1"},
										&ast.BasicLit{ValuePos: 44, Kind: 5, Value: "2"},
										&ast.BasicLit{ValuePos: 47, Kind: 5, Value: "3"},
									},
									Rbrace: 48,
								},
							},
						},
						Data: nil,
						Type: nil,
					},
				},
				Lbrack: 52,
				Low:    &ast.BasicLit{ValuePos: 53, Kind: 5, Value: "0"},
				High:   &ast.BasicLit{ValuePos: 55, Kind: 5, Value: "2"},
				Max:    nil,
				Slice3: false,
				Rbrack: 56,
			},
			want: "a[0:2]",
		},
		{
			arg: &ast.TypeAssertExpr{
				X: &ast.Ident{
					NamePos: 62,
					Name:    "a",
					Obj: &ast.Object{
						Kind: 4,
						Name: "a",
						Decl: &ast.ValueSpec{
							Doc: nil,
							Names: []*ast.Ident{
								{
									NamePos: 34,
									Name:    "a",
									Obj:     &ast.Object{},
								},
							},
							Type: &ast.InterfaceType{
								Interface: 36,
								Methods: &ast.FieldList{
									Opening: 45,
									List:    nil,
									Closing: 46,
								},
								Incomplete: false,
							},
							Values:  nil,
							Comment: nil,
						},
						Data: int(0),
						Type: nil,
					},
				},
				Lparen: 64,
				Type: &ast.Ident{
					NamePos: 65,
					Name:    "string",
					Obj:     nil,
				},
				Rparen: 71,
			},
			want: "a.(string)",
		},
		{
			arg: &ast.UnaryExpr{
				OpPos: 35,
				Op:    13,
				X:     &ast.BasicLit{ValuePos: 36, Kind: 5, Value: "1"},
			},
			want: "-1",
		},
		{
			arg: &ast.Ident{
				NamePos: 65,
				Name:    "string",
				Obj:     nil,
			},
			want: "string",
		},
	} {
		t.Run(fmt.Sprintf("exprToString(%T)", test.arg), func(t *testing.T) {
			if _, ok := test.arg.(*ast.Ident); ok {
				return
			}

			a.Equal(t, test.want, exprToString(test.arg))
		})

		t.Run(fmt.Sprintf("argName(%T)", test.arg), func(t *testing.T) {
			a.Equal(t, test.want, argName(test.arg))
		})
	}
}

// TestArgNames verifies that argNames() is able to find the q.Q() call in the
// sample text and extract the argument names. For example, if q.q(a, b, c) is
// in the sample text, argNames() should return []string{"a", "b", "c"}.
func TestArgNames(t *testing.T) {
	const filename = "../cmd/main.go"
	want := []string{"a", "b", "c", "d", "e", "f", "g"}
	got, ok := argNames(filename, 18)
	a.True(t, ok)
	a.Equal(t, want, got)
}

func TestArgNamesBadFilename(t *testing.T) {
	_, ok := argNames("BAD FILENAME", 666)
	a.False(t, ok)
}

// TestArgWidth verifies that argWidth() returns the correct number of printable
// characters in a string.
func TestArgWidth(t *testing.T) {
	for _, test := range []struct {
		arg       string
		wantWidth int
	}{
		{colorize("myVar", _csiCyan), 5},
		{colorize(`"myStringLiteral"`, _csiCyan), 17},
		{colorize("func (n int) { return n > 0 }(1)", _csiCyan), 32},
		{colorize("myVar", _csiBold), 5},
		{colorize("3.14", _csiCyan), 4},
		{colorize("你好", _csiCyan), 2},
	} {
		a.Equal(t, test.wantWidth, argWidth(test.arg))
	}
}

// TestFormatArgs verifies that formatArgs() produces the expected string.
func TestFormatArgs(t *testing.T) {
	for id, test := range map[int]struct {
		args []any
		want []string
	}{
		1: {
			args: []any{123},
			want: []string{colorize("int(123)", _csiCyan)},
		},
		2: {
			args: []any{123, 3.14, "hello world"},
			want: []string{
				colorize("int(123)", _csiCyan),
				colorize("float64(3.14)", _csiCyan),
				colorize("hello world", _csiCyan),
			},
		},
		3: {
			args: []any{[]string{"goodbye", "world"}},
			want: []string{
				colorize(`[]string{"goodbye", "world"}`, _csiCyan),
			},
		},
		4: {
			args: []any{
				[]struct{ a, b int }{
					{1, 2}, {2, 3}, {3, 4},
				},
			},
			want: []string{
				colorize(`[]struct { a int; b int }{
    {a:1, b:2},
    {a:2, b:3},
    {a:3, b:4},
}`, _csiCyan),
			},
		},
	} {
		t.Run(fmt.Sprintf("TEST %d", id), func(t *testing.T) {
			a.Equal(t, test.want, formatArgs(test.args...))
		})
	}
}

// TestPrependArgName verifies that prependArgName() correctly merges a slice of
// variable names and a slice of variabe values into name=value strings.
func TestPrependArgName(t *testing.T) {
	for _, test := range []struct {
		names  []string
		values []string
		want   []string
	}{
		{
			names:  []string{"myVar"},
			values: []string{colorize("int(100)", _csiCyan)},
			want:   []string{fmt.Sprintf("%s=%s", colorize("myVar", _csiBold), colorize("int(100)", _csiCyan))},
		},
		{
			names:  []string{"", "myFloat"},
			values: []string{colorize("hello", _csiCyan), colorize("float64(3.14)", _csiCyan)},
			want: []string{
				colorize("hello", _csiCyan),
				fmt.Sprintf("%s=%s", colorize("myFloat", _csiBold), colorize("float64(3.14)", _csiCyan)),
			},
		},
		{
			names: []string{"myStructSlice", "", "myFunc"},
			values: []string{
				colorize("[]*Foo{&Foo{123, 234}, &Foo{345, 456}}", _csiCyan),
				colorize("int(-666)", _csiCyan),
				colorize("func (n int) bool { return n > 0 }", _csiCyan),
			},
			want: []string{
				fmt.Sprintf("%s=%s", colorize("myStructSlice", _csiBold), colorize("[]*Foo{&Foo{123, 234}, &Foo{345, 456}}", _csiCyan)),
				colorize("int(-666)", _csiCyan),
				fmt.Sprintf("%s=%s", colorize("myFunc", _csiBold), colorize("func (n int) bool { return n > 0 }", _csiCyan)),
			},
		},
	} {
		a.Equal(t, test.want, prependArgName(test.names, test.values))
	}
}

func TestIsQCall(t *testing.T) {
	for id, test := range map[int]struct {
		expr *ast.CallExpr
		want bool
	}{
		1: {
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "Q"},
			},
			want: true,
		},
		2: {
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "R"},
			},
			want: false,
		},
		3: {
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{Name: "q"},
				},
			},
			want: true,
		},
		4: {
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{Name: "Q"},
				},
			},
			want: false,
		},
		5: {
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.BadExpr{},
				},
			},
			want: false,
		},
		6: {
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "q"},
			},
			want: false,
		},
	} {
		t.Run(fmt.Sprintf("TEST %d", id), func(t *testing.T) {
			a.Equal(t, test.want, isQCall(test.expr))
		})
	}
}
