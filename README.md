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
It is opened by client.
Server immediatly responds with status 220.
It uses the `telnet` protocol.
Client is responsible for requesting the closure of connection.

Other one is called `data connection` and can be either oppened by server (`active`) or by client(`pasive`).
`pasive` is usually used for circumventing firewall restrictions.
It is used for transfering contents of files.
Client has to listen on DTP before sending transfer command.
When server receives the transfer command it opens the connection
and sends confirming reply. Server is responsible for managing connection.
 `PORT` command is used to specify DTP and it forces reconnecting of DTC.

Controll connection is opened from client to port `21`.
After `PASV` command part of response is port to which client should connect.
Server maybe returns unique port for each connected client?

**How does server pair DTC A control connection**?

- one way it to make each DTP unique
  - doesnt work behind nat
  - simplest to implement 
- other is to assume 1 client per ip addr

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
- `LIST [Directory]`
  - format is not specified (user ls)

### Replies

- each command generates at least one response
- response starts with 3 digit number followed by text (separated by space)

#### Response codes

First digit:

- `1xy` Positive Preliminary reply
  - command was accepted, but no complete
  - at most one per command

- `2xy` Positive Completion reply
  - command completed succesfully

- `3xy` Positive Intermediate reply
  - waiting for another command
  - used in command sequences

- `4xy` Transient Negative Completion reply
  - non-pernament -> retrie later

- `5xy` Permanent Negative Completion reply
  - pernament -> dontr retry

Second digit:
- `x0z` Syntax
  - also for not implemented stuff
- `x1z` Information
  - replies to requests for informations
- `x2z` Connections
- `x3z` Authentication and accounting
  - used in auth flow
- `x5z` File system
  - stuff like file not found etc.

### Minimum implementation

```text
TYPE - ASCII Non-print
         MODE - Stream
         STRUCTURE - File, Record
         COMMANDS - USER, QUIT, PORT,
                    TYPE, MODE, STRU,
                      for the default values
                    RETR, STOR,
                    NOOP.
```



## Knowledge sources

- [rfc 959](https://datatracker.ietf.org/doc/html/rfc959)
- [LIST Response format](https://stackoverflow.com/questions/4564603/format-of-the-data-returned-by-the-ftp-list-command)
- [LIST response format to use](https://stackoverflow.com/questions/2443007/ftp-list-format)