// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"fmt"
	"strings"
	"testing"
)

// TestOutput verifies that logger.output() prints the expected output to the
// log buffer.
func TestOutput(t *testing.T) {
	testCases := []struct {
		args []string
		want string
	}{
		{
			args: []string{fmt.Sprintf("%s=%s", colorize("a", bold), colorize("int(1)", cyan))},
			want: fmt.Sprintf("%s %s=%s\n", colorize("0.000s", yellow), colorize("a", bold), colorize("int(1)", cyan)),
		},
	}

	for _, tc := range testCases {
		l := logger{}
		l.output(tc.args...)

		got := l.buf.String()
		if got != tc.want {
			argString := strings.Join(tc.args, ", ")
			t.Fatalf("\nlogger.output(%s)\ngot:  %swant: %s", argString, got, tc.want)
		}
	}
}
