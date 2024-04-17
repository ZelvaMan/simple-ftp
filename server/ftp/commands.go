package ftp

import (
	"errors"
	"fmt"
	"log"
	"server/ftp/connection"
	"server/respones"
	"server/sequences"
	"slices"
	"strings"
)

var publicCommands = []string{"USER", "PASS"}

func (session *SessionInfo) handleCommand(commandLine string) error {

	command, argument, hasArguments := strings.Cut(commandLine, " ")
	if !hasArguments {
		command = commandLine
	}

	// only allow some commands
	if !session.isLoggedIn && !slices.Contains(publicCommands, command) {
		err := session.Respond(respones.NotLoggedIn())
		if err != nil {
			return fmt.Errorf("sending not logged in response: %s", err)
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
		log.Printf("Extended passive mode requested")
		dataConn, err := connection.OpenPassiveDataConnection()
		if err != nil {
			return fmt.Errorf("error opening data controlConnection: %s", err)
		}
		// listener started
		session.dataConnection = dataConn

		log.Printf("Data conneciton listener started, gonna send response")
		// send port to listened on
		err = session.Respond(respones.EPSVEnabled(dataConn.Port()))
	case "PASV":
		log.Printf("passive controlConnection requested")
		dataConn, err := connection.OpenPassiveDataConnection()
		if err != nil {
			return fmt.Errorf("error opening data controlConnection: %s", err)
		}
		// listener started
		session.dataConnection = dataConn
	default:
		log.Printf("Command %s is not implemented", command)

		err = session.Respond(respones.NotImplemented())
	}

	if err != nil {
		return fmt.Errorf("handling command: %s", err)
	}

	return nil
}

func (session *SessionInfo) handleUSER(username string) error {
	session.username = username

	err := session.Respond(respones.PasswordNeeded())
	if err != nil {
		return respondError("user", err)

	}

	session.commandSequence = sequences.PASSWORD
	return nil
}

func (session *SessionInfo) handlePASS(password string) error {
	// check sequence
	if session.commandSequence != sequences.PASSWORD {
		log.Printf("wrong sequence")
		err := session.Respond(respones.BadSequence())

		if err != nil {
			return respondError("pass", err)
		}
	}

	log.Printf("trying to authenticate user %s", session.username)

	// wrong password/username
	if !authenticateUser(session.username, password) {
		log.Printf("Wrong user name or pasword")

		err := session.Respond(respones.NotLoggedIn())
		if err != nil {
			return respondError("pass", err)
		}

		return nil
	}

	log.Printf("user authenticated")

	err := session.Respond(respones.UserLoggedIn())
	if err != nil {
		return respondError("pass", err)

	}

	// login ok
	session.isLoggedIn = true
	session.commandSequence = sequences.NONE

	return nil
}

func (session *SessionInfo) handleLIST(requestedPath string) error {
	// if no path is specified, use cwd
	if requestedPath == "" {
		requestedPath = session.cwd
	}
	files, err := session.filesystem.List(requestedPath)

	var builder strings.Builder

	for _, file := range files {
		builder.WriteString(fmt.Sprintf("%s\r\n", file.String()))
	}

	printListReader := strings.NewReader(builder.String())

	// notify client that we will stand sending response
	err = session.Respond(respones.SendingResponse())
	if err != nil {
		return respondError("list", err)

	}

	// send data using data connection
	err = session.dataConnection.Send(session.transmissionMode, printListReader, nil)
	if err != nil {
		return err
	}
	log.Printf("data written to data controlConnection")

	// acknowledge that all data was send
	err = session.Respond(respones.ListOk())
	return err
}

func (session *SessionInfo) handleSYST() error {
	err := session.Respond(respones.ServerSystem())
	if err != nil {
		return respondError("syst", err)
	}

	return nil
}

func (session *SessionInfo) handleFEAT() error {
	features := []string{}

	err := session.Respond(respones.ListFeatures(features))

	if err != nil {
		return respondError("feat", err)
	}

	return nil
}

func (session *SessionInfo) handlePWD() error {
	err := session.Respond(respones.SendPWD(session.cwd))
	if err != nil {
		return respondError("pwd", err)
	}

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

	err := session.Respond(respones.CommandOkay())

	if err != nil {
		return respondError("type", err)
	}

	return nil
}

func (session *SessionInfo) handleCWD(argument string) error {
	// TODO do some validation
	session.cwd = argument

	log.Printf("CWD changed to %s", session.cwd)

	err := session.Respond(respones.FileActionOk())
	if err != nil {
		return respondError("cwd", err)
	}

	return nil
}

func (session *SessionInfo) handleMODE(argument string) error {
	switch argument {
	case "S":
		session.transmissionMode = connection.MODE_STREAM
	case "B":
		session.transmissionMode = connection.MODE_STREAM
	case "C":
		session.transmissionMode = connection.MODE_COMPRESSED
	default:
		return fmt.Errorf("")
	}

	err := session.Respond(respones.CommandOkay())
	if err != nil {
		return respondError("mode", err)
	}

	return nil
}

func (session *SessionInfo) handleRETR(argumnet string) error {

	fileReader, err := session.filesystem.Retrieve(argumnet)
	if err != nil {
		return fmt.Errorf("retrieving file from fs: %s", err)
	}
	log.Printf("filereader retrieved, sending file...")
	err = session.Respond(respones.SendingResponse())
	if err != nil {
		return respondError("retr", err)
	}

	err = session.dataConnection.Send(session.transmissionMode, fileReader, nil)
	if err != nil {
		return fmt.Errorf("sending file: %s", err)
	}

	err = session.Respond(respones.DataSendClosingConnection())
	if err != nil {
		return respondError("retr", err)
	}

	return nil
}
