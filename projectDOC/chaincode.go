/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type Chaincode struct {
}

func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Init()", fcn, params)
	InitFunc()
	return shim.Success(nil)
}

//chaincode update시에 DOCIndex를 다시 1로 변경하는 것을 막기 위함
func (cc *Chaincode) InitUpgrade(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Init()", fcn, params)
	return shim.Success(nil)
}

func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + fcn)

	switch fcn {
	case "createDocument":
		return cc.createDocument(stub, params)
	case "deleteDocument":
		return cc.deleteDocument(stub, params)
	case "checkIntegrity":
		return cc.checkIntegrity(stub, params)
	case "queryAllDocs":
		return cc.queryAllDocs(stub)
	case "queryDoc":
		return cc.queryDoc(stub, params)
	default:
		fmt.Println("invoke did not find func: " + fcn)
		return shim.Error("Received unknown function invocation")
	}
}

//문서의 metadata 생성
func (cc *Chaincode) createDocument(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//EOwnerID : args[0], EOwnerName: args[1], EOwnerEmail : args[2],	EOwnerDepartMentName: args[3],
	//EOwnerPosition : args[4],	EOwnerTelephone : args[5], DOCString : args[6], EOwnerKey(키 소유여부) : args[7]
	//test args == ["IDstudent01","jhyun","happy@cau","c-link","student","01012345678", "hello this is DOC string", "true"]

	var studentData DocMetadata

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	//json에서 정보를 받아온다
	studentData = getInfo(args)

	if studentData.DocTF == false {
		return shim.Error("invalid studentData")
	}

	studentData.DocTimeStmp = time.Now().Format("2006-01-02 15:04:05")
	studentData.DocStatus = true

	key := studentData.DocIndex

	metaDataAsBytes, _ := json.Marshal(studentData)

	stub.PutState(key, metaDataAsBytes)

	return shim.Success(nil)
}

//문서 파기시 데이터의 상태(Docstatus)를 false로 바꿈
func (cc *Chaincode) deleteDocument(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//docIndex : args[0]
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	metaDataAsBytes, err := stub.GetState(args[0])
	var meta DocMetadata

	if err != nil {
		return shim.Error("failed to GetState")
	}

	json.Unmarshal(metaDataAsBytes, &meta)
	meta.DocStatus = false

	metaDataAsBytes, _ = json.Marshal(meta)
	stub.PutState(args[0], metaDataAsBytes)
	fmt.Println("succcess to change status")

	return shim.Success(nil)
}

//문서의 무결성을 증명하기 위해 사용
func (cc *Chaincode) checkIntegrity(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//DocID : args[0], 문서의 원본 내용 : args[1]
	//test args == ["DOC0001", "hello this is DOC string"]

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	metaDataAsBytes, err := stub.GetState(args[0])
	var metaTmp DocMetadata

	if err != nil {
		return shim.Error("failed to GetState")
	}

	json.Unmarshal(metaDataAsBytes, &metaTmp)
	checkIDDummy := metaTmp.encryptedC

	//개인키를 불러온 후 검증
	key := loadKeyPri()
	docBytes := decryptDoc(checkIDDummy, key)

	if bytes.Equal(docBytes, []byte("c-link")) {
		return shim.Error("invalid Private key")
	}

	//원본증명
	if checkDocHash(metaTmp.DocHash[:], []byte(args[1])) == true {
		fmt.Println("true")

		//같은 문서라는 것을 확인했을 경우, 확인한 시간을 기록한다.
		metaTmp.CheckTimeStmp = time.Now().Format("2006-01-02 15:04:05")
		metaDataAsBytes, _ = json.Marshal(metaTmp)
		stub.PutState(args[0], metaDataAsBytes)

		return shim.Success(nil)
	}
	fmt.Println("false")
	return shim.Error("false")

}

//모든 문서의 metadata 값을 로드함
func (cc *Chaincode) queryAllDocs(stub shim.ChaincodeStubInterface) sc.Response {
	startKey := "DOC0"
	endKey := "DOC9999"

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemeberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemeberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")

		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemeberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("== queryAll:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//DOC ID로 query
func (cc *Chaincode) queryDoc(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//DocIndex : args[0]
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	metaDataAsBytes, _ := stub.GetState(args[0])

	return shim.Success(metaDataAsBytes)
}
