package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/mrchi/lark-dalle3-bot/pkg/dispatcher"
	larkee "github.com/mrchi/lark-dalle3-bot/pkg/larkee"
)

var commandBalance = dispatcher.Command{
	Prefix:  "/balance",
	HelpMsg: "**/balance** Get tokens balance of Bing cookie",
	Execute: func(prompt string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string) {
		balance, err := bingClient.GetTokenBalance()
		var replyMsg string
		if err != nil {
			replyMsg = fmt.Sprintf("[Error]%s", err.Error())
		} else {
			replyMsg = fmt.Sprintf("Tokens left %d.", balance)
		}
		larkeeClient.ReplyTextMessage(replyMsg, messageId, tanantKey)
	},
}

var commandPrompt = dispatcher.Command{
	Prefix:  "/prompt",
	HelpMsg: "**/prompt &#60;Your prompt&#62;** Create image with prompt",
	Execute: func(prompt string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string) {
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
	},
}

func commandHelpExecute(helpMsgs []string, larkeeClient *larkee.LarkClient, messageId string, tanantKey string) {
	msg := "Welcome. Supported commands:\n\n" + strings.Join(helpMsgs, "\n")
	larkeeClient.ReplyMarkdownMessage(msg, messageId, tanantKey)
}
