package dispatcher

import (
	"sort"
	"strings"

	"github.com/mrchi/lark-dalle3-bot/pkg/larkee"
)

type Command struct {
	Prefix  string
	HelpMsg string
	Execute func(prompt string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string)
}

type CommandDispatcher struct {
	larkeeClient       *larkee.LarkClient
	prefixes           []string
	helpMsgs           []string
	prefixCommandMap   map[string]Command
	commandHelpExecute func(helpMsgs []string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string)
}

func NewCommandDispatcher(
	larkeeClient *larkee.LarkClient,
	commandHelpExecute func(helpMsgs []string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string),
	commands ...Command,
) *CommandDispatcher {
	dispatcher := CommandDispatcher{larkeeClient: larkeeClient, commandHelpExecute: commandHelpExecute, prefixCommandMap: make(map[string]Command)}
	for _, command := range commands {
		dispatcher.prefixes = append(dispatcher.prefixes, command.Prefix)
		dispatcher.helpMsgs = append(dispatcher.helpMsgs, command.HelpMsg)
		dispatcher.prefixCommandMap[command.Prefix] = command
	}
	// 按照前缀长度逆序排序，避免出现 /a /ab /abc 时，/ab 会被 /a 匹配的情况
	sort.SliceStable(dispatcher.prefixes, func(i, j int) bool {
		return len(dispatcher.prefixes[i]) > len(dispatcher.prefixes[j])
	})
	return &dispatcher
}

func (dispatcher *CommandDispatcher) Dispatch(text string, messageId string, tanantKey string) {
	for _, prefix := range dispatcher.prefixes {
		if strings.HasPrefix(text, prefix) {
			prompt := strings.TrimSpace(strings.TrimPrefix(text, prefix))
			go dispatcher.prefixCommandMap[prefix].Execute(prompt, dispatcher.larkeeClient, messageId, tanantKey)
			return
		}
	}
	go dispatcher.commandHelpExecute(dispatcher.helpMsgs, dispatcher.larkeeClient, messageId, tanantKey)
}
