package main

// private -> 문서로 반환
// return은 public만

import (
	"bufio"
	"bytes"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
)

//DOCIndex의 번호를 저장하기 위한 변수
var docIndex int

//docIndex 변수를 초기화
func InitFunc() {
	docIndex = 1
}

func certify(args []string) MetaData {

	var data EOwner
	var sys SysInfo
	var meta MetaData

	data = EOwner{
		EOwnerKey:            args[7],
		EOwnerID:             args[0],
		EOwnerName:           args[1],
		EOwnerEmail:          args[2],
		EOwnerDepartMentName: args[3],
		EOwnerPosition:       args[4],
		EOwnerTelephone:      args[5]}
	// docNumber = len(data) - 1
	fmt.Println(data)
	if (data.EOwnerID == "") || (data.EOwnerName == "") || (data.EOwnerDepartMentName == "") || (data.EOwnerPosition == "") {
		meta.tf = false
		return meta
	}

	meta.tf = true
	meta.DocNum = docIndex
	docIndex++
	meta.EOwner = data
	meta.Sysinfo = sys

	return meta
}

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) []byte {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		fmt.Println(err)
	}
	return ciphertext
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	if err != nil {
		fmt.Println(err)
	}
	return plaintext
}

func createKey() {
	//GENERATE
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048) // 개인 키와 공개키 생성
	if err != nil {
		fmt.Println(err)
		return
	}

	// WRITE to pem file
	pemPrivateFile, err := os.Create("private_key.pem")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var pemPrivateBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	err = pem.Encode(pemPrivateFile, pemPrivateBlock)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pemPrivateFile.Close()
}

func loadKeyPri() *rsa.PrivateKey {
	privateKeyFile, err := os.Open("private_key.pem")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	size := pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)
	data, _ := pem.Decode([]byte(pembytes))
	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	return privateKeyImported
}

func encryptDoc(pubkey *rsa.PublicKey) []byte {
	s := "c-link"
	encryptedMsg := EncryptWithPublicKey([]byte(s), pubkey) // 공개키로 암호화

	return encryptedMsg
}

func decryptDoc(encryptedMsg []byte, prikey *rsa.PrivateKey) []byte {
	decryptedMsg := DecryptWithPrivateKey(encryptedMsg, prikey) // 개인키로 복호화

	return decryptedMsg
}

func checkDocHash(docHash, args []byte) bool {
	checkArgs := sha512.Sum512(args)
	return bytes.Equal(docHash, checkArgs[:])
}

func getInfo(args []string) DocMetadata {
	var sendingData DocMetadata
	var arrMeta MetaData
	arrMeta = certify(args)

	//만약 생성된 키가 없다면 키를 생성해준다.
	if arrMeta.EOwner.EOwnerKey == "false" {
		createKey()
	}

	prikey := loadKeyPri()
	pubkey := &prikey.PublicKey
	text := encryptDoc(pubkey)

	sendingData.DocTF = arrMeta.tf

	strNum := strconv.Itoa(arrMeta.DocNum)

	for i := len(strNum); i < 4; i++ {
		strNum = "0" + strNum
	}
	strNum = "DOC" + strNum
	sendingData.DocIndex = strNum

	//EOwner의 값은 한개의 byte배열로 받음
	sendingData.EOwner = []byte(fmt.Sprintf("%v", arrMeta.EOwner))

	//테스트 단계에서는 Clink와 test라는 임의의 값을 생성했다.
	sendingData.SysInfo = "Clink"
	sendingData.ClsScheme = "test"

	//SHA512를 사용, 문서 내용의 무결성을 보장한다.
	sendingData.DocHash = sha512.Sum512([]byte(args[6]))
	sendingData.encryptedC = text

	return sendingData
}
