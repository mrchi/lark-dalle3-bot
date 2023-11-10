package botconfig

import (
	"encoding/json"
	"io"
	"os"
)

type BotConfig struct {
	BingCookie            string `json:"bing_cookie"`
	LarkVerificationToken string `json:"lark_verification_token"`
	LarkEventEncryptKey   string `json:"lark_event_encrypt_key"`
	LarkAppID             string `json:"lark_app_id"`
	LarkAppSecret         string `json:"lark_app_secret"`
	LarkLogLevel          int    `json:"lark_log_level"`
	LarkEventServerAddr   string `json:"lark_event_server_addr"`
	IsFeishu              bool   `json:"is_feishu"`
	GPTAPIKey             string `json:"gpt_api_key"`
}

func ReadConfigFromFile(filePath string) (*BotConfig, error) {
	var config BotConfig
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
