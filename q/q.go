// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

// nolint: gochecknoglobals
var (
	// std is the singleton logger.
	std logger

	// CallDepth allows setting the number of levels runtime.Caller will
	// skip when looking up the caller of the q.Q function. This allows
	// the `q` package to be wrapped by a project-specific wrapping function,
	// which would increase the depth by at least one. It's better to not
	// include calls to `q.Q` in released code at all and scrub them before,
	// a build is created, but in some cases it might be useful to provide
	// builds that do include the additional debug output provided by `q.Q`.
	// This also allows the consumer of the package to control what happens
	// with leftover `q.Q` calls. Defaults to 2, because the user code calls
	// q.Q(), which calls getCallerInfo().
	CallDepth = 2
)

// Q pretty-prints the given arguments
func Q(v ...any) string {
	std.mu.Lock()
	defer std.mu.Unlock()

	args := formatArgs(v...)
	_, file, line, err := getCallerInfo()
	if err != nil {
		return std.output(args...) // no name=value printing
	}

	// q.Q(foo, bar, baz) -> []string{"foo", "bar", "baz"}
	names, err := argNames(file, line)
	if err != nil {
		return std.output(args...) // no name=value printing
	}

	// Convert the arguments to name=value strings.
	args = prependArgName(names, args)
	return std.output(args...)
}
