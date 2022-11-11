package server

import (
	"context"
	"fmt"
	"net"
)

type tcp struct {
	port   string
	closed bool
}

func NewTCP(port string) *tcp {
	return &tcp{
		port:   port,
		closed: false,
	}
}

func (t *tcp) Initialize() error { return nil }
func (t *tcp) Cleanup() error    { return nil }

func (t *tcp) Execute(ctx context.Context) error {
	// Start TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", t.port))
	if err != nil {
		return err
	}

	runError := make(chan error)

	// This will close when ctx.Done() is closed
	go func() {
		// Accept all new connections
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					// This is the clean shutdown case
					runError <- nil
				default:
					// Listener received an error
					runError <- err
				}

				return
			}

			// received a new connection to handle
			go t.handleConn(conn)
		}
	}()

	// wait for shutdown or an error from the TCP server
	for {
		select {
		case err := <-runError:
			if !t.closed {
				listener.Close()
			}

			// TODO - wait for connections to drain

			// NOTE This will be nil on a graceful shutdown
			return err
		case <-ctx.Done():
			listener.Close()
		}
	}
}

func (t *tcp) handleConn(conn net.Conn) {

}
