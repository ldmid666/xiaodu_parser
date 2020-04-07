package main

import (
	"bytes"
	"github.com/pkg/errors"
	"log"
	"crypto/aes"
	"crypto/cipher"
	"github.com/golang/protobuf/proto"
	devProto "xiaodu_parser/dev_proto"

)

func CreateMarshalDown(downMsg  *devProto.Payload) ([]byte,error) {
	pData, err := proto.Marshal(downMsg)
	if err != nil {
		log.Println("[CreateMarshalDown]proto Marshal error",err)
		//panic(err)
		e := errors.Wrap(err,"[CreateMarshalDown]proto marshal error")
		return nil, e
	}

	aes_data,err :=encrypt(pData)
	if err != nil {
		log.Println("[CreateMarshalDown]Encrypt error",err)
		//panic(err)
		e := errors.Wrap(err,"[CreateMarshalDown]encrypt data error")
		return nil, e
	}
	return aes_data,nil
}


var key = []byte("B31F2A75FBF94099")

var iv = []byte("1234567890123456")

func encrypt(origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))

	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
	//return base64.StdEncoding.EncodeToString(crypted), nil
}
// 注意这里传入的是个base64加密的
func decrypt(crypted []byte) ([]byte, error) {
	//decodeData,err:=base64.StdEncoding.DecodeString(crypted)
	decodeData := crypted
	//if err != nil {
	//	return "",err
	//}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(decodeData))
	blockMode.CryptBlocks(origData, decodeData)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext) % blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length - 1])
	return origData[:(length - unpadding)]
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext) % blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length - 1])
	return origData[:(length - unpadding)]
}

