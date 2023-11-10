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
	"github.com/sashabaranov/go-openai"
)

var (
	config     *botconfig.BotConfig
	bingClient *bingdalle3.BingDalle3
	gptClient  *openai.Client
)

func init() {
	var err error
	config, err = botconfig.ReadConfigFromFile("./config.json")
	if err != nil {
		panic(err)
	}
	bingClient = bingdalle3.NewBingDalle3(config.BingCookie)
	gptClient = openai.NewClient(config.GPTAPIKey)
}

func main() {
	var larkeeClient *larkee.LarkClient
	if config.IsFeishu {
		larkeeClient = larkee.NewFeishuClient(config.LarkAppID, config.LarkAppSecret, larkcore.LogLevel(config.LarkLogLevel))
		log.Println("Initialize client for Feishu")
	} else {
		larkeeClient = larkee.NewLarkClient(config.LarkAppID, config.LarkAppSecret, larkcore.LogLevel(config.LarkLogLevel))
		log.Println("Initialize client for Lark")
	}
	commandDispatcher := cmddispatcher.NewCommandDispatcher(larkeeClient, commandHelpExecute, commandBalance, commandPrompt, commandPromptPro)

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

	urlPath := "/dalle3"
	http.HandleFunc(
		urlPath,
		httpserverext.NewEventHandlerFunc(
			larkEventDispatcher,
			larkevent.WithLogLevel(larkcore.LogLevel(config.LarkLogLevel)),
		),
	)
	log.Printf("Start server at %s, url path is %s\n", config.LarkEventServerAddr, urlPath)
	http.ListenAndServe(config.LarkEventServerAddr, nil)
}
