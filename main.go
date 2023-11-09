package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	bingdalle3 "github.com/mrchi/bing-dalle3"
	"github.com/mrchi/lark-dalle3-bot/internal/botconfig"
	cmddispatcher "github.com/mrchi/lark-dalle3-bot/pkg/dispatcher"
	larkee "github.com/mrchi/lark-dalle3-bot/pkg/larkee"
)

var (
	config     *botconfig.BotConfig
	bingClient *bingdalle3.BingDalle3
)

func init() {
	var err error
	config, err = botconfig.ReadConfigFromFile("./config.json")
	if err != nil {
		panic(err)
	}
	bingClient = bingdalle3.NewBingDalle3(config.BingCookie)
}

func main() {
	var larkeeClient *larkee.LarkClient
	if config.IsFeishu {
		larkeeClient = larkee.NewFeishuClient(config.LarkAppID, config.LarkAppSecret, larkcore.LogLevel(config.LarkLogLevel))
	} else {
		larkeeClient = larkee.NewLarkClient(config.LarkAppID, config.LarkAppSecret, larkcore.LogLevel(config.LarkLogLevel))
	}
	commandDispatcher := cmddispatcher.NewCommandDispatcher(larkeeClient, commandHelpExecute, commandBalance, commandPrompt)

	larkEventDispatcher := dispatcher.NewEventDispatcher(config.LarkVerificationToken, config.LarkEventEncryptKey)
	larkEventDispatcher.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		// 获取文本消息内容
		var msgContent larkee.LarkTextMessage
		json.Unmarshal([]byte(*event.Event.Message.Content), &msgContent)
		// 过滤 @ 信息
		text := regexp.MustCompile(`\s*@_all|@_user_\d+\s*`).ReplaceAllString(msgContent.Text, "")

		commandDispatcher.Dispatch(text, *event.Event.Message.MessageId, event.TenantKey())
		return nil
	},
	)

	http.HandleFunc(
		"/dalle3",
		httpserverext.NewEventHandlerFunc(
			larkEventDispatcher,
			larkevent.WithLogLevel(larkcore.LogLevel(config.LarkLogLevel)),
		),
	)
	log.Printf("start server at: %s\n", config.LarkEventServerAddr)
	http.ListenAndServe(config.LarkEventServerAddr, nil)
}
