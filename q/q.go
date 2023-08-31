package q

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"runtime"
	"strings"
	"unicode/utf8"
)

type color string

const (
	// ANSI color escape codes.
	_csiBold  color = "\033[1m"
	_csiCyan  color = "\033[36m"
	_csiReset color = "\033[0m" // "reset everything"

	_maxLineWidth = 80
)

// output writes to the log buffer. Each log message is prepended with a
// timestamp. Long lines are broken at 80 characters.
func output(args ...string) string {
	var buf bytes.Buffer

	// Subsequent lines have to be indented by the width of the timestamp.
	padding := "" // padding is the space between args.
	lineArgs := 0 // number of args printed on the current log line.
	lineWidth := 0
	for _, arg := range args {
		argWidth := argWidth(arg)
		lineWidth += argWidth + len(padding)

		// Some names in name=value strings contain newlines. Insert indentation
		// after each newline so they line up.
		arg = strings.ReplaceAll(arg, "\n", "\n")

		// Break up long lines. If this is first arg printed on the line
		// (lineArgs == 0), it makes no sense to break up the line.
		if lineWidth > _maxLineWidth && lineArgs != 0 {
			fmt.Fprint(&buf, "\n")
			lineArgs = 0
			lineWidth = argWidth
			padding = ""
		}
		fmt.Fprint(&buf, padding, arg)
		lineArgs++
		padding = " "
	}

	return buf.String()
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
		return ""
	}
}

// argNames finds the q.Q() call at the given filename/line number and
// returns its arguments as a slice of strings. If the argument is a literal,
// argNames will return an empty string at the index position of that argument.
// For example, q.Q(ip, port, 5432) would return []string{"ip", "port", ""}.
// argNames returns an error if the source text cannot be parsed.
func argNames(filename string, line int) ([]string, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, false
	}

	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, is := n.(*ast.CallExpr)
		if !is {
			// The node is not a function call.
			return true // visit next node
		}

		if fset.Position(call.End()).Line != line {
			// The node is a function call, but it's on the wrong line.
			return true
		}

		if !isQCall(call) {
			// The node is a function call on correct line, but it's not a Q()
			// function.
			return true
		}

		for _, arg := range call.Args {
			names = append(names, argName(arg))
		}

		return true
	})

	return names, true
}

// argWidth returns the number of characters that will be seen when the given
// argument is printed at the terminal.
func argWidth(arg string) int {
	// Strip zero-width characters.
	replacer := strings.NewReplacer(
		"\n", "",
		"\t", "",
		"\r", "",
		"\f", "",
		"\v", "",
		string(_csiBold), "",
		string(_csiCyan), "",
		string(_csiReset), "",
	)
	s := replacer.Replace(arg)

	return utf8.RuneCountInString(s)
}

// colorize returns the given text encapsulated in ANSI escape codes that
// give the text color in the terminal.
func colorize(text string, c color) string {
	return string(c) + text + string(_csiReset)
}

// exprToString returns the source text underlying the given ast.Expr.
func exprToString(arg ast.Expr) string {
	var b strings.Builder
	if err := printer.Fprint(&b, token.NewFileSet(), arg); err != nil {
		return ""
	}

	// CallExpr will be multi-line and indented with tabs. replace tabs with
	// spaces so we can better control formatting during output().
	return strings.ReplaceAll(b.String(), "\t", "    ")
}

// getCallerInfo returns the file, and line number of the caller
func getCallerInfo(skip int) (file string, line int, ok bool) {
	_, file, line, ok = runtime.Caller(skip)
	return file, line, ok
}

// isQCall returns true if the given function call expression is Q() or q.Q().
func isQCall(n *ast.CallExpr) bool {
	return isQFunction(n) ||
		isPackage(n, "q") ||
		isPackage(n, "fmt") ||
		isPackage(n, "a")
}

// isQFunction returns true if the given function call expression is Q().
func isQFunction(n *ast.CallExpr) bool {
	ident, is := n.Fun.(*ast.Ident)
	if !is {
		return false
	}

	return ident.Name == "Q"
}

// isPackage returns true if the given function call expression is in the q package.
// Since Q() is the only exported function from the q package, this is
// sufficient for determining that we've found Q() in the source text.
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

// Q -> getCallerInfo
const CallDepth = 2

// Q - get names of arguments from source code. Returns nil if failed to get.
func Q(v ...any) []string {
	file, line, ok := getCallerInfo(CallDepth)
	if !ok {
		return nil
	}

	// q.Q(foo, bar, baz) -> []string{"foo", "bar", "baz"}
	names, ok := argNames(file, line)
	if !ok {
		return nil
	}

	return names
}
