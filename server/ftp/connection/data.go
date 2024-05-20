package connection

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
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

const CHUNK_SIZE = 1024

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
	if err != nil {
		return nil, fmt.Errorf("starting listener for passive data ControlConnection: %s", err)
	}

	log.Printf("data ControlConnection listener started")
	connectionChan := make(chan *net.Conn)

	address := *listener.Addr().(*net.TCPAddr)

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

func (dataConnection *DataConnection) FormatAddressForPASV() (string, error) {

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
	if dataConnection == nil {
		return nil
	}

	var err error

	if dataConnection.isReady {
		err = (*dataConnection.connection).Close()
	}
	dataConnection.isReady = false
	dataConnection.reader = nil
	dataConnection.writer = nil
	dataConnection.connection = nil

	if err != nil {
		return err
	}

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

		// using buffered reader and writer for performance
		dataConnection.reader = bufio.NewReader(*dataConnection.connection)
		dataConnection.writer = bufio.NewWriter(*dataConnection.connection)
		dataConnection.isReady = true
	}

	return nil
}

func (dataConnection *DataConnection) Send(mode TransmissionMode, dataReader io.Reader, cancelChannel chan bool) error {
	// ensure that data connection exists and is ready
	err := dataConnection.WaitForDataConnection()
	if err != nil {
		return fmt.Errorf("waiting for data connection: %s", err)
	}

	// TODO think about cancelation
	switch mode {
	case MODE_STREAM:
		return dataConnection.sendStreamData(dataReader, cancelChannel)
	}

	return fmt.Errorf("unsupported mode")
}

func (dataConnection *DataConnection) Receive(mode TransmissionMode, dataWriter io.Writer) error {
	log.Printf("waiting for data connection to receive data from client")

	// ensure that data connection exists and is ready
	err := dataConnection.WaitForDataConnection()
	if err != nil {
		return fmt.Errorf("waiting for data connection: %s", err)
	}

	switch mode {
	case MODE_STREAM:

		log.Printf("start receiving data form client in stream mode")
		// TODO handle timeout
		// TODO handle file size limit
		_, err := dataConnection.reader.WriteTo(dataWriter)

		if err != nil {
			return fmt.Errorf("copying data from socket to file: %s", err)
		}

		err = dataConnection.Close()
		if err != nil {
			return fmt.Errorf("closing DTC after finished transfer: %s", err)
		}
	}

	return nil
}

func (dataConnection *DataConnection) sendStreamData(dataReader io.Reader, cancel chan bool) error {

	log.Printf("start sending dataReader")
	chunk := 0

	// send dataReader in chunks
	for {
		select {
		case <-cancel:
			// maybe better way is to just abort the connection
			return fmt.Errorf("cancelation requested")
		default:
		}

		// TODO use read directly to prevent blocking
		// this works ok, if transmission is slow, if its blocked, abort will not be detected
		// maybe just set write deadline of connection async, that way we can also just use copy
		// when we capture we then check if cancel request was send and if so we assume it is the reason for error?
		writtenSize, err := io.CopyN(dataConnection.writer, dataReader, CHUNK_SIZE)
		chunk += 1

		//log.Printf("Written chunk %d size: %d",
		//	chunk, writtenSize)

		// TODO remove
		time.Sleep(10 * time.Millisecond)

		// finished sending dataReader
		if errors.Is(err, io.EOF) || writtenSize < CHUNK_SIZE {
			log.Printf("EOF reached, transfer is ok")
			break
		}

		if err != nil {
			return fmt.Errorf("copy dataReader from filereader to socker: %s", err)
		}

		err = dataConnection.writer.Flush()

		if err != nil {
			return fmt.Errorf("flushing DTC after copy: %s", err)
		}

	}

	log.Printf("finished sending dataReader")

	err := dataConnection.Close()

	if err != nil {
		return fmt.Errorf("closing DTC after finished transfer: %s", err)
	}

	return nil
}
