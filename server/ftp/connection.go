package ftp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type connection struct {
	rawConnection *net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	isReady       bool
}

type dataConnection struct {
	connection           *connection
	isReady              bool
	newConnectionChannel chan *connection
	address              string
}

func newConnection(rawConnection *net.Conn) *connection {
	reader := bufio.NewReader(*rawConnection)
	writer := bufio.NewWriter(*rawConnection)

	conn := &connection{
		rawConnection: rawConnection,
		reader:        reader,
		writer:        writer,
		isReady:       true,
	}

	return conn
}

// starts listening for connection
// when connection is ready, send connection in channel
func openPassiveDataConnection() (*dataConnection, error) {
	listener, err := net.Listen("tcp", ":")
	address := listener.Addr().String()
	if err != nil {
		return nil, fmt.Errorf("starting listener for passive data connection: %s", err)
	}

	connectionChan := make(chan *connection)

	// in another thread listen to new connection
	// in case there is error in connection we can just abort current command and new connection will be used for another command
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting data connection: %s ", err)
			}

			connectionChan <- newConnection(&conn)
		}
	}()

	return &dataConnection{
		connection:           nil,
		isReady:              false,
		address:              address,
		newConnectionChannel: connectionChan,
	}, nil
}

func (dataConnection *dataConnection) getDataConnection() *connection {
	if dataConnection == nil {
		// no pasv called
		// TODO handle error
	}

	if !dataConnection.isReady {
		// there should be timeout
		dataConnection.connection = <-dataConnection.newConnectionChannel
		dataConnection.isReady = true
	}

	return dataConnection.connection
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
	n, err := conn.writer.WriteString(msg + "\n")

	if err != nil {
		return fmt.Errorf("writing line to connection: %s", err)
	}

	// this can be optimized
	err = conn.writer.Flush()
	if err != nil {
		return fmt.Errorf("writing flushing connection: %s", err)
	}

	log.Printf("%d byres written", n)

	return nil
}

func (conn *connection) close() error {
	err := (*conn.rawConnection).Close()
	if err != nil {
		return fmt.Errorf("Error closing connection: %s", err)
	}

	return nil
}

func (dataConn *dataConnection) FormatAddress() (string, error) {

	ipPart, portPart, _ := strings.Cut(dataConn.address, ":")

	parts := make([]string, 0)
	parts = append(parts, strings.Split(ipPart, ".")...)

	port, _ := strconv.Atoi(portPart)
	// port = p1*256+p2
	parts = append(parts, strconv.Itoa(port/256))
	parts = append(parts, strconv.Itoa(port%256))

	formatted := fmt.Sprintf("(%s)", strings.Join(parts, ","))
	return formatted, nil
}

func (dataConnection *dataConnection) close() {
	// todo handle
}
