package larkee

import (
	"context"
	"fmt"
	"io"
	"sync"

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

type ImageUploadResult struct {
	imageKey string
	errMsg   string
}

func (lc *LarkClient) ReplyTextMessage(content string, receiveOpenId string, tenantKey string) error {
	msgContent, err := NewLarkTextMessageContent(content)
	if err != nil {
		return err
	}
	msgBody := larkim.NewCreateMessageReqBodyBuilder().
		MsgType(larkim.MsgTypeText).Content(msgContent).ReceiveId(receiveOpenId).
		Build()
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeOpenId).Body(msgBody).
		Build()
	resp, err := lc.client.Im.Message.Create(context.Background(), req, larkcore.WithTenantKey(tenantKey))
	if err != nil {
		return err
	} else if !resp.Success() {
		return LarkAPIError{Code: resp.Code, Msg: resp.Msg}
	} else {
		return nil
	}
}

func (lc *LarkClient) UploadImages(images []io.Reader) []ImageUploadResult {
	result := make([]ImageUploadResult, len(images))

	var wg sync.WaitGroup
	wg.Add(len(images))
	for idx, image := range images {
		go func(idx int, image io.Reader) {
			defer wg.Done()
			reqBody := larkim.NewCreateImageReqBodyBuilder().ImageType(larkim.ImageTypeMessage).Image(image).Build()
			req := larkim.NewCreateImageReqBuilder().Body(reqBody).Build()
			resp, err := lc.client.Im.Image.Create(context.Background(), req)
			if err != nil {
				result[idx].errMsg = err.Error()
			} else if !resp.Success() {
				result[idx].errMsg = LarkAPIError{Code: resp.Code, Msg: resp.Msg}.Error()
			} else {
				result[idx].imageKey = *resp.Data.ImageKey
			}
		}(idx, image)
	}
	wg.Wait()
	return result
}

func (lc *LarkClient) ReplyImagesInteractiveMessage(images []ImageUploadResult, content string, receiveOpenId string, tenantKey string) error {
	msgContent, err := NewLarkInteractiveMessageContent(images, content)
	if err != nil {
		return err
	}
	msgBody := larkim.NewCreateMessageReqBodyBuilder().
		MsgType(larkim.MsgTypeInteractive).
		Content(string(msgContent)).
		ReceiveId(receiveOpenId).Build()
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeOpenId).
		Body(msgBody).
		Build()
	resp, err := lc.client.Im.Message.Create(context.Background(), req, larkcore.WithTenantKey(tenantKey))
	if err != nil {
		return err
	} else if !resp.Success() {
		return LarkAPIError{Code: resp.Code, Msg: resp.Msg}
	} else {
		return nil
	}
}

func NewLarkClient(appId string, appSecret string, logLevel larkcore.LogLevel) *LarkClient {
	client := lark.NewClient(appId, appSecret, lark.WithLogLevel(logLevel))
	return &LarkClient{client: client}
}
