package transport

import (
	"fmt"
	"net"
	"os"
)

type TransportType int

const (
	TransportWindows TransportType = iota
	TransportUnix
)

type Transport struct {
	SocketPath string        // The path to the unix socket or windows named pipe
	Listener   net.Listener  // The actual listener
	Type       TransportType // Kind an enum that describes the transport type
}

func NewTransport() (*Transport, error) {
	ln, addr, _type, err := listen() // Create the listener OS-based
	if err != nil {
		return nil, err
	}

	t := &Transport{SocketPath: addr, Listener: ln, Type: _type}
	os.Setenv("BACKER_HOST", t.Uri())

	return t, nil
}

func (t *Transport) Uri() string {
	switch t.Type {
	case TransportUnix:
		return fmt.Sprintf("unix://%s", t.SocketPath)
	case TransportWindows:
		return fmt.Sprintf("npipe://%s", t.SocketPath)
	default:
		return ""
	}
}

func (t *Transport) Close() {
	// Stop accepting connections
	if t.Listener != nil {
		_ = t.Listener.Close()
	}

	// On UNIX, unlink the socket *after* closing listener
	if t.Type == TransportUnix && t.SocketPath != "" {
		_ = os.Remove(t.SocketPath)
	}

	// This only affects THIS process (not your shell)
	os.Unsetenv("BACKER_HOST")
}
