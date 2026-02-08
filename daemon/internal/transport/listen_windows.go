//go:build windows

package transport

import (
	"errors"
	"net"
)

func listen() (net.Listener, string, TransportType, error) {
	return nil, "", TransportWindows, errors.New("Windows version not yet implemented")
}
