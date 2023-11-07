package larkee

import (
	"encoding/json"
)

type LarkTextMessage struct {
	Text string `json:"text"`
}

type LarkInteractiveMessage struct {
	Elements []any `json:"elements"`
}

type LarkInteractiveMsgText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type LarkInteractiveMsgColumnSet struct {
	Tag     string                     `json:"tag"`
	Columns []LarkInteractiveMsgColumn `json:"columns"`
}

type LarkInteractiveMsgColumn struct {
	Tag           string                  `json:"tag"`
	Width         string                  `json:"width"`
	Weight        int                     `json:"weight"`
	VerticalAlign string                  `json:"vertical_align"`
	Elements      []LarkInteractiveMsgImg `json:"elements"`
}

type LarkInteractiveMsgImg struct {
	Tag     string                 `json:"tag"`
	ImgKey  string                 `json:"img_key"`
	Alt     LarkInteractiveMsgText `json:"alt"`
	Mode    string                 `json:"mode"`
	Preview bool                   `json:"preview"`
}

func NewLarkTextMessageContent(text string) (string, error) {
	msg := LarkTextMessage{Text: text}
	msgContent, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(msgContent), nil
}

func NewLarkImagesInteractiveContent(prompt string, imageKeys []string) (string, error) {
	textModule := LarkInteractiveMsgText{Tag: "markdown", Content: prompt}
	msg := LarkInteractiveMessage{Elements: []any{textModule}}

	// 每列展示 2 个图片
	for i := 0; i < len(imageKeys); i += 2 {
		columnSet := LarkInteractiveMsgColumnSet{Tag: "column_set"}
		columnSet.Columns = append(
			columnSet.Columns,
			LarkInteractiveMsgColumn{
				Tag:           "column",
				Width:         "weighted",
				Weight:        1,
				VerticalAlign: "top",
				Elements: []LarkInteractiveMsgImg{
					{
						Tag:     "img",
						ImgKey:  imageKeys[i],
						Alt:     LarkInteractiveMsgText{Tag: "plain_text", Content: ""},
						Mode:    "fit_horizontal",
						Preview: true,
					},
				},
			},
		)
		if i+1 < len(imageKeys) {
			columnSet.Columns = append(
				columnSet.Columns,
				LarkInteractiveMsgColumn{
					Tag:           "column",
					Width:         "weighted",
					Weight:        1,
					VerticalAlign: "top",
					Elements: []LarkInteractiveMsgImg{
						{
							Tag:     "img",
							ImgKey:  imageKeys[i+1],
							Alt:     LarkInteractiveMsgText{Tag: "plain_text", Content: ""},
							Mode:    "fit_horizontal",
							Preview: true,
						},
					},
				},
			)
		}
		msg.Elements = append(msg.Elements, columnSet)
	}

	msgContent, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(msgContent), nil
}

func NewLarkMarkdownContent(content string) (string, error) {
	textModule := LarkInteractiveMsgText{Tag: "markdown", Content: content}
	msg := LarkInteractiveMessage{Elements: []any{textModule}}

	msgContent, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(msgContent), nil
}
