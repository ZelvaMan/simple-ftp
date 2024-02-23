# simple-ftp

## Server



## Notes

### SHORTS

EOF = Enf of file
FTP = FIle transfer protocol
DTC = Data transfer connection
DTP = Data transfer port

### Connections

FTP uses 2 connections for communication between clien and server.
Both are full duplex.

First one called `control connection` and is used for sending command.
It is opened by client. It uses the `telnet` protocol.
Client is responsible for requesting the closure of connection.

Other one is called `data connection` and can be either oppened by server (`active`) or by client(`pasive`).
`pasive` is usually used for circumventing firewall restrictions.
It is used for transfering contents of files.
Client has to listen on DTP before sending transfer command.
When server receives the transfer command it opens the connection
and sends confirming reply. Server is responsible for managing connection.
 `PORT` command is used to specify DTP and it forces reconnecting of DTC.

### Data Transfer

client specifies the type. Default file structure is `FILE`.

#### Data types

- ASCI = mandatory, uses `8-bit NVT-ASCII`
- Image = raw bytes

#### Tranmission modes

- Stream
  - stupid
  - requires restart of DTC after each file transmitted
  - no way to detect if connection is closed by accident

##### Stream

If used with file structure you have to close connection to mark EOF

## Knowledge sources

- [rfc 959](https://datatracker.ietf.org/doc/html/rfc959)
