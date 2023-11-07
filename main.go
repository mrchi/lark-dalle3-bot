package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	bingdalle3 "github.com/mrchi/bing-dalle3"
	larkee "github.com/mrchi/lark-dalle3-bot/pkg/larkee"
)

var (
	config                 BotConfig
	bingClient             *bingdalle3.BingDalle3
	larkeeClient           *larkee.LarkClient
	larkEventDispatcher    *dispatcher.EventDispatcher
	regexRemoveAt          = regexp.MustCompile(`@_all|@_user_\d+\s*`)
	regexExtractCmdAndBody = regexp.MustCompile(`\s*(/balance|/prompt|/help)\s*(.*)`)
	helpMessage            = []string{
		"欢迎使用 DALL·E 3 Bot。目前支持以下命令：",
		"",
		"**/balance** 查询 Cookie 剩余额度",
		"**/prompt &#60;Your prompt&#62;** 生成图片",
		"**/help** 查看帮助",
	}
)

type BotConfig struct {
	BingCookie            string `json:"bing_cookie"`
	LarkVerificationToken string `json:"lark_verification_token"`
	LarkEventEncryptKey   string `json:"lark_event_encrypt_key"`
	LarkAppID             string `json:"lark_app_id"`
	LarkAppSecret         string `json:"lark_app_secret"`
	LarkLogLevel          int    `json:"lark_log_level"`
	LarkEventServerAddr   string `json:"lark_event_server_addr"`
}

func init() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Read config failed.", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln("Read config failed.", err)
	}

	if err := json.Unmarshal(content, &config); err != nil {
		log.Fatalln("Wrong format in config file.", err)
	}

	bingClient = bingdalle3.NewBingDalle3(config.BingCookie)
	larkeeClient = larkee.NewLarkClient(config.LarkAppID, config.LarkAppSecret, larkcore.LogLevel(config.LarkLogLevel))
	larkEventDispatcher = dispatcher.NewEventDispatcher(config.LarkVerificationToken, config.LarkEventEncryptKey)
}

func messageHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	messageId := *event.Event.Message.MessageId
	tanantKey := event.TenantKey()

	// 忽略非文本消息
	if *event.Event.Message.MessageType != "text" {
		go commandHelpHandler(messageId, tanantKey)
		return nil
	}

	// 获取文本消息内容
	var msgContent larkee.LarkTextMessage
	err := json.Unmarshal([]byte(*event.Event.Message.Content), &msgContent)
	if err != nil {
		log.Printf("Unmarshal message content failed, %s", err.Error())
		return nil
	}

	// 过滤 @ 信息，分离命令和 body
	text := regexRemoveAt.ReplaceAllString(msgContent.Text, "")
	matches := regexExtractCmdAndBody.FindStringSubmatch(text)
	if matches == nil {
		go commandHelpHandler(messageId, tanantKey)
		return nil
	}

	switch matches[1] {
	case "/help":
		go commandHelpHandler(messageId, tanantKey)
	case "/balance":
		go commandBalanceHandler(messageId, tanantKey)
	case "/prompt":
		go commandPromptHandler(strings.TrimSpace(matches[2]), messageId, tanantKey)
	}
	return nil
}

func commandHelpHandler(messageId, tanantKey string) {
	larkeeClient.ReplyMarkdownMessage(strings.Join(helpMessage, "\n"), messageId, tanantKey)
}

func commandBalanceHandler(messageId, tanantKey string) {
	balance, err := bingClient.GetTokenBalance()
	var replyMsg string
	if err != nil {
		replyMsg = fmt.Sprintf("[Error]%s", err.Error())
	} else {
		replyMsg = fmt.Sprintf("Tokens left %d.", balance)
	}
	larkeeClient.ReplyTextMessage(replyMsg, messageId, tanantKey)
}

func commandPromptHandler(prompt string, messageId, tanantKey string) {
	// 判断 prompt 不为空
	if prompt == "" {
		larkeeClient.ReplyTextMessage("[Error]Prompt is empty", messageId, tanantKey)
		return
	}

	// 提交创建请求
	writingId, err := bingClient.CreateImage(prompt)
	if err != nil {
		larkeeClient.ReplyTextMessage(fmt.Sprintf("[Error]%s", err.Error()), messageId, tanantKey)
		return
	}

	// 返回一些提示信息
	messages := []string{"Creating now...", "WritingID is " + writingId}
	balance, err := bingClient.GetTokenBalance()
	var balanceMsg string
	if err != nil {
		balanceMsg = fmt.Sprintf("Tokens left invalid, error: %s.", err.Error())
	} else if balance == 0 {
		balanceMsg = "Tokens run out, image generation may take longer."
	} else {
		balanceMsg = fmt.Sprintf("Tokens left %d.", balance)
	}
	messages = append(messages, balanceMsg)
	larkeeClient.ReplyTextMessage(strings.Join(messages, "\n"), messageId, tanantKey)

	// 获取生成结果
	imageUrls, err := bingClient.QueryResult(writingId, prompt)
	if err != nil {
		larkeeClient.ReplyTextMessage(fmt.Sprintf("[Error]%s", err.Error()), messageId, tanantKey)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(imageUrls))
	imageKeys := make([]string, len(imageUrls))
	for idx, imageUrl := range imageUrls {
		go func(idx int, imageUrl string) {
			defer wg.Done()
			reader, err := bingClient.DownloadImage(imageUrl)
			if err != nil {
				log.Printf("Download image failed, %s", err.Error())
				return
			}
			imageKey, err := larkeeClient.UploadImage(reader)
			if err != nil {
				log.Printf("Upload image failed, %s", err.Error())
				return
			}
			imageKeys[idx] = imageKey
		}(idx, imageUrl)
	}
	wg.Wait()
	larkeeClient.ReplyImagesInteractiveMessage(prompt, imageKeys, messageId, tanantKey)
}

func main() {
	larkEventDispatcher.OnP2MessageReceiveV1(messageHandler)

	http.HandleFunc(
		"/dalle3",
		httpserverext.NewEventHandlerFunc(
			larkEventDispatcher,
			larkevent.WithLogLevel(larkcore.LogLevel(config.LarkLogLevel)),
		),
	)

	log.Printf("start server at: %s\n", config.LarkEventServerAddr)
	err := http.ListenAndServe(config.LarkEventServerAddr, nil)
	if err != nil {
		panic(err)
	}
}
