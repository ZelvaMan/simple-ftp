package connection

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
)

const TELNET_IAC = 255 // interrupt as command
const TELNET_IP = 244
const TELNET_DATA_MARK = 242

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

	return dataToCommand(line), nil
}

var bytesToSkip = []byte{TELNET_IAC, TELNET_IP, TELNET_DATA_MARK}

func dataToCommand(data []byte) string {
	dataStartIdx := 0

	// it is just raw data
	if len(data) >= 2 && data[0] == TELNET_IAC {

		// TELNET Command has to be at least 2 bytes long
		dataStartIdx = 2

		// skip all control bytes
		// maybe only skip if first byte is IAC?
		if dataStartIdx < len(bytesToSkip) && slices.Contains(bytesToSkip, data[dataStartIdx]) {
			dataStartIdx++
		}

		log.Printf("command line has telnet control sequence, skipping %d bytes", dataStartIdx)
	}

	return string(data[dataStartIdx:])
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
