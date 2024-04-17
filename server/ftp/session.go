package ftp

import (
	"log"
	"net"
	"server/fs"
	"server/fs/mapedfs"
	"server/ftp/connection"
	"server/respones"
)

type SessionInfo struct {
	controlConnection *connection.ControlConnection
	dataConnection    *connection.DataConnection
	cwd               string
	isLoggedIn        bool
	username          string
	commandSequence   string
	dataType          connection.DataType
	dataFormat        connection.DataFormat
	transmissionMode  connection.TransmissionMode
	filesystem        fs.Filesystem
}

func createSession(controlConnection *net.Conn) (*SessionInfo, error) {
	// create fs
	filesystem, err := mapedfs.CreateFS("/home/jrada/git/simple-ftp/test-fs")
	if err != nil {
		return nil, err
	}

	session := &SessionInfo{
		controlConnection: connection.NewConnection(controlConnection),
		dataConnection:    nil,
		cwd:               "/",
		isLoggedIn:        false,
		username:          "",
		commandSequence:   "",
		dataType:          connection.TYPE_ASCII,
		dataFormat:        connection.FORMAT_NON_PRINT,
		transmissionMode:  connection.MODE_STREAM,
		filesystem:        filesystem,
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
		line, err := session.controlConnection.ReceiveLine()

		if err != nil {
			log.Printf("Error reading line from control controlConnection: %s", err)

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

	log.Printf("Closing controlConnection")

	// close the controlConnection
	_ = session.controlConnection.Close()

	_ = session.dataConnection.Close()
}

// Abort about session is case of server shutdown
func (session *SessionInfo) Abort() {

	// TODO send abort message
}

// Respond send response on control controlConnection. Adds newline.
func (session *SessionInfo) Respond(message string) error {
	log.Printf("Server response: %s", message)
	return session.controlConnection.SendString(message + "\r\n")
}
