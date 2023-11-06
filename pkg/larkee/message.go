package larkee

import "encoding/json"

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

func NewLarkInteractiveMessageContent(images []ImageUploadResult, text string) (string, error) {
	var columns []LarkInteractiveMsgColumn
	for _, image := range images {
		element := LarkInteractiveMsgImg{
			Tag:     "img",
			ImgKey:  image.imageKey,
			Alt:     LarkInteractiveMsgText{Tag: "plain_text", Content: image.errMsg},
			Mode:    "fit_horizontal",
			Preview: true,
		}
		columns = append(
			columns,
			LarkInteractiveMsgColumn{
				Tag:           "column",
				Width:         "weighted",
				Weight:        1,
				VerticalAlign: "top",
				Elements:      []LarkInteractiveMsgImg{element},
			},
		)
	}
	imgModule := LarkInteractiveMsgColumnSet{Tag: "column_set", Columns: columns}
	textModule := LarkInteractiveMsgText{Tag: "markdown", Content: text}

	msg := LarkInteractiveMessage{Elements: []any{textModule, imgModule}}
	msgContent, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(msgContent), nil
}
