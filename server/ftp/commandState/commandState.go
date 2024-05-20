package commandState

import (
	"sync"
)

// CommandState saves information about progress of long running command (upload/download)
type CommandState struct {
	AbortChan         chan bool // allows outside process to AbortChan command
	finishedBroadcast *sync.Cond
	running           bool
	lock              *sync.Mutex
}

func New() *CommandState {
	return &CommandState{lock: &sync.Mutex{}, AbortChan: make(chan bool), finishedBroadcast: sync.NewCond(&sync.Mutex{}), running: false}
}

func (command *CommandState) Start() {
	command.lock.Lock()
	// to prevent leftover AbortChan request from canceling next command
	command.AbortChan = make(chan bool)
	command.finishedBroadcast = sync.NewCond(command.lock)
	command.running = true
	command.lock.Unlock()
}

func (command *CommandState) Finish() {
	command.lock.Lock()
	command.running = false
	command.lock.Unlock()

}

func (command *CommandState) Abort() {
	command.lock.Lock()
	command.AbortChan <- true
	command.lock.Unlock()
}

func (command *CommandState) IsRunning() bool {
	return command.running
}
