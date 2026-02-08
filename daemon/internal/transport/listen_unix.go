//go:build linux || darwin

package transport

import (
	"net"
	"os"
	"path/filepath"

	"github.com/lmriccardo/backer/deamon/internal/core"
)

func getUnixSockPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, core.UNIX_SOCK_NAME)
	}
	return filepath.Join(os.TempDir(), core.UNIX_SOCK_NAME)
}

func listen() (net.Listener, string, TransportType, error) {
	sock_path := getUnixSockPath()
	_ = os.Remove(sock_path) // Remove stale socket if present

	// Start listening on the unix socket
	ln, err := net.Listen("unix", sock_path)
	if err != nil {
		return nil, "", TransportUnix, err
	}

	// Restrict access to the sock file ( 0600 usually )
	if err := os.Chmod(sock_path, 0600); err != nil {
		_ = ln.Close()
		return nil, "", TransportUnix, err
	}

	return ln, sock_path, TransportUnix, nil
}
