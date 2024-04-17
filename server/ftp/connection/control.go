package connection

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type ControlConnection struct {
	rawConnection *net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
}

func NewConnection(rawConnection *net.Conn) *ControlConnection {
	reader := bufio.NewReader(*rawConnection)
	writer := bufio.NewWriter(*rawConnection)

	conn := &ControlConnection{
		rawConnection: rawConnection,
		reader:        reader,
		writer:        writer,
	}

	return conn
}

func (conn *ControlConnection) ReceiveLine() (string, error) {

	line, _, err := conn.reader.ReadLine()

	if err != nil {

		if err == io.EOF {
			return "", fmt.Errorf("ControlConnection closed (EOF)")
		}
		return "", fmt.Errorf("reading line from ControlConnection: %s", err)

	}

	//log.Printf("ReceiveLine: %s", line)

	return string(line), nil
}

func (conn *ControlConnection) SendString(msg string) error {
	_, err := conn.writer.WriteString(msg)

	if err != nil {
		return fmt.Errorf("writing line to ControlConnection: %s", err)
	}

	// this can be optimized
	err = conn.writer.Flush()
	if err != nil {
		return fmt.Errorf("flushing data to ControlConnection: %s", err)
	}

	//log.Printf("%d byres written: %s", n, msg)

	return nil
}

func (conn *ControlConnection) Close() error {
	if conn == nil {
		log.Printf("Tried to close connection that was nul")
		return nil
	}

	err := (*conn.rawConnection).Close()
	if err != nil {
		return fmt.Errorf("error closing ControlConnection: %s", err)
	}

	return nil
}
