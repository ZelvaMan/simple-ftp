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
	defer func() {
		// captures irrecoverable error, like send
		if err := recover(); err != nil {
			log.Println("PANIC: panic occurred in session :", err)
		}

		// ensure the connection are closed
		_ = session.controlConnection.Close()
		_ = session.dataConnection.Close()
	}()

	log.Printf("session is starting...")

	session.RespondOrPanic(respones.ServerReady())

	for {
		line, err := session.controlConnection.ReceiveLine()

		if err != nil {
			log.Printf("Error reading line from control controlConnection: %s", err)
			break

		}

		// maybe handle if not response have been send
		err = session.handleCommand(line)

		if err != nil {
			log.Printf("handling command")
			break
		}

	}
}

// Abort about session is case of server shutdown
func (session *SessionInfo) Abort() {

	// TODO send abort message
}

// Respond send response on control controlConnection. Adds newline.
func (session *SessionInfo) Respond(message string) error {
	// TODO maybe panic in case of error.
	// error in sending data in unrecoverable condition, that will force termination of session.
	// I am not sure if this is okay practice

	log.Printf("Server response: %s", message)
	return session.controlConnection.SendString(message + "\r\n")

}

// RespondOrPanic wrapper around Respond that panics on error
func (session *SessionInfo) RespondOrPanic(message string) {

	err := session.Respond(message)
	if err != nil {
		panic(err)
	}
}
