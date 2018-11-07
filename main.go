package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

// TODO: 替换开发者信息（明文模式）
// 开启加密模式请填写EncodingAESKey字段
var config = Config{
	AppID:     "",
	AppSecret: "",
	Token:     "",
}

// 用户会话
var sessions = map[string]Bing{}

// 自定义菜单
const menu = `{"button":[{"type":"click","name":"开始游戏","key":"Start"},
{"name":"选择回答","sub_button":[{"type":"click","name":"是","key":"Yes"},
{"type":"click","name":"否","key":"No"},{"type":"click","name":"不知道","key":"Pass"}]}]}`

// 根据用户发送的消息生成回复
func getResponse(header *MessageHeader, msg Message) ([]byte, error) {
	// 初始会话
	var uid = header.FromUserName
	bing, ok := sessions[uid]
	if !ok {
		if _bing, err := NewBing(); err != nil {
			return MakeReply(header, TextReply{Content: "小冰崩溃了 :-("})
		} else {
			sessions[uid] = _bing
			bing = _bing
		}
	}
	// 交互
	switch msg.(type) {
	case TextMessage:
		var content = msg.(TextMessage).Content
		if content == "开始" {
			if _bing, err := NewBing(); err != nil {
				return MakeReply(header, TextReply{Content: "小冰崩溃了 :-("})
			} else {
				sessions[uid] = _bing
				return MakeReply(header, TextReply{Content: _bing.send("开始")})
			}
		} else {
			return MakeReply(header, TextReply{Content: bing.send(content)})
		}
	case SubscribeEvent:
		return MakeReply(header, TextReply{Content: `我是小冰，想挑战我的【读心术】吗？
规则很简单。你在心里想好一个人的名字，然后按下【开始】。我将问你15个问题，之后，我就会轻松地猜到那个人是谁。
我已经准备好了，开始吧？`})
	case MenuClickEvent:
		var answer int
		switch msg.(MenuClickEvent).EventKey {
		case "Start":
			if _bing, err := NewBing(); err != nil {
				return MakeReply(header, TextReply{Content: "小冰崩溃了 :-("})
			} else {
				sessions[uid] = _bing
				return MakeReply(header, TextReply{Content: _bing.send("开始")})
			}
		case "Yes":
			answer = Yes
		case "No":
			answer = No
		default:
			answer = Pass
		}
		return MakeReply(header, TextReply{Content: bing.Next(answer)})
	default:
		return MakeReply(header, TextReply{Content: "啥？"})
	}
}

// 接口验签中间件
func checker(c *gin.Context) {
	signature := c.Query("signature")
	nonce := c.Query("nonce")
	timestamp := c.Query("timestamp")

	if !CheckSignature(config, timestamp, nonce, signature) {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func main() {
	router := gin.New()
	// 开发者认证接口
	router.GET("/wechat", checker, func(c *gin.Context) {
		echostr := c.Query("echostr")
		c.String(http.StatusOK, echostr)
	})
	// 生成微信菜单
	client := NewClient(config)
	if err := client.SetMenu(menu); err != nil {
		fmt.Println(err)
	} else {
		log.Println("创建菜单成功")
	}
	// 消息处理接口
	router.POST("/wechat", checker, func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, "")
			return
		}
		header, message, err := Unmarshal(&body)
		if err != nil {
			c.String(http.StatusBadRequest, "")
			return
		}
		resp, err := getResponse(header, message)
		if err != nil {
			c.String(http.StatusOK, "")
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(resp)
	})

	router.Run(":4321")
}
