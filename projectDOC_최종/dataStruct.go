package main

//EOwner 는 보관자의 정보를 나타낸다
type EOwner struct {
	EOwnerKey            string		//보관자 key 유무
	EOwnerID             string		//보관자 id
	EOwnerName           string		//보관자명
	EOwnerEmail          string		//전자메일
	EOwnerDepartMentName string		//부서명
	EOwnerPosition       string		//직취명
	EOwnerTelephone      string		//전화번호
}

//SysInfo 시스템정보를 나타낸다
type SysInfo struct {
	// 전자화 문서 작성 일시
	ESDocCreateDate string
	// 전자화 문서 정보
	ESDocCreateVersion int
	ESDocType          string
	// 전자화 프로그램
	ESProgOS      string
	ESProgAppName string
	ESProgVersion int
	// 전자화 장비 (스캐너)
	EDMaker         string
	EDProductName   string
	EDProductSerial string
	EDProductIP     string
	//타임스탬프
	TimeStmpToken  string
	TimeStmpIssuer string
	TimeStmpDate   string
	//해쉬
	HashValue     []byte
	HashAlgorithm string
}

//MetaData
type MetaData struct {
	tf      bool
	DocNum  int
	EOwner  EOwner
	Sysinfo SysInfo
}

//DocMetadata 블록체인에 담기는 구조체
type DocMetadata struct {
	//인풋 데이터 유효성 검증
	DocTF bool
	//문서 인덱스 (key값)
	DocIndex string //DocNum
	//보관자 정보
	EOwner []byte //한값으로
	//요청받은 시스템 정보
	SysInfo string //"Clink" 값으로 대체
	//문서 분류정보
	ClsScheme string //"test"
	//문서의 해시값
	DocHash [64]byte
	//블록체인에 담기는 시점
	DocTimeStmp string
	//현재 파기 여부
	DocStatus bool
	//암호화 데이터 - 개인키 인증을 위해 필요
	encryptedC []byte
	//열람했다는 데이터
	CheckTimeStmp string
}
