package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const (
	LISTEN_ADDR  = "localhost:8000"
	LOG_LEVEL    = larkcore.LogLevelDebug
	HELP_MESSAGE = `欢迎使用 DALL·E 3 Bot。目前支持以下命令：
- /balance - 查询 Cookie 剩余额度
- /prompt <Your prompt> - 生成图片
- /help - 查看帮助
`
)

var (
	larkClient             *lark.Client
	larkEventDispatcher    *dispatcher.EventDispatcher
	regexRemoveAt          = regexp.MustCompile(`@_all|@_user_\d+\s*`)
	regexExtractCmdAndBody = regexp.MustCompile(`\s*(/balance|/prompt|/help)\s*(.*)`)
)

type TextMessage struct {
	Text string `json:"text"`
}

func init() {
	verificationToken := os.Getenv("VERIFICATION_TOKEN")
	eventEncryptKey := os.Getenv("ENCRYPT_KEY")
	appId := os.Getenv("APP_ID")
	appSecret := os.Getenv("APP_SECRET")

	larkClient = lark.NewClient(appId, appSecret, lark.WithLogLevel(LOG_LEVEL))
	larkEventDispatcher = dispatcher.NewEventDispatcher(verificationToken, eventEncryptKey)
}

func messageHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	fmt.Println(larkcore.Prettify(event))
	// 忽略非文本消息
	if *event.Event.Message.MessageType != "text" {
		return nil
	}
	// 获取文本消息内容
	var msgContent TextMessage
	err := json.Unmarshal([]byte(*event.Event.Message.Content), &msgContent)
	if err != nil {
		return err
	}
	// 过滤 @ 信息，分离命令和 body
	text := regexRemoveAt.ReplaceAllString(msgContent.Text, "")
	matches := regexExtractCmdAndBody.FindStringSubmatch(text)
	if matches == nil {
		return nil
	}
	switch matches[1] {
	case "/balance":
		fmt.Println("Balance")
	case "/prompt":
		fmt.Println("Prompt", matches[2])
	}
	return nil
}

func main() {
	larkEventDispatcher.OnP2MessageReceiveV1(messageHandler)

	http.HandleFunc("/", httpserverext.NewEventHandlerFunc(larkEventDispatcher, larkevent.WithLogLevel(LOG_LEVEL)))

	log.Printf("start server at: %s\n", LISTEN_ADDR)
	err := http.ListenAndServe(LISTEN_ADDR, nil)
	if err != nil {
		panic(err)
	}
}
