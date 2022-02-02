package subcommands

type SubCommand interface {
	Name() string
	Exec([]string)
}