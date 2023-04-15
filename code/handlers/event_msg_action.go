package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"start-feishubot/services/openai"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	msg = append(msg, openai.Messages{
		Role: "user", Content: a.info.qParsed,
	})

	msg1 := a.handler.sessionCache.GetMsg("xjh:" + *a.info.sessionId)
	fmt.Println(msg1)
	type Options struct {
		ParentMessageId string `json:"parentMessageId"`
		ConversationId  string `json:"conversationId"`
	}
	type RequestBody struct {
		Options       Options `json:"options"`
		Prompt        string  `json:"prompt"`
		SystemMessage string  `json:"systemMessage"`
	}

	requestBody := RequestBody{
		Options:       Options{},
		Prompt:        a.info.qParsed,
		SystemMessage: "You are ChatGPT, a large language model trained by OpenAI. Follow the user's instructions carefully. Respond using markdown.",
	}

	if len(msg) >= 2 {

		m := msg1[len(msg1)-1]
		fmt.Println(m)
		requestBody = RequestBody{
			Options: Options{
				ParentMessageId: m.Role,
				ConversationId:  m.Content,
			},
			Prompt:        a.info.qParsed,
			SystemMessage: "You are ChatGPT, a large language model trained by OpenAI. Follow the user's instructions carefully.",
		}

	}
	// fmt.Println(config)

	//JSON序列化
	configData, _ := json.Marshal(requestBody)
	param := bytes.NewBuffer([]byte(configData))
	url := "http://localhost:3002/api/chat-process"
	//构建http请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, param)

	if err != nil {
		fmt.Println(err)
		return true
	}
	//header
	req.Header.Add("Content-Type", "application/json")

	//发送请求
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	//返回结果
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(string(body))

	var config map[string]interface{}
	err = json.Unmarshal([]byte(string(body)), &config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config)

	msg = append(msg, openai.Messages{
		Role: "assistant", Content: config["data"].(map[string]interface{})["text"].(string),
	})
	a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)

	msg1 = nil
	msg1 = append(msg1, openai.Messages{
		Role:    config["data"].(map[string]interface{})["id"].(string),
		Content: config["data"].(map[string]interface{})["conversationId"].(string),
	})
	a.handler.sessionCache.SetMsg("xjh:"+*a.info.sessionId, msg1)

	// if new topic
	if len(msg) == 2 {
		//fmt.Println("new topic", msg[1].Content)
		sendNewTopicCard(*a.ctx, a.info.sessionId, a.info.msgId,
			config["data"].(map[string]interface{})["text"].(string))
		return false
	}
	err = replyMsg(*a.ctx, config["data"].(map[string]interface{})["text"].(string), a.info.msgId)
	if err != nil {
		replyMsg(*a.ctx, fmt.Sprintf(
			"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}
	return true
}
