package fics

import (
	"bufio"
	"net"
)

type SoughtGame struct {
}

type TelnetClient struct {
	conn            net.Conn
	MessageReceiver chan string
	MessageSender   chan string
}

func NewTelnetClient(url string) (*TelnetClient, error) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		return nil, err
	}

	recv := make(chan string)
	send := make(chan string)

	go readRoutine(conn, recv)
	go writeRoutine(send, conn)

	return &TelnetClient{
		conn:            conn,
		MessageReceiver: recv,
		MessageSender:   send,
	}, nil
}

func (c *TelnetClient) Close() {
	c.conn.Close()
}

func readRoutine(src net.Conn, dst chan string) {
	defer close(dst)

	reader := bufio.NewReader(src)
	msg := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return // Connection closed or EOF
		}
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			msg += line
			dst <- msg
			msg = ""
		} else {
			msg += line
		}
	}
}

func writeRoutine(src chan string, dst net.Conn) {
	for msg := range src {
		msg = msg + "\r\n"

		total_written := 0
		for total_written < len(msg) {
			n, err := dst.Write([]byte(msg[total_written:]))
			if err != nil || n == 0 {
				return // Connection closed or EOF
			}
			total_written += n
		}
	}
}
