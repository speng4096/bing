package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const apiURL = "https://api.weixin.qq.com/cgi-bin"

type Client struct {
	config Config
	cache  Cache
}

func NewClient(cfg Config) Client {
	var cache Cache
	if cfg.Cache == nil {
		cache = NewSimpleCache()
	} else {
		cache = cfg.Cache
	}
	return Client{
		config: cfg,
		cache:  cache,
	}
}

func (c Client) getToken() (string, error) {
	cacheValue, err := c.cache.Get()
	// 缓存有效
	if err == nil {
		return cacheValue, nil
	}
	// 缓存已失效或为空, 拉取AccessToken
	values := url.Values{}
	values.Add("grant_type", "client_credential")
	values.Add("appid", c.config.AppID)
	values.Add("secret", c.config.AppSecret)
	resp, err := http.Get(apiURL + "/token?" + values.Encode())
	if err != nil {
		return "", fmt.Errorf("拉取AccessToken错误: %s", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("拉取AccessToken错误: %s", err)
	}
	j := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	json.Unmarshal(b, &j)
	if j.AccessToken == "" {
		return "", fmt.Errorf("拉取AccessToken错误: %s", b)
	}
	log.Printf("获取到AccessToken: %s\n", j.AccessToken)
	c.cache.Set(j.AccessToken, j.ExpiresIn)
	return j.AccessToken, nil
}

func (c Client) get(path string) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}
	r, err := http.Get(apiURL + path + "?access_token=" + token)
	if err != nil {
		return "", fmt.Errorf("微信接口返回错误: %s", err)
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return "", fmt.Errorf("获取微信接口响应错误: %s", err)
	}
	return string(b), nil
}

func (c Client) post(path string, body string) (string, error) {
	token, err := c.getToken()
	if err != nil {
		return "", err
	}
	reader := strings.NewReader(body)
	r, err := http.Post(apiURL+path+"?access_token="+token, "", reader)
	if err != nil {
		return "", fmt.Errorf("微信接口返回错误: %s", err)
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return "", fmt.Errorf("获取微信接口响应错误: %s", err)
	}
	return string(b), nil
}

func (c Client) SetMenu(menu string) error {
	if s, err := c.post("/menu/create", menu); err != nil {
		return err
	} else if strings.Index(s, `"ok"`) == -1 {
		return fmt.Errorf("创建菜单失败: %s", s)
	} else {
		return nil
	}
}

func (c Client) GetMenu() (string, error) {
	return c.get("/menu/get")
}
