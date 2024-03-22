package command

type CommandRegistry struct {
	Commands map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		Commands: make(map[string]Command),
	}
}

func (r *CommandRegistry) RegisterCommand(commandName string, handler Command) {
	r.Commands[commandName] = handler
}

func (r *CommandRegistry) GetCommandHandler(commandName string) Command {
	return r.Commands[commandName]
}
