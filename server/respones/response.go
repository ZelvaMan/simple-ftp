package respones

import (
	"fmt"
	"log"
	"strings"
)

func formatResponse(responseCode int, message string) string {
	return fmt.Sprintf("%d %s", responseCode, message)
}

func UserLoggedIn() string {
	return formatResponse(230, "User logged in, proceed.")
}

func PasswordNeeded() string {
	return formatResponse(331, "User name okay, need password.")
}

func BadSequence() string {
	return formatResponse(503, "Bad sequence of commands.")
}
func Ready() string {
	return formatResponse(220, "zmftp ready for new user.")
}

func NotLoggedIn() string {
	return formatResponse(530, "Not logged in / incorrect password.")
}

func NotImplemented() string {
	return formatResponse(502, "Command not implemented.")
}

func ServerSystem() string {
	return formatResponse(215, "Zelvaman ultimate server")
}

func ListFeatures(features []string) string {
	var builder strings.Builder

	builder.WriteString("211- Features: \n")

	for _, feat := range features {
		log.Printf("enabled feat: %s", feat)

		builder.WriteString(" " + feat + "\n")
	}

	builder.WriteString("211 end")
	return builder.String()
}

func EPSVEnabled(portNumber int) string {
	message := fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", portNumber)
	return formatResponse(229, message)
}

func ListOk() string {
	return formatResponse(226, "Directory listing ok")
}

func ListSendingResponse() string {
	return formatResponse(150, "Here comes the directory listing")
}

func NotAllowed() string {
	return formatResponse(553, "Requested action not taken.")
}

func SendPWD(path string) string {
	msg := fmt.Sprintf(" \"%s\" Returning working director", path)
	return formatResponse(257, msg)

}

func TypeChanged() string {
	return formatResponse(200, "Data type changed")
}

func FileActionOk() string {
	return formatResponse(250, "Requested file action okay, completed.")
}
