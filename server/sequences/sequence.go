package sequences

const PASSWORD = "PASS"
const NONE = "NONE"
const RenameTo = "RENAME_TO"

type SequenceInfo interface {
	NextPhase() string
}

type LoginSequence struct {
	Username string
}

func (seq *LoginSequence) NextPhase() string {
	return PASSWORD
}

func NewLoginSequence(username string) *LoginSequence {
	return &LoginSequence{Username: username}
}

type RenameSequence struct {
	RenameFromPath string
}

func (seq *RenameSequence) NextPhase() string {
	return RenameTo
}

func NewRenameSequence(renameFromPath string) *RenameSequence {
	return &RenameSequence{RenameFromPath: renameFromPath}
}
