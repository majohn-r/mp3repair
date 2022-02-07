package subcommands

type CommandProcessor interface {
	Name() string
	Exec([]string)
}