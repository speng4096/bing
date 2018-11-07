package main

import (
	"fmt"
	"github.com/imroc/req"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"
	"v2ray.com/core/common/uuid"
)

type Bing struct {
	client   *req.Req
	senderID string
}

// 接口
const (
	baseURL  = "http://webapps.msxiaobing.com"
	entryURL = baseURL + "/mindreader"
	respURL  = baseURL + "/simplechat/getresponse?workflow=Q20"
	authURL  = baseURL + "/api/wechatAuthorize/signature?url=" + entryURL
)

// 回答
const (
	Yes = iota
	No
	Pass
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var answerRe = regexp.MustCompile(`"Text":"([^"]+)"`)
var headers = req.Header{
	"Accept-Encoding":  "",
	"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7",
	"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1)",
	"Content-Type":     "application/json",
	"X-Requested-With": "XMLHttpRequest",
	"Referer":          entryURL,
}
var cookieSession, cookieUser *http.Cookie

// 生成随机字符串
func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// 新建会话
func NewBing() (Bing, error) {
	// 请求首页，获取Cookie:cpid,salt,ARRAffinity
	client := req.New()
	r, _ := client.Get(entryURL)
	resp := r.Response()
	if resp.StatusCode != 200 {
		return Bing{}, fmt.Errorf("请求首页失败")
	}
	// 随机生成Cookie:ai_session_id,ai_user
	now := fmt.Sprintf("%.1f", float64(time.Now().UnixNano()/1e6))
	aiSession := fmt.Sprintf("%s|%s|%s", randString(5), now, now)
	aiUser := fmt.Sprintf("%s|%s", randString(5), fmt.Sprintf(time.Now().UTC().Format("2006-01-02T15:04:05.999Z")))
	cookieSession = &http.Cookie{Name: "ai_session_id", Value: aiSession}
	cookieUser = &http.Cookie{Name: "ai_user", Value: aiUser}

	// 请求签名页面
	r, _ = client.Get(authURL, headers)
	resp = r.Response()
	if resp.StatusCode != 200 {
		return Bing{}, fmt.Errorf("请求签名页失败")
	}

	// 请求回复页，获取Cookie:cookieid
	_senderID := uuid.New()
	senderID := _senderID.String()
	body := fmt.Sprintf(`{"SenderId":"%s","Content":{"Text":"玩","Image":"","Metadata":{"Q20H5Enter":"true"}}}`, senderID)
	r, _ = client.Post(respURL, headers, body, cookieUser, cookieSession)
	resp = r.Response()
	if resp.StatusCode != 200 {
		return Bing{}, fmt.Errorf("新建游戏失败")
	}

	return Bing{
		client:   client,
		senderID: senderID,
	}, nil
}

// 与小冰聊天
func (b Bing) send(a string) string {
	body := fmt.Sprintf(`{"SenderId":"%s","Content":{"Text":"%s","Image":""}}`, b.senderID, a)
	r, _ := b.client.Post(respURL, body, headers, cookieUser, cookieSession)
	html, err := r.ToString()
	resp := r.Response()
	if err != nil || resp.StatusCode != 200 {
		return "小冰失联了……"
	}
	items := answerRe.FindAllStringSubmatch(html, -1)
	// 从响应包中取出回复文本
	if len(items) > 0 {
		var q, gap string
		for _, v := range items {
			q += gap + v[1]
			gap = " "
		}
		log.Println("A:", a, "Q:", q)
		return q
	} else {
		return "小冰不知怎么回答"
	}
}

// 回答游戏选项
func (b Bing) Next(answer int) string {
	var s string
	switch answer {
	case Yes:
		s = "是"
	case No:
		s = "不是"
	case Pass:
		s = "不知道"
	}
	return b.send(s)
}
