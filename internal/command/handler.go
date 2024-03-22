package command

type Command interface {
	// executes the command
	Execute() string
}

// SimpleErrorCommand return error message with given input
type SimpleErrorCommand struct {
	msg         string
	templateMsg string
}

func NewSimpleErrorCommand(msg string) *SimpleErrorCommand {
	return &SimpleErrorCommand{
		msg:         msg,
		templateMsg: "ERR unknown command",
	}
}

func (c *SimpleErrorCommand) Execute() string {
	return c.templateMsg + " - " + c.msg
}

// SimpleStringCommand return simple string
type SimpleStringCommand struct {
	msg string
}

func NewSimpleStringCommand(msg string) *SimpleStringCommand {
	return &SimpleStringCommand{
		msg: msg,
	}
}

func (c *SimpleStringCommand) Execute() string {
	return c.msg
}

// IntegerCommand return integer
type IntegerCommand struct {
	msg string
}

func NewIntegerCommand(msg string) *IntegerCommand {
	return &IntegerCommand{
		msg: msg,
	}
}

// to be deprecated: the command itself should only contains the infomatin for the command logic execution
// for example, the SET command should have `key“ and `value`, and there should be a object accessing it and interact with the cache store
func (c *IntegerCommand) Execute() string {
	return c.msg
}

// BulkStringCommand return bulk string
type BulkStringCommand struct {
	msg string
}

func NewBulkStringCommand(msg string) *BulkStringCommand {
	return &BulkStringCommand{
		msg: msg,
	}
}

func (c *BulkStringCommand) Execute() string {
	return c.msg
}

// ArrayCommand return array
type ArrayCommand struct {
	msg string
}

func NewArrayCommand(msg string) *ArrayCommand {
	return &ArrayCommand{
		msg: msg,
	}
}

func (c *ArrayCommand) Execute() string {
	return c.msg
}
