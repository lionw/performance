package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var AccessToken string

type DingTalkMessage struct {
	MsgType  string    `json:"msgtype"`
	Text     *Text     `json:"text,omitempty"`
	Link     *Link     `json:"link,omitempty"`
	Markdown *Markdown `json:"markdown,omitempty"`
	At       *At       `json:"at,omitempty"`
}

type Text struct {
	Content string `json:"content"`
}

type Link struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicURL     string `json:"picUrl"`
	MessageURL string `json:"messageUrl"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type At struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

func generateSign(secret string, timestamp int64) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%d\n%s", timestamp, secret)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func SendDingTalkMessage(webhookURL, secret string, message DingTalkMessage) error {
	var finalURL string
	timestamp := time.Now().UnixNano() / 1e6

	if secret != "" {
		sign := generateSign(secret, timestamp)
		finalURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, url.QueryEscape(sign))
	} else {
		finalURL = webhookURL
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message error: %v", err)
	}

	// 创建自定义HTTP客户端
	client := &http.Client{
		Timeout: 15 * time.Second, // 增加超时时间
	}

	// 创建请求
	req, err := http.NewRequest("POST", finalURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post message error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send message failed, status code: %d", resp.StatusCode)
	}

	// 读取响应体以确保完整传输
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("read response error: %v", err)
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("dingtalk api error: %v, %v", result["errcode"], result["errmsg"])
	}

	return nil
}

func SendWithRetry(webhookURL, secret string, message DingTalkMessage, maxRetries int) error {

	if webhookURL == "" {
		if AccessToken == "" {
			fmt.Println("未设置dingtalk webhookAccessToken")
			return fmt.Errorf("dingtalk webhook url is empty")
		}

		webhookURL = fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%v", AccessToken)
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := SendDingTalkMessage(webhookURL, secret, message); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
		}
	}
	return fmt.Errorf("after %d retries, last error: %v", maxRetries, lastErr)
}
