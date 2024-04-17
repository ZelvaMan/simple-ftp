package ftp

import (
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
	default:
		log.Printf("Command %s is not implemented", command)

		session.RespondOrPanic(respones.NotImplemented())
	}

	if err != nil {
		return fmt.Errorf("handling command: %s", err)
	}

	return nil
}

func (session *SessionInfo) handleUSER(username string) error {
	session.username = username

	session.RespondOrPanic(respones.PasswordNeeded())

	session.commandSequence = sequences.PASSWORD
	return nil
}

func (session *SessionInfo) handlePASS(password string) error {
	// check sequence
	if session.commandSequence != sequences.PASSWORD {
		log.Printf("wrong sequence")

		session.RespondOrPanic(respones.BadSequence())
	}

	log.Printf("trying to authenticate user %s", session.username)

	// wrong password/username
	if !authenticateUser(session.username, password) {
		log.Printf("Wrong user name or pasword")

		session.RespondOrPanic(respones.NotLoggedIn())

		return nil
	}

	log.Printf("user authenticated")

	session.RespondOrPanic(respones.UserLoggedIn())

	// login ok
	session.isLoggedIn = true
	session.commandSequence = sequences.NONE

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
	session.RespondOrPanic(respones.ListOk())
	return nil
}

func (session *SessionInfo) handleSYST() error {
	session.RespondOrPanic(respones.ServerSystem())

	return nil
}

func (session *SessionInfo) handleFEAT() error {
	features := []string{}

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
	joinedPath := filepath.Join(session.cwd, requestedPath)

	fileReader, err := session.filesystem.Retrieve(joinedPath)
	if err != nil {
		return fmt.Errorf("retrieving file from fs: %s", err)
	}
	log.Printf("filereader retrieved, sending file...")
	session.RespondOrPanic(respones.SendingResponse())

	err = session.dataConnection.Send(session.transmissionMode, fileReader, nil)
	if err != nil {
		return fmt.Errorf("sending file: %s", err)
	}

	session.RespondOrPanic(respones.DataSendClosingConnection())

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
