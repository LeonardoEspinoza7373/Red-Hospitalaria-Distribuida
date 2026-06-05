package transport

import (
	"fmt"
	"net"
	"time"

	"github.com/LeonardoEspinoza7373/Red-Hospitalaria-Distribuida/internal/protocol"
)

const dialTimeout = 2 * time.Second

func SendMessage(addr string, msg protocol.Message) error {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	data, err := msg.Encode()
	if err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("write to %s: %w", addr, err)
	}

	return nil
}
