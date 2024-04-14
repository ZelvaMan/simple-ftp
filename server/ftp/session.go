package ftp

import (
	"log"
	"net"
	"server/respones"
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

type DataType string
type DataFormat string

type SessionInfo struct {
	controlConnection *connection
	dataConnection    *dataConnection
	cwd               string
	isLoggedIn        bool
	username          string
	commandSequence   string
	dataType          DataType
	dataFormat        DataFormat
}

func createSession(controlConnection *net.Conn) (*SessionInfo, error) {

	session := &SessionInfo{
		controlConnection: newConnection(controlConnection),
		dataConnection:    nil,
		cwd:               "/",
		isLoggedIn:        false,
		username:          "",
		commandSequence:   "",
		dataType:          TYPE_ASCII,
		dataFormat:        FORMAT_NON_PRINT,
	}

	return session, nil

}

func (session *SessionInfo) Start() {
	log.Printf("session is starting...")

	err := session.Respond(respones.Ready())
	if err != nil {
		log.Printf("Error sending hello msg: %s", err)
	}

	for {
		line, err := session.controlConnection.readLine()

		if err != nil {
			log.Printf("Error reading line from control connection: %s", err)

			// TODO let server know that session was closed
			break

		}

		log.Printf("line received from control '%s'", line)

		// maybe handle if not response have been send
		err = session.handleCommand(line)

		if err != nil {
			log.Printf("handling command")
			break
		}

	}

	log.Printf("Closing connection")

	// close the connection
	_ = session.controlConnection.close()
	session.dataConnection.close()
}

// Abort about session is case of server shutdown
func (session *SessionInfo) Abort() {

	// TODO send abort message
}

// Respond send response on control connection. Adds newline.
func (session *SessionInfo) Respond(message string) error {
	log.Printf("Server response: %s", message)
	return session.controlConnection.write(message + "\r\n")
}
