package main

import (
	"fmt"
	"time"
)

type Cache interface {
	Get() (string, error)            // 取出accessToken, 当error!=时, accessToken为空或已过期
	Set(value string, ttl int) error // 设置新的accessToken和对应的有效时间(ttl)
}

type SimpleCache struct {
	Value  string
	Expire int64
}

var defaultCache *SimpleCache

func NewSimpleCache() *SimpleCache {
	if defaultCache == nil {
		defaultCache = &SimpleCache{}
	}
	return defaultCache
}

func (d SimpleCache) Get() (string, error) {
	if d.Expire < time.Now().Unix() {
		return "", fmt.Errorf("value已过期")
	} else {
		return d.Value, nil
	}
}

func (d *SimpleCache) Set(value string, ttl int) error {
	d.Value = value
	d.Expire = time.Now().Unix() + int64(ttl)
	return nil
}
