package main

import (
	"encoding/xml"
	"fmt"
)

// Reply表示所有回复类型
type Reply interface{}

var (
	_ Reply = TextReply{}
	_ Reply = ImageReply{}
	_ Reply = VoiceReply{}
	_ Reply = VideoReply{}
	_ Reply = MusicReply{}
	_ Reply = NewsReply{}
)

// 所有回复的共有成员
type ReplyHeader struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int32  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
}

// 加密回复
type EncryptReply struct {
	Encrypt      string `xml:"Encrypt"`
	MsgSignature string `xml:"MsgSignature"`
	TimeStamp    string `xml:"TimeStamp"`
	Nonce        string `xml:"Nonce"`
}

// 所有回复
type TextReply struct {
	Content string `xml:"Content"`
}
type ImageReply struct {
	MediaID string `xml:"MediaId"`
}
type VoiceReply struct {
	MediaID string `xml:"MediaId"`
}
type VideoReply struct {
	MediaID     string `xml:"MediaId"`
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
}
type MusicReply struct {
	MusicURL     string `xml:"MusicURL"`
	HQMusicUrl   string `xml:"HQMusicUrl"`
	ThumbMediaID string `xml:"ThumbMediaId"`
	Title        string `xml:"Title"`
	Description  string `xml:"Description"`
}
type NewsItem struct {
	PicURL      string `xml:"PicUrl"`
	URL         string `xml:"Url"`
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
}
type NewsReply struct {
	XMLName  xml.Name   `xml:"Articles"`
	Articles []NewsItem `xml:"item"`
}

// 用于渲染XML
type textReply struct {
	XMLName xml.Name `xml:"xml"`
	ReplyHeader
	TextReply
}
type imageReply struct {
	XMLName xml.Name `xml:"xml"`
	ReplyHeader
	ImageReply
}
type voiceReply struct {
	XMLName xml.Name `xml:"xml"`
	ReplyHeader
	VoiceReply
}
type videoReply struct {
	XMLName xml.Name `xml:"xml"`
	ReplyHeader
	VideoReply
}
type musicReply struct {
	XMLName xml.Name `xml:"xml"`
	ReplyHeader
	MusicReply
}
type newsReply struct {
	XMLName xml.Name `xml:"xml"`
	ReplyHeader
	ArticleCount int8   `xml:"ArticleCount"`
	ArticlesXML  []byte `xml:",innerxml"`
}

// 构造回复XML文本
func MakeReply(header *MessageHeader, reply Reply) ([]byte, error) {
	switch reply.(type) {
	case TextReply:
		text := textReply{
			ReplyHeader: ReplyHeader{
				ToUserName:   header.FromUserName,
				FromUserName: header.ToUserName,
				MsgType:      "text",
			},
			TextReply: reply.(TextReply),
		}
		return xml.Marshal(text)
	case ImageReply:
		image := imageReply{
			ReplyHeader: ReplyHeader{
				ToUserName:   header.FromUserName,
				FromUserName: header.ToUserName,
				MsgType:      "image",
			},
			ImageReply: reply.(ImageReply),
		}
		return xml.Marshal(image)
	case VoiceReply:
		voice := voiceReply{
			ReplyHeader: ReplyHeader{
				ToUserName:   header.FromUserName,
				FromUserName: header.ToUserName,
				MsgType:      "voice",
			},
			VoiceReply: reply.(VoiceReply),
		}
		return xml.Marshal(voice)
	case VideoReply:
		video := videoReply{
			ReplyHeader: ReplyHeader{
				ToUserName:   header.FromUserName,
				FromUserName: header.ToUserName,
				MsgType:      "video",
			},
			VideoReply: reply.(VideoReply),
		}
		return xml.Marshal(video)
	case MusicReply:
		music := musicReply{
			ReplyHeader: ReplyHeader{
				ToUserName:   header.FromUserName,
				FromUserName: header.ToUserName,
				MsgType:      "music",
			},
			MusicReply: reply.(MusicReply),
		}
		return xml.Marshal(music)
	case NewsReply:
		articleCount := int8(len(reply.(NewsReply).Articles))
		articlesXML, err := xml.Marshal(reply)
		if err != nil {
			return nil, err
		}
		news := newsReply{
			ReplyHeader: ReplyHeader{
				ToUserName:   header.FromUserName,
				FromUserName: header.ToUserName,
				MsgType:      "news",
			},
			ArticleCount: articleCount,
			ArticlesXML:  articlesXML,
		}
		a, _ := xml.MarshalIndent(news, "", "")
		fmt.Println("articlesXML:", string(a), ";")
		return xml.Marshal(news)
	default:
		return nil, fmt.Errorf("不支持的reply类型, reply=%s, type(reply)=%T", reply, reply)
	}
}
