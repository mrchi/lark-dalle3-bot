package larkclient

import (
	"context"
	"encoding/json"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type LarkTextMessage struct {
	Text string `json:"text"`
}

type LarkClient struct {
	client *lark.Client
}

func (lc *LarkClient) ReplyTextMessage(content string, receiveOpenId string, tenantKey string) error {
	msgContent, err := json.Marshal(LarkTextMessage{Text: content})
	if err != nil {
		return err
	}
	msgBody := larkim.NewCreateMessageReqBodyBuilder().
		MsgType(larkim.MsgTypeText).
		Content(string(msgContent)).
		ReceiveId(receiveOpenId).Build()
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeOpenId).
		Body(msgBody).
		Build()
	_, err = lc.client.Im.Message.Create(context.Background(), req, larkcore.WithTenantKey(tenantKey))
	return err
}

func NewLarkClient(appId string, appSecret string, logLevel larkcore.LogLevel) *LarkClient {
	client := lark.NewClient(appId, appSecret, lark.WithLogLevel(logLevel))
	return &LarkClient{client: client}
}
