package command

type CommandHandler interface {
	Execute(args []string) string
}
