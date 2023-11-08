package command

type SlashCommand struct {
	Prefix      string
	Description string
	Execute     func(content string)
}

type CommandDispatcher struct {
	prefixes         []string
	prefixCommandMap map[string]SlashCommand
}
