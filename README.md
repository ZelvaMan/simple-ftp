# simple-ftp

## Server

First goal is to implement simple server that can be connected to using any ftp client and displays
hardcoded list of files.

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

- Block mode
  - no spacing between blocks
  - every block starts with 3 byte header
    - 1. descriptor (16 = Restart maker, 64 = EOF)
    - 2-3. **BYTE** count
- Compressed
  - optional

#### Restart and recovery

- data sender inserts marker with arbritary data (only meaningfull for sender)
- receiver can then use latest marker as poin from which to restart connection
- marker can be cursor possition in fiel
- only use printable characters in marker (because its send in command)

### Commands

- `USER`
  - usually first command to be transmitted by client
- `PASS`
  - should follow immediatly after `USER`
- `LOGOUT`
- `CWD [Path]`
- `CDUP` Change to parent directory
  - special case of `CWD`
- `PASV`
  - request server to start listening for DTC
- `TYPE [Type] [Optional format]`
  - changes data type
  - default is ASCII non print
- `MODE [Mode]`
  - default is stream 
  - Modes:
    - S = Stream
    - B = Block
    - C = Compressed

### Replies

- each command generates at least one response
- response starts with 3 digit number followed by text (separated by space)

#### Response codes

First digit:

- `1XY` Positive Preliminary reply
  - command was accepted, but no complete
  - at most one per command

- `2XY` Positive Completion reply
  - command completed succesfully

- `3XY` Positive Intermediate reply
  - waiting for another command
  - used in command sequences

- `4XY` Transient Negative Completion reply

- `5XY` 

Second digit:


## Knowledge sources

- [rfc 959](https://datatracker.ietf.org/doc/html/rfc959)
- [LIST Response format](https://stackoverflow.com/questions/4564603/format-of-the-data-returned-by-the-ftp-list-command)
- [LIST response format to use](https://stackoverflow.com/questions/2443007/ftp-list-format)