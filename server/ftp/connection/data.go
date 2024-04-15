package connection

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	TYPE_ASCII  DataType = "A"
	TYPE_EBCDIC DataType = "E"
	TYPE_IMAGE  DataType = "I"
	TYPE_LOCAL  DataType = "L"
)

const (
	FORMAT_NON_PRINT DataFormat = "N"
	FORMAT_TELNET    DataFormat = "T"
	FORMAT_ASA       DataFormat = "C"
)

const (
	MODE_STREAM     TransmissionMode = "S"
	MODE_BLOCK      TransmissionMode = "B"
	MODE_COMPRESSED TransmissionMode = "C"
)

type DataType string
type DataFormat string
type TransmissionMode string

type DataConnection struct {
	connection           *net.Conn
	reader               *bufio.Reader
	writer               *bufio.Writer
	isReady              bool
	newConnectionChannel chan *net.Conn
	address              net.TCPAddr
}

// OpenPassiveDataConnection starts listening for ControlConnection
// when ControlConnection is ready, send ControlConnection in channel
func OpenPassiveDataConnection() (*DataConnection, error) {
	listener, err := net.Listen("tcp", ":")
	address := *listener.Addr().(*net.TCPAddr)
	if err != nil {
		return nil, fmt.Errorf("starting listener for passive data ControlConnection: %s", err)
	}

	log.Printf("data ControlConnection listener started")
	connectionChan := make(chan *net.Conn)

	// in another thread listen to new ControlConnection
	// in case there is error in ControlConnection we can just abort current command and new ControlConnection will be used for another command
	go func() {
		for {
			conn, err := listener.Accept()

			if err != nil {
				log.Printf("Error accepting data ControlConnection: %s ", err)
			}

			log.Printf("New dtc accepted")

			connectionChan <- &conn
		}
	}()

	return &DataConnection{
		connection:           nil,
		isReady:              false,
		address:              address,
		newConnectionChannel: connectionChan,
	}, nil
}

func (dataConnection *DataConnection) FormatAddress() (string, error) {

	ipPart, portPart, _ := strings.Cut(dataConnection.address.String(), ":")

	parts := make([]string, 0)
	parts = append(parts, strings.Split(ipPart, ".")...)

	port, _ := strconv.Atoi(portPart)
	// port = p1*256+p2
	parts = append(parts, strconv.Itoa(port/256))
	parts = append(parts, strconv.Itoa(port%256))

	formatted := fmt.Sprintf("(%s)", strings.Join(parts, ","))
	return formatted, nil
}

func (dataConnection *DataConnection) Port() int {
	return dataConnection.address.Port
}

func (dataConnection *DataConnection) Close() error {

	err := (*dataConnection.connection).Close()
	if err != nil {
		return err
	}

	return nil
}

func (dataConnection *DataConnection) SendString(text string) error {
	log.Printf("Sending string down the dtc length: %d", len(text))
	_, err := dataConnection.writer.WriteString(text)
	if err != nil {
		return err
	}

	err = dataConnection.writer.Flush()
	if err != nil {
		return err
	}

	log.Printf("data send")

	return nil
}

func (dataConnection *DataConnection) Send(mode TransmissionMode, dataReader io.Reader, cancel chan bool) error {
	// TODO think about cancelation

	return nil
}

func (dataConnection *DataConnection) WaitForDataConnection() error {
	if dataConnection == nil {
		return fmt.Errorf("no data connection listener started, you need to first send EPSV or PASV")
	}

	if !dataConnection.isReady {
		// there should be timeout
		log.Printf("waiting for data connection")
		// wait until client connects to data ControlConnection
		dataConnection.connection = <-dataConnection.newConnectionChannel

		dataConnection.reader = bufio.NewReader(*dataConnection.connection)
		dataConnection.writer = bufio.NewWriter(*dataConnection.connection)

		dataConnection.isReady = true
	}

	return nil
}
