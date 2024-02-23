# simple-ftp

## Notes

### Connections

FTP uses 2 connections for communication between clien and server.
Both are full duplex.

First one called `control connection` and is used for sending command.
It is opened by client. It uses the `telnet` protocol.
Client is responsible for requesting the closure of connection.

Other one is called `data connection` and can be either oppened by server (`active`) or by client(`pasive`).
`pasive` is usually used for circumventing firewall restrictions.
It is used for transfering contents of files.

### Creating a connection

Client opens the controll connection to server.


## Knowledge sources

- [rfc 959](https://datatracker.ietf.org/doc/html/rfc959)
