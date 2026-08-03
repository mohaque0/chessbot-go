package fics

import (
	"io"
	"net"
	"strings"
	"time"
)

type TelnetClient struct {
	conn    net.Conn
	Recv    chan string
	Send    chan string
	debugWr io.Writer
}

func NewTelnetClient(url string, debugWr io.Writer) (*TelnetClient, error) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		return nil, err
	}

	recv := make(chan string)
	send := make(chan string)

	go readRoutine(conn, recv, debugWr)
	go writeRoutine(send, conn)

	return &TelnetClient{
		conn:    conn,
		Recv:    recv,
		Send:    send,
		debugWr: debugWr,
	}, nil
}

func (c *TelnetClient) Close() {
	c.conn.Close()
}

func readRoutine(src net.Conn, dst chan string, debugWr io.Writer) {
	defer close(dst)

	buf := make([]byte, 4096)
	var accum strings.Builder

	for {
		src.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, err := src.Read(buf)
		if n > 0 {
			cleaned := stripTelnetIAC(buf[:n])
			if debugWr != nil {
				debugWr.Write(cleaned)
			}
			accum.Write(cleaned)
			// Flush complete lines. FICS uses \n\r (not standard \r\n),
			// so split on \n and strip any \r.
			for {
				content := accum.String()
				idx := strings.IndexByte(content, '\n')
				if idx < 0 {
					break
				}
				line := strings.ReplaceAll(content[:idx], "\r", "")
				rest := content[idx+1:]
				rest = strings.TrimLeft(rest, "\r")
				dst <- line
				accum.Reset()
				accum.WriteString(rest)
			}
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout — flush any partial data (handles prompts
				// like "login: " that don't end with \n).
				if accum.Len() > 0 {
					dst <- strings.ReplaceAll(accum.String(), "\r", "")
					accum.Reset()
				}
				continue
			}
			if accum.Len() > 0 {
				dst <- accum.String()
			}
			return
		}
	}
}

// stripTelnetIAC removes telnet IAC (0xFF) command sequences from raw data.
// IAC sequences are 2-3 bytes: IAC + command [+ option].
func stripTelnetIAC(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == 0xFF && i+1 < len(data) {
			cmd := data[i+1]
			if cmd == 0xFF {
				out = append(out, 0xFF)
				i++
			} else if cmd >= 0xFB && cmd <= 0xFE {
				// WILL/WONT/DO/DONT — 3-byte sequence
				i += 2
			} else {
				// Other 2-byte commands (e.g. IAC GA)
				i++
			}
		} else {
			out = append(out, data[i])
		}
	}
	return out
}

func writeRoutine(src chan string, dst net.Conn) {
	for msg := range src {
		msg = msg + "\r\n"

		total_written := 0
		for total_written < len(msg) {
			n, err := dst.Write([]byte(msg[total_written:]))
			if err != nil || n == 0 {
				return
			}
			total_written += n
		}
	}
}
