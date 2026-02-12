package dingtalk

import (
	"fmt"
)

func SendDingTalkWarning(content string, atMobiles []string, isAtAll bool) error {

	// 发送文本消息
	textMsg := DingTalkMessage{
		MsgType: "text",
		Text: &Text{
			Content: "SaaS报警=>" + content,
		},
		At: &At{
			AtMobiles: atMobiles, // 要@的手机号
			IsAtAll:   isAtAll,
		},
	}

	// 使用重试机制发送
	if err := SendWithRetry("", "", textMsg, 3); err != nil {
		fmt.Printf("最终发送失败: %v\n", err)
		return err
	}

	return nil
}
