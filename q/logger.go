// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type color string

const (
	// ANSI color escape codes.
	bold     color = "\033[1m"
	yellow   color = "\033[33m"
	cyan     color = "\033[36m"
	endColor color = "\033[0m" // "reset everything"

	maxLineWidth = 80
)

// logger writes pretty logs to the $TMPDIR/q file. It takes care of opening and
// closing the file. It is safe for concurrent use.
type logger struct {
	mu  sync.Mutex   // protects all the other fields
	buf bytes.Buffer // collects writes before they're flushed to the log file
}

// output writes to the log buffer. Each log message is prepended with a
// timestamp. Long lines are broken at 80 characters.
func (l *logger) output(args ...string) string {
	l.buf.Reset()

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
		if lineWidth > maxLineWidth && lineArgs != 0 {
			fmt.Fprint(&l.buf, "\n")
			lineArgs = 0
			lineWidth = argWidth
			padding = ""
		}
		fmt.Fprint(&l.buf, padding, arg)
		lineArgs++
		padding = " "
	}

	return l.buf.String()
}

// shortFile takes an absolute file path and returns just the <directory>/<file>,
// e.g. "foo/bar.go".
func shortFile(file string) string {
	dir := filepath.Base(filepath.Dir(file))
	file = filepath.Base(file)

	return filepath.Join(dir, file)
}
