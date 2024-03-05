package command

type CommandRegistry struct {
	Commands map[string]CommandHandler
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		Commands: make(map[string]CommandHandler),
	}
}

func (r *CommandRegistry) RegisterCommand(commandName string, handler CommandHandler) {
	r.Commands[commandName] = handler
}

func (r *CommandRegistry) GetCommandHandler(commandName string) CommandHandler {
	return r.Commands[commandName]
}
