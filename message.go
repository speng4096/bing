// 编解码微信消息
package main

import (
	"encoding/xml"
	"fmt"
	"regexp"
)

// Message表示所有消息类型, 包含普通消息(XXXMessage)和事件消息(XXXEvent)
type Message interface{}

var (
	_ Message = TextMessage{}
	_ Message = ImageMessage{}
	_ Message = VoiceMessage{}
	_ Message = VideoMessage{}
	_ Message = ShortVideoMessage{}
	_ Message = LocationMessage{}
	_ Message = LinkMessage{}

	_ Message = SubscribeEvent{}
	_ Message = UnSubscribeEvent{}
	_ Message = ScanEvent{}
	_ Message = LocationEvent{}
	_ Message = MenuClickEvent{}
	_ Message = MenuViewEvent{}
)
var (
	reEvent = regexp.MustCompile(`<Event><!\[CDATA\[(\w+)]]></Event>`)
	reType  = regexp.MustCompile(`<MsgType><!\[CDATA\[(\w+)]]></MsgType>`)
)

// Message共有成员，可用于构造Reply
type MessageHeader struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int32  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	// MsgID可用于排重
	// 对于事件消息, 微信文档建议MsgID=FromUserName+CreateTime
	MsgID string `xml:"MsgId"`
}

// 加密消息
type EncryptMessage struct {
	ToUserName string
	Encrypt    string
}

// 文本消息
type TextMessage struct {
	MessageHeader
	Content string `xml:"Content"`
}

// 图片消息
type ImageMessage struct {
	MessageHeader
	PicURL  string `xml:"PicUrl"`
	MediaID string `xml:"MediaId"`
}

// 语音消息
type VoiceMessage struct {
	MessageHeader
	MediaId string `xml:"MediaId"`
	Format  string `xml:"Format"`
}

// 视频消息
type VideoMessage struct {
	MessageHeader
	MediaID      string `xml:"MediaId"`
	ThumbMediaID string `xml:"ThumbMediaId"`
}

// 小视频消息
type ShortVideoMessage struct {
	MessageHeader
	MediaID      string `xml:"MediaId"`
	ThumbMediaID string `xml:"ThumbMediaId"`
}

// 地理位置消息
type LocationMessage struct {
	MessageHeader
	X     float32 `xml:"Location_X"`
	Y     float32 `xml:"Location_Y"`
	Scale int32   `xml:"Scale"`
	Label string  `xml:"Label"`
}

// 链接消息
type LinkMessage struct {
	MessageHeader
	Title       float32 `xml:"Title"`
	Description float32 `xml:"Description"`
	Url         int32   `xml:"Url"`
}

// 关注事件
type SubscribeEvent struct {
	MessageHeader
	Event string `xml:"Event"` // subscribe
	// 当用户扫描二维码关注时, EventKey和Ticket不为空
	// qrscene_为前缀，后面为二维码的参数值
	EventKey string `xml:"EventKey"` // TODO: 可提取出参数值, 存储为int32
	Ticket   string `xml:"Ticket"`   // 二维码的ticket，可用来换取二维码图片
}

// 取消关注事件
type UnSubscribeEvent struct {
	MessageHeader
	Event string `xml:"Event"` // unsubscribe
}

// 已关注用户扫描带参数二维码事件
type ScanEvent struct {
	MessageHeader
	Event    string `xml:"Event"`    // SCAN
	EventKey int32  `xml:"EventKey"` // 创建二维码时的二维码scene_id
	Ticket   string `xml:"Ticket"`   // 二维码的ticket，可用来换取二维码图片
}

// 上报地理位置事件
type LocationEvent struct {
	MessageHeader
	Event     string  `xml:"Event"`     // LOCATION
	Latitude  float32 `xml:"Latitude"`  // 地理位置纬度
	Longitude float32 `xml:"Longitude"` // 地理位置经度
	Precision float32 `xml:"Precision"` // 地理位置精度
}

// 点击自定义菜单拉取消息
type MenuClickEvent struct {
	MessageHeader
	Event    string `xml:"Event"`    // CLICK
	EventKey string `xml:"EventKey"` // 事件KEY值，与自定义菜单接口中KEY值对应
}

// 点击自定义菜单跳转链接
type MenuViewEvent struct {
	MessageHeader
	Event    string `xml:"Event"`    // VIEW
	EventKey string `xml:"EventKey"` // 事件KEY值，设置的跳转URL
}

func unmarshalMessage(msgType string, xmlBytes *[]byte) (*MessageHeader, Message, error) {
	switch msgType {
	case "text":
		msg := TextMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	case "image":
		msg := ImageMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	case "voice":
		msg := VoiceMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	case "video":
		msg := VideoMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	case "shortvideo":
		msg := ShortVideoMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	case "location":
		msg := LocationMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	case "link":
		msg := LinkMessage{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		return header, msg, err
	default:
		return nil, nil, fmt.Errorf("错误的消息类型: %s", msgType)
	}
}

func unmarshalEvent(msgEvent string, xmlBytes *[]byte) (*MessageHeader, Message, error) {
	switch msgEvent {
	case "subscribe":
		msg := SubscribeEvent{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		header.MsgID = header.FromUserName + string(header.CreateTime)
		return header, msg, err
	case "unsubscribe":
		msg := UnSubscribeEvent{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		header.MsgID = header.FromUserName + string(header.CreateTime)
		return header, msg, err
	case "SCAN":
		msg := ScanEvent{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		header.MsgID = header.FromUserName + string(header.CreateTime)
		return header, msg, err
	case "LOCATION":
		msg := LocationEvent{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		header.MsgID = header.FromUserName + string(header.CreateTime)
		return header, msg, err
	case "CLICK":
		msg := MenuClickEvent{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		header.MsgID = header.FromUserName + string(header.CreateTime)
		return header, msg, err
	case "VIEW":
		msg := MenuViewEvent{}
		err := xml.Unmarshal(*xmlBytes, &msg)
		header := &msg.MessageHeader
		header.MsgID = header.FromUserName + string(header.CreateTime)
		return header, msg, err
	default:
		return nil, nil, fmt.Errorf("错误的事件类型: %s", msgEvent)
	}
}

// 反序列化微信消息为struct
func Unmarshal(body *[]byte) (*MessageHeader, Message, error) {
	items := reType.FindSubmatch(*body)
	if len(items) == 0 {
		return nil, nil, fmt.Errorf("微信消息格式错误, 未找到MsgType字段")
	}
	if msgType := string(items[1]); msgType == "event" {
		eventItems := reEvent.FindSubmatch(*body)
		event := string(eventItems[1])
		return unmarshalEvent(event, body)
	} else {
		return unmarshalMessage(msgType, body)
	}
}
