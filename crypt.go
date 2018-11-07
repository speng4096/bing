package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	mrand "math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MsgCrypt struct {
	Config
	aesKey []byte
	iv     []byte
}

func NewMsgCrypt(cfg Config) (MsgCrypt, error) {
	crypt := MsgCrypt{
		Config: cfg,
	}
	b64Key := []byte(crypt.EncodingAESKey + "=")
	key := make([]byte, base64.StdEncoding.DecodedLen(len(b64Key)))
	keyLen, err := base64.StdEncoding.Decode(key, b64Key)
	if err != nil {
		return MsgCrypt{}, fmt.Errorf("EncodingAESKey解码错误, err:%s", err)
	}
	if keyLen != 32 {
		return MsgCrypt{}, fmt.Errorf("EncodingAESKey长度错误")
	}
	crypt.aesKey = key[:keyLen]
	crypt.iv = key[:16]
	mrand.Seed(time.Now().UnixNano())
	return crypt, nil
}

// 用于删除解密后明文的补位字符
func (m MsgCrypt) decodePKCS7(text []byte) []byte {
	pad := int(text[len(text)-1])

	if pad < 1 || pad > 32 {
		pad = 0
	}

	return text[:len(text)-pad]
}

// 用于对需要加密的明文进行填充补位
func (m MsgCrypt) encodePKCS7(text []byte) []byte {
	const BlockSize = 32

	amountToPad := BlockSize - len(text)%BlockSize

	for i := 0; i < amountToPad; i++ {
		text = append(text, byte(amountToPad))
	}

	return text
}

// 微信消息体签名
func (m MsgCrypt) GetSignature(timestamp string, nonce string, msgEncrypt string) string {
	items := []string{m.Token, timestamp, nonce, msgEncrypt}
	sort.Strings(items)
	hash := sha1.New()
	io.WriteString(hash, strings.Join(items, ""))

	return fmt.Sprintf("%x", hash.Sum(nil))
}

//  微信消息解密
func (m *MsgCrypt) Decrypt(xmlEncrypt *[]byte, msgSignature string) ([]byte, error) {
	var msgLen int32
	// xml解码
	encryptMessage := EncryptMessage{}
	xml.Unmarshal(*xmlEncrypt, &encryptMessage)
	// base64解码
	deciphered, err := base64.StdEncoding.DecodeString(encryptMessage.Encrypt)
	if err != nil {
		return nil, err
	}
	// aes解码
	c, err := aes.NewCipher(m.aesKey)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(c, m.iv)
	cbc.CryptBlocks(deciphered, deciphered)
	decoded := m.decodePKCS7(deciphered)
	buf := bytes.NewBuffer(decoded[16:20])
	binary.Read(buf, binary.BigEndian, &msgLen)
	if int(20+msgLen) > len(decoded) {
		return nil, fmt.Errorf("消息解密错误")
	}
	msgDecrypt := decoded[20 : 20+msgLen]
	return msgDecrypt, nil
}

// 微信消息加密
func (m *MsgCrypt) Encrypt(xmlBytes *[]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int32(len(*xmlBytes))); err != nil {
		return nil, err
	}
	msgLen := buf.Bytes()
	randBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, randBytes); err != nil {
		return nil, err
	}
	msgBytes := bytes.Join([][]byte{randBytes, msgLen, *xmlBytes, []byte(m.AppID)}, nil)
	encoded := m.encodePKCS7(msgBytes)

	c, err := aes.NewCipher(m.aesKey)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCEncrypter(c, m.iv)
	cbc.CryptBlocks(encoded, encoded)

	b64Encoded := base64.StdEncoding.EncodeToString(encoded)
	// 构造加密后的xml
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := strconv.Itoa(mrand.Int())
	encryptReply := EncryptReply{
		Encrypt:      b64Encoded,
		TimeStamp:    timestamp,
		Nonce:        nonce,
		MsgSignature: m.GetSignature(timestamp, nonce, b64Encoded),
	}
	return xml.Marshal(encryptReply)
}

// 微信接口验签
func CheckSignature(cfg Config, timestamp string, nonce string, signature string) bool {
	items := []string{cfg.Token, timestamp, nonce}
	sort.Strings(items)
	hash := sha1.New()

	io.WriteString(hash, strings.Join(items, ""))
	return fmt.Sprintf("%x", hash.Sum(nil)) == signature
}
