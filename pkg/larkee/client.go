package larkee

import (
	"bytes"
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type LarkAPIError struct {
	Code int
	Msg  string
}

func (e LarkAPIError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}

type LarkClient struct {
	client *lark.Client
}

func (lc *LarkClient) ReplyTextMessage(content string, messageId string, tenantKey string) error {
	msgContent, err := NewLarkTextMessageContent(content)
	if err != nil {
		return err
	}
	msgBody := larkim.NewReplyMessageReqBodyBuilder().
		MsgType(larkim.MsgTypeText).Content(msgContent).
		Build()
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageId).Body(msgBody).
		Build()
	resp, err := lc.client.Im.Message.Reply(context.Background(), req, larkcore.WithTenantKey(tenantKey))
	if err != nil {
		return err
	} else if !resp.Success() {
		return LarkAPIError{Code: resp.Code, Msg: resp.Msg}
	} else {
		return nil
	}
}

func (lc *LarkClient) UploadImage(image *[]byte) (string, error) {
	reader := bytes.NewReader(*image)
	reqBody := larkim.NewCreateImageReqBodyBuilder().ImageType(larkim.ImageTypeMessage).Image(reader).Build()
	req := larkim.NewCreateImageReqBuilder().Body(reqBody).Build()
	resp, err := lc.client.Im.Image.Create(context.Background(), req)
	if err != nil {
		return "", err
	} else if !resp.Success() {
		return "", LarkAPIError{Code: resp.Code, Msg: resp.Msg}
	} else {
		return *resp.Data.ImageKey, nil
	}
}

func (lc *LarkClient) ReplyImagesInteractiveMessage(prompt string, imageKeys []string, messageId string, tenantKey string) error {
	msgContent, err := NewLarkImagesInteractiveContent(prompt, imageKeys)
	if err != nil {
		return err
	}
	msgBody := larkim.NewReplyMessageReqBodyBuilder().
		MsgType(larkim.MsgTypeInteractive).Content(msgContent).
		Build()
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageId).Body(msgBody).
		Build()
	resp, err := lc.client.Im.Message.Reply(context.Background(), req, larkcore.WithTenantKey(tenantKey))
	if err != nil {
		return err
	} else if !resp.Success() {
		return LarkAPIError{Code: resp.Code, Msg: resp.Msg}
	} else {
		return nil
	}
}

func (lc *LarkClient) ReplyMarkdownMessage(content string, messageId string, tenantKey string) error {
	msgContent, err := NewLarkMarkdownContent(content)
	if err != nil {
		return err
	}
	msgBody := larkim.NewReplyMessageReqBodyBuilder().
		MsgType(larkim.MsgTypeInteractive).Content(msgContent).
		Build()
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageId).Body(msgBody).
		Build()
	resp, err := lc.client.Im.Message.Reply(context.Background(), req, larkcore.WithTenantKey(tenantKey))
	if err != nil {
		return err
	} else if !resp.Success() {
		return LarkAPIError{Code: resp.Code, Msg: resp.Msg}
	} else {
		return nil
	}
}

func NewLarkClient(appId string, appSecret string, logLevel larkcore.LogLevel) *LarkClient {
	client := lark.NewClient(appId, appSecret, lark.WithLogLevel(logLevel), lark.WithOpenBaseUrl(lark.LarkBaseUrl))
	return &LarkClient{client: client}
}

func NewFeishuClient(appId string, appSecret string, logLevel larkcore.LogLevel) *LarkClient {
	client := lark.NewClient(appId, appSecret, lark.WithLogLevel(logLevel), lark.WithOpenBaseUrl(lark.FeishuBaseUrl))
	return &LarkClient{client: client}
}
