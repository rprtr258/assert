package assert

import (
	"cmp"
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

func Use[E any](t T, f func() (E, error)) E {
	t.Helper()
	res, err := f()
	NoError(t, err)
	return res
}

func UseJSON[E any](t T, data []byte) E {
	t.Helper()
	var res E
	NoError(t, json.Unmarshal(data, &res))
	return res
}

func UseTempDir(t T) string {
	t.Helper()
	res := Use(t, func() (string, error) {
		return os.MkdirTemp(os.TempDir(), "")
	})
	t.Cleanup(func() {
		NoError(t, os.RemoveAll(res))
	})
	return res
}

// FileConfig file parameters to use. If file already exists, it will be used.
// Temporary file is created otherwise.
type FileConfig struct {
	// Filename to create file with. If not provided, random name is generated.
	Filename string
	// Dir to create file in. If not provided, temp dir is used.
	Dir string
	// Content to write to file. If not provided, empty file is created.
	Content []byte
	// Mode to create file with. If not provided, 0o644 is used.
	// If file already exists, it will be checked having this mode.
	Mode os.FileMode
}

func UseFile(t T, cfg FileConfig) *os.File {
	t.Helper()

	cfg.Dir = cmp.Or(cfg.Dir, os.TempDir())
	cfg.Mode = cmp.Or(cfg.Mode, 0o644)

	var file *os.File
	filename := filepath.Join(cfg.Dir, cfg.Filename)
	if stat, err := os.Stat(filename); err == nil {
		// file already exists, use it
		if stat.IsDir() {
			t.Fatalf("file %q is a directory", filename)
		}
		if stat.Mode() != cfg.Mode {
			t.Fatalf("file %q has mode %o, expected %o", filename, stat.Mode(), cfg.Mode)
		}

		file, err = os.OpenFile(filename, os.O_RDWR, cfg.Mode)
		NoError(t, err)
	} else {
		// file does not exist, create it
		NoError(t, os.MkdirAll(cfg.Dir, 0o755))

		const tries = 10000
		for try := 0; try < tries; try++ {
			name := filename + strconv.Itoa(rand.Int())
			file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, cfg.Mode)
			if err == nil {
				break
			} else if !os.IsExist(err) {
				t.Fatalf("failed to create file %q", name)
			} else if try == tries-1 {
				t.Fatalf("file %q already exists", name)
			}
		}
		t.Cleanup(func() {
			NoError(t, file.Close())
		})
	}

	if cfg.Content != nil {
		_, err := file.Write(cfg.Content)
		NoError(t, err)

		_, err = file.Seek(0, io.SeekStart)
		NoError(t, err)
	}

	return file
}

func UseReadAll(t T, r io.Reader) []byte {
	t.Helper()
	return Use(t, func() ([]byte, error) {
		content, err := io.ReadAll(r)
		return content, err
	})
}

func UseFileContent(t T, filename string) []byte {
	t.Helper()
	file := UseFile(t, FileConfig{Filename: filename})
	return UseReadAll(t, file)
}

// NOTE: envs are set for whole processes, so all parallel tests will see them
func UseEnv(tb testing.TB, key, value string) {
	tb.Helper()
	// cleanup is done inside
	tb.Setenv(key, value)
}

func UsePanic(t T, f func()) (res any) {
	t.Helper()
	defer func() {
		if res = recover(); res == nil {
			t.Fatalf("no panic")
		}
	}()
	f()
	return
}

func UseTcpPort(t T, address string) int {
	t.Helper()
	tt := Wrap(t)

	net.Listen("tcp", ":0")
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(address), Port: 0})
	NoError(tt.Msgf("listen tcp %q:0", address), err)
	defer l.Close()

	f, err := l.File()
	NoError(tt.Msg("open socket file"), err)

	err = syscall.SetsockoptLinger(int(f.Fd()), syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger{Onoff: 0, Linger: 0})
	NoError(tt.Msg("set linger option"), err)

	_, portStr, err := net.SplitHostPort(l.Addr().String())
	NoError(tt.Msgf("split host port: %q", l.Addr().String()), err)

	port, err := strconv.Atoi(portStr)
	NoError(tt.Msgf("parse port: %q", portStr), err)
	return port
}
