package ftp

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"server/ftp/connection"
	"server/respones"
	"server/sequences"
	"slices"
	"strings"
)

var publicCommands = []string{"USER", "PASS"}

// handleCommand returned error means, that session is in irrecoverable state and we have to close it
func (session *SessionInfo) handleCommand(commandLine string) error {
	log.Printf("Received command '%s'", commandLine)

	command, argument, hasArguments := strings.Cut(commandLine, " ")
	if !hasArguments {
		command = commandLine
	}

	// only allow some commands
	if !session.isLoggedIn && !slices.Contains(publicCommands, command) {
		session.RespondOrPanic(respones.NotLoggedIn())

		return nil
	}

	if session.command.IsRunning() {
		log.Printf("trying to handle command while another is running, %s", commandLine)

		switch command {
		case "ABOR":

			_ = session.handleABOR()

		default:
			// TODO use better error

			// better way would be to place command in some queue to be processed later
			session.RespondOrPanic(respones.BadSequence())
		}

		return nil
	}

	var err error
	switch command {
	case "USER":
		err = session.handleUSER(argument)
	case "PASS":
		err = session.handlePASS(argument)
	case "LIST":
		err = session.handleLIST(argument)
	case "SYST":
		err = session.handleSYST()
	case "FEAT":
		err = session.handleFEAT()
	case "PWD":
		err = session.handlePWD()
	case "CWD":
		err = session.handleCWD(argument)
	case "TYPE":
		err = session.handleTYPE(argument)
	case "MODE":
		err = session.handleMODE(argument)
	case "RETR":
		err = session.handleRETR(argument)
	case "EPSV":
		err = session.handleEPSV()
	case "PASV":
		err = session.handlePASV()
	case "STOR":
		err = session.handleSTOR(argument)
	case "QUIT":
		err = session.handleQUIT()
	case "ABOR":
		err = session.handleABOR()
	case "RNFR":
		err = session.handleRNFR(argument)
	case "RNTO":
		err = session.handleRNTO(argument)
	default:
		log.Printf("Command %s is not implemented", command)

		session.RespondOrPanic(respones.NotImplemented())
	}

	// TODO create custom error struct to separate non recoverable errors

	// unrecoverable error while processing command
	if err != nil {
		return fmt.Errorf("unrecoverable error while handling command: %s", err)
	}

	return nil
}

func (session *SessionInfo) handleUSER(username string) error {

	session.RespondOrPanic(respones.PasswordNeeded())

	session.commandSequence = sequences.NewLoginSequence(username)
	return nil
}

func (session *SessionInfo) handlePASS(password string) error {
	loginSequence, ok := session.commandSequence.(*sequences.LoginSequence)

	// check sequence
	if !ok {
		log.Printf("wrong command sequence")

		session.RespondOrPanic(respones.BadSequence())
	}

	log.Printf("trying to authenticate user %s", loginSequence.Username)

	// wrong password/username
	if !authenticateUser(loginSequence.Username, password) {
		log.Printf("Wrong user name or pasword")

		session.RespondOrPanic(respones.NotLoggedIn())

		return nil
	}

	log.Printf("user authenticated")

	session.RespondOrPanic(respones.UserLoggedIn())

	// login ok
	session.username = loginSequence.Username
	session.isLoggedIn = true
	session.commandSequence = nil

	return nil
}

func (session *SessionInfo) handleLIST(requestedPath string) error {
	// if no path is specified, use cwd
	joinedPath := filepath.Join(session.cwd, requestedPath)

	files, err := session.filesystem.List(joinedPath)
	if err != nil {

	}

	printListReader := strings.NewReader(files.String())

	// notify client that we will stand sending response
	session.RespondOrPanic(respones.SendingResponse())

	// send data using data connection
	err = session.dataConnection.Send(session.transmissionMode, printListReader, nil)
	if err != nil {
		return err
	}
	log.Printf("data written to data controlConnection")

	// acknowledge that all data was send
	session.RespondOrPanic(respones.FileActionOk())
	return nil
}

func (session *SessionInfo) handleSYST() error {
	session.RespondOrPanic(respones.System())

	return nil
}

func (session *SessionInfo) handleFEAT() error {
	features := []string{"SIZE"}

	session.RespondOrPanic(respones.ListFeatures(features))

	return nil
}

func (session *SessionInfo) handlePWD() error {
	session.RespondOrPanic(respones.SendPWD(session.cwd))

	log.Printf("returned working directory")
	return nil
}

func (session *SessionInfo) handleTYPE(params string) error {
	dataType, dataFormat, hasFormatSet := strings.Cut(params, " ")

	switch dataType {
	case "A":
		session.dataType = connection.TYPE_ASCII
	case "E":
		session.dataType = connection.TYPE_EBCDIC
	case "I":
		session.dataType = connection.TYPE_IMAGE
	case "L":
		// TODO return some unsupported error
		session.dataType = connection.TYPE_LOCAL
	default:
		return errors.New("data type not supported")
	}

	if hasFormatSet && (session.dataType == connection.TYPE_ASCII || session.dataType == connection.TYPE_EBCDIC) {
		switch dataFormat {
		case "N":
			session.dataFormat = connection.FORMAT_NON_PRINT
		case "T":
			session.dataFormat = connection.FORMAT_TELNET
		case "C":
			session.dataFormat = connection.FORMAT_ASA
		default:
			return errors.New("data format not supported")
		}
	}

	session.RespondOrPanic(respones.CommandOkay())

	return nil
}

func (session *SessionInfo) handleCWD(argument string) error {
	// TODO do some validation
	session.cwd = argument

	log.Printf("CWD changed to %s", session.cwd)

	session.RespondOrPanic(respones.FileActionOk())

	return nil
}

func (session *SessionInfo) handleMODE(argument string) error {
	switch argument {
	case "S":
		session.transmissionMode = connection.MODE_STREAM
	case "B":
		session.transmissionMode = connection.MODE_BLOCK
	case "C":
		session.transmissionMode = connection.MODE_COMPRESSED
	default:
		return fmt.Errorf("")
	}

	session.RespondOrPanic(respones.CommandOkay())

	return nil
}

func (session *SessionInfo) handleRETR(requestedPath string) error {
	// TODO handle resume

	joinedPath := filepath.Join(session.cwd, requestedPath)

	// if command would not close, session would be locked until abort is issued
	session.command.Start()

	go func() {

		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: runing goroutine for sending file: %s", err)
			}
			log.Printf("async task defer")
			// TODO maybe abort whole session
			session.command.Finish()
		}()

		fileReader, err := session.filesystem.Retrieve(joinedPath)
		if err != nil {
			log.Printf("Error getting reader for file: %s", err)
			session.RespondOrPanic(respones.FileUnavailable(requestedPath))
			return
		}

		log.Printf("filereader retrieved, sending file...")

		session.RespondOrPanic(respones.SendingResponse())

		err = session.dataConnection.Send(session.transmissionMode, fileReader, session.command.AbortChan)
		if err != nil {
			log.Printf("Error sending file: %s", err)
			session.RespondOrPanic(respones.GenericError())
			return
		}

		session.RespondOrPanic(respones.DataSendClosingConnection())

	}()

	return nil
}

func (session *SessionInfo) handlePASV() error {
	log.Printf("passive controlConnection requested")
	dataConn, err := connection.OpenPassiveDataConnection()
	if err != nil {
		return fmt.Errorf("error opening data controlConnection: %s", err)
	}
	// listener started
	session.dataConnection = dataConn

	return nil
}

func (session *SessionInfo) handleEPSV() error {
	log.Printf("Extended passive mode requested")
	dataConn, err := connection.OpenPassiveDataConnection()
	if err != nil {
		return fmt.Errorf("error opening data controlConnection: %s", err)
	}
	// listener started
	session.dataConnection = dataConn

	log.Printf("Data conneciton listener started")
	// send port to listened on
	session.RespondOrPanic(respones.EPSVEnabled(dataConn.Port()))

	return nil
}

func (session *SessionInfo) handleSTOR(destination string) error {
	session.RespondOrPanic(respones.StartUpload())

	// TODO save to temp file
	uploadBuffer := new(bytes.Buffer)

	log.Printf("start receiving data...")

	err := session.dataConnection.Receive(session.transmissionMode, uploadBuffer)
	if err != nil {
		log.Printf("Error processing:  %s", err)
		session.RespondOrPanic(respones.TransferAborted())
	}

	log.Printf("data received")

	err = session.filesystem.Store(destination, uploadBuffer)
	if err != nil {
		log.Printf("Error storing file: %s", err)

		session.RespondOrPanic(respones.GenericError())
	}

	log.Printf("File saved to fs succesfully")
	session.RespondOrPanic(respones.FileActionOk())
	return nil
}

func (session *SessionInfo) handleQUIT() error {
	session.RespondOrPanic(respones.ClosingControlConnection())

	return nil
}

func (session *SessionInfo) handleABOR() error {
	log.Printf("ABOR command received")

	if session.command.IsRunning() {
		session.command.Abort()

		session.RespondOrPanic(respones.TransferAborted())
	}

	session.RespondOrPanic(respones.DataSendClosingConnection())

	return nil
}

func (session *SessionInfo) handleRNFR(renameFromPath string) error {

	exists, err := session.filesystem.Exists(renameFromPath)
	if err != nil {
		log.Printf("fs exists error: %s", err)
		session.RespondOrPanic(respones.GenericError())
	}

	// validate path exists
	if !exists {
		session.RespondOrPanic(respones.FileUnavailable(renameFromPath))
		return nil
	}

	session.commandSequence = sequences.NewRenameSequence(renameFromPath)

	session.RespondOrPanic(respones.PendingFurtherAction("rnto"))

	return nil
}

func (session *SessionInfo) handleRNTO(renameToPath string) error {
	renameSequence, ok := session.commandSequence.(*sequences.RenameSequence)

	// check sequence
	if !ok {
		log.Printf("wrong command sequence")

		session.RespondOrPanic(respones.BadSequence())
	}

	err := session.filesystem.Rename(renameSequence.RenameFromPath, renameToPath)
	if err != nil {
		log.Printf("Error renaming file: %s", err)

		session.RespondOrPanic(respones.GenericError())

		return nil
	}

	session.RespondOrPanic(respones.FileActionOk())

	session.commandSequence = nil

	return nil
}
