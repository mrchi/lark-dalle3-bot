package dispatcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mrchi/lark-dalle3-bot/pkg/larkee"
)

type Command struct {
	Prefix      string
	Description string
	Execute     func(prompt string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string)
}

type CommandDispatcher struct {
	larkeeClient       *larkee.LarkClient
	prefixes           []string
	prefixCommandMap   map[string]Command
	commandHelpExecute func(helpMessage string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string)
}

func NewCommandDispatcher(larkeeClient *larkee.LarkClient, commands ...Command) *CommandDispatcher {
	dispatcher := CommandDispatcher{larkeeClient: larkeeClient}
	for _, command := range commands {
		dispatcher.prefixes = append(dispatcher.prefixes, command.Prefix)
		dispatcher.prefixCommandMap[command.Prefix] = command
	}
	// 按照前缀长度逆序排序，避免出现 /a /ab /abc 时，/ab 会被 /a 匹配的情况
	sort.SliceStable(dispatcher.prefixes, func(i, j int) bool {
		return len(dispatcher.prefixes[i]) > len(dispatcher.prefixes[j])
	})
	return &dispatcher
}

func (dispatcher *CommandDispatcher) GenMarkdownHelpMsg() string {
	var helpMessage []string
	for _, prefix := range dispatcher.prefixes {
		command := dispatcher.prefixCommandMap[prefix]
		helpMessage = append(helpMessage, fmt.Sprintf("**%s** %s", command.Prefix, command.Description))
	}
	return strings.Join(helpMessage, "\n")
}

func (dispatcher *CommandDispatcher) Dispatch(text string, messageId string, tanantKey string) {
	for _, prefix := range dispatcher.prefixes {
		if strings.HasPrefix(text, prefix) {
			prompt := strings.TrimSpace(strings.TrimPrefix(text, prefix))
			go dispatcher.prefixCommandMap[prefix].Execute(prompt, dispatcher.larkeeClient, messageId, tanantKey)
			return
		}
	}
	go dispatcher.commandHelpExecute(dispatcher.GenMarkdownHelpMsg(), dispatcher.larkeeClient, messageId, tanantKey)
}
