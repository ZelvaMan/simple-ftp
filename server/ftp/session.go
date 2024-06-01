package ftp

import (
	"log"
	"net"
	"server/fs"
	"server/fs/mapedfs"
	"server/ftp/commandState"
	"server/ftp/connection"
	"server/respones"
	"server/sequences"
)

type SessionInfo struct {
	controlConnection *connection.ControlConnection
	dataConnection    *connection.DataConnection
	cwd               string
	isLoggedIn        bool
	username          string
	commandSequence   sequences.SequenceInfo
	dataType          connection.DataType
	dataFormat        connection.DataFormat
	transmissionMode  connection.TransmissionMode
	filesystem        fs.Filesystem
	command           *commandState.CommandState
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
		commandSequence:   nil,
		dataType:          connection.TYPE_ASCII,
		dataFormat:        connection.FORMAT_NON_PRINT,
		transmissionMode:  connection.MODE_STREAM,
		filesystem:        filesystem,
		command:           commandState.New(),
	}

	return session, nil

}

func (session *SessionInfo) Start() {
	defer func() {
		log.Printf("Defer function start")
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
			log.Printf("error while handling command")
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
	log.Printf("Server response: %s", message)
	return session.controlConnection.SendString(message + "\r\n")

}

// RespondOrPanic wrapper around Respond that panics on error, because it is unrecoverable error
func (session *SessionInfo) RespondOrPanic(message string) {

	err := session.Respond(message)
	if err != nil {
		log.Printf("PANIC: error while responding to client: %s", err)
		panic(err)
	}
}
