package q

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"runtime"
	"strings"
)

type color string

const (
	// ANSI color escape codes.
	_csiBold  color = "\033[1m"
	_csiCyan  color = "\033[36m"
	_csiReset color = "\033[0m" // "reset everything"

	_maxLineWidth = 80
)

// exprToString returns the source text underlying the given ast.Expr.
func exprToString(arg ast.Expr) string {
	var b strings.Builder
	if err := printer.Fprint(&b, token.NewFileSet(), arg); err != nil {
		return ""
	}

	// CallExpr will be multi-line and indented with tabs. replace tabs with
	// spaces so we can better control formatting during output().
	return b.String()
}

// argName returns the source text of the given argument if it's a variable or
// an expression. If the argument is something else, like a literal, argName
// returns an empty string.
func argName(arg ast.Expr) string {
	switch a := arg.(type) {
	case *ast.Ident:
		switch {
		case a.Obj == nil:
			return a.Name
		case a.Obj.Kind == ast.Var,
			a.Obj.Kind == ast.Con:
			return a.Obj.Name
		default:
			return ""
		}
	case *ast.BinaryExpr,
		*ast.CallExpr,
		*ast.IndexExpr,
		*ast.KeyValueExpr,
		*ast.ParenExpr,
		*ast.SelectorExpr,
		*ast.SliceExpr,
		*ast.TypeAssertExpr,
		*ast.UnaryExpr:
		return exprToString(arg)
	default:
		return exprToString(arg)
	}
}

// isPackage returns true if the given function call expression is in the packageName package.
func isPackage(n *ast.CallExpr, packageName string) bool {
	switch sel := n.Fun.(type) {
	case *ast.SelectorExpr: // SelectorExpr example: a.B()
		switch ident := sel.X.(type) { // sel.X is the part that precedes the .
		case *ast.Ident:
			return ident.Name == packageName
		}
	}
	return false
}

// isBareFunction returns true if the given function call expression is <funcName>().
func isBareFunction(n *ast.CallExpr, funcName string) bool {
	switch ident := n.Fun.(type) {
	case *ast.Ident:
		return ident.Name == funcName
	}
	return false
}

// isFuncCall returns true if the given function call expression is
// <funcName>() or <pkgName>.<funcName>().
func isFuncCall(n *ast.CallExpr, pkgName, funcName string) bool {
	return isBareFunction(n, funcName) || isPackage(n, pkgName)
}

// argNames finds the q.Q() call at the given filename/line number and
// returns its arguments as a slice of strings. If the argument is a literal,
// argNames will return an empty string at the index position of that argument.
// For example, q.Q(ip, port, 5432) would return []string{"ip", "port", ""}.
// argNames returns an error if the source text cannot be parsed.
func argNames(filename string, line int, pkgName, funcName string) ([]string, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, false
	}

	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		switch call := n.(type) {
		case *ast.CallExpr:
			if fset.Position(call.Pos()).Line == line && isFuncCall(call, pkgName, funcName) {
				for _, arg := range call.Args {
					names = append(names, argName(arg))
				}
			}
		}

		return true
	})
	return names, true
}

// assert.* -> Q >> runtime.Caller
const CallDepth = 2

// TODO: check not pkgName, but full package name, as it might be aliased
func Q(pkgName, funcName string) []string {
	_, file, line, ok := runtime.Caller(CallDepth)
	if !ok {
		return nil
	}

	// <pkgName>.<funcName>(foo, bar, baz) -> []string{"foo", "bar", "baz"}
	names, ok := argNames(file, line, pkgName, funcName)
	if !ok {
		return nil
	}

	return names
}
