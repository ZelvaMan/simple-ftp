package ftp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type connection struct {
	rawConnection *net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
}

func newConnection(rawConnection *net.Conn) *connection {
	reader := bufio.NewReader(*rawConnection)
	writer := bufio.NewWriter(*rawConnection)

	conn := &connection{
		rawConnection: rawConnection,
		reader:        reader,
		writer:        writer,
	}

	return conn
}

func (conn *connection) readLine() (string, error) {
	line, _, err := conn.reader.ReadLine()

	if err != nil {

		if err == io.EOF {
			return "", fmt.Errorf("connection closed (EOF)")
		}
		return "", fmt.Errorf("reading line from connection: %s", err)

	}

	log.Printf("readLine: %s", line)

	return string(line), nil
}

func (conn *connection) writeLine(msg string) error {
	_, err := conn.writer.WriteString(msg + "\n")
	if err != nil {
		return fmt.Errorf("writing line to connection: %s", err)
	}

	return nil
}
