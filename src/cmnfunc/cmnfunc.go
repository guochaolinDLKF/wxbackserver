package cmnfunc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"utils/log"
	"strings"
	"crypto/x509"
	"crypto/rsa"
)

type BaseJsonBean struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"msg"`
}

type JsonData map[string]interface{}

type ErrMsg struct {
	Msg string
	Dat interface{}
}

type CDATAText struct {
	Text string `xml:",innerxml"`
}

func Init() {
	reservedMap = map[byte]string{
		0x21: "%21", // !
		0x23: "%23", // #
		0x24: "%24", // $
		0x26: "%26", // &
		0x27: "%27", // '
		0x28: "%28", // (
		0x29: "%29", // )
		0x2a: "%2a", // *
		0x2b: "%2b", // +
		0x2c: "%2c", // ,
		0x2f: "%2f", // /
		0x3a: "%3a", // :
		0x3b: "%3b", // ;
		0x3d: "%3d", // =
		0x3f: "%3f", // ?
		0x40: "%40", // @
		0x5b: "%5b", // [
		0x5d: "%5d", // ]
	}

	for i := 0; i < 4; i++ {
		lh[i] = &loghelper{WHour: -1}
	}
	//日志路径，是否打印到控制台
	log.Init(Cfg["root"]+Cfg["logdir"], true)
}

func Value2CDATA(v string) CDATAText {
	return CDATAText{"<![CDATA[" + v + "]>"}
}

func NewBaseJsonBean() *BaseJsonBean {
	return &BaseJsonBean{}
}

func RspNull(w http.ResponseWriter) {
	fmt.Fprint(w, "")
}

func RspStr(w http.ResponseWriter, str *string) {
	fmt.Fprint(w, *str)
}

func RspData(w http.ResponseWriter, rsp interface{}) interface{} {
	fmt.Fprint(w, rsp)
	return rsp
}

func RspWithLogCmn(w http.ResponseWriter, rsp interface{}) {
	RspNull(w)
	log.Cmn(fmt.Sprintf("%v", rsp))
}

func RspWithLogErr(w http.ResponseWriter, rsp interface{}) {
	RspNull(w)
	log.Err(fmt.Sprintf("%v", rsp))
}

func RspWithJson(w http.ResponseWriter, code int, msg string, jdat interface{}) {
	result := NewBaseJsonBean()
	result.Code = code
	result.Message = msg
	result.Data = jdat
	bytes, _ := json.Marshal(result)
	fmt.Fprint(w, string(bytes))
}

func SendGet(req string, dat *string) bool {
	resp, err1 := http.Get(req)
	if err1 != nil {
		log.Err(fmt.Sprintf("SendJsonGet: 给小程序发送同步数据失败:%s",err1))
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body);
	if err!=nil{
		log.Err(fmt.Sprintf("SendJsonGet: 接受小程序发送同步数据读取失败:%s",err))
		return false
	}
	*dat=string(body)
	return true
}


func SendJsonGet(req string, dat *JsonData) bool {
	resp, err := http.Get(req)
	if err != nil {
		log.Err("SendJsonGet:")
		log.Err(err)
		return false
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Err("SendJsonGet:")
		log.Err(err)
		return false
	}
	log.Err("SendJsonGet:"+string(body))
	err = json.Unmarshal(body, dat)
	if err != nil {
		log.Err("SendJsonGet解析json失败")
		log.Err(err)
		return false
	}
	return true
}
func SendPost2(req string, postData string, dat *JsonData) bool {
	resp, err := http.Post(req, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		log.Err(err)
		return false
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Err(err)
		return false
	}

	err = json.Unmarshal(body, &dat)
	if err != nil {
		fmt.Println("json解析：",err)
		log.Err(err)
		return false
	}

	return true
}




type ecbDecrypter ecb

func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}
func (x *ecbDecrypter) BlockSize() int { return x.blockSize }
func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func AesDecrypt(crypted, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("err is:", err)
	}
	blockMode := NewECBDecrypter(block)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	//fmt.Println("source is :", origData, string(origData))
	return origData
}
func SendJsonPost(req string, postData []byte, dat *JsonData) bool {
	resp, err := http.Post(req, "application/x-www-form-urlencoded", bytes.NewBuffer(postData))
	if err != nil {
		log.Err(err)
		return false
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Err(err)
		return false
	}

	err = json.Unmarshal(body, dat)
	if err != nil {
		log.Err(err)
		return false
	}

	return true
}

func CheckSignature(signature string, timestamp string, nonce string, token string) bool {
	tmpArr := []string{token, timestamp, nonce}
	sort.Strings(tmpArr)
	tmpStr := tmpArr[0] + tmpArr[1] + tmpArr[2]
	h := sha1.New()
	io.WriteString(h, tmpStr)
	tmpStr = fmt.Sprintf("%x", h.Sum(nil))
	return tmpStr == signature
}

func CalcMsgSig(tok string, tmsp string, nonce string, encMsg *string) string {
	tmpArr := []string{tok, tmsp, nonce, *encMsg}
	sort.Strings(tmpArr)
	tmpStr := tmpArr[0] + tmpArr[1] + tmpArr[2] + tmpArr[3]
	h := sha1.New()
	io.WriteString(h, tmpStr)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ValidateSignature(tok string, tmsp string, nonce string, encMsg *string, msgSig string) bool {
	if len(tok) == 0 || len(nonce) == 0 || len(*encMsg) == 0 || len(tmsp) == 0 {
		return false
	}
	return CalcMsgSig(tok, tmsp, nonce, encMsg) == msgSig
}

func PKCS7Padding(encData []byte, blockSize int) []byte {
	padding := blockSize - len(encData)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(encData, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func DecryptMsg(encMsg *string, msgSig string,
	tmsp string, nonce string, tok string, ecodingAesKey string,
	appId string) ([]byte, error) {
	// 签名验证
	if !ValidateSignature(tok, tmsp, nonce, encMsg, msgSig) {
		return nil, errors.New("validate faild!")
	}
	// 计算AES KEY
	aesKey, err := base64.StdEncoding.DecodeString(ecodingAesKey + "=")
	if err != nil {
		return nil, err
	}
	// 由BASE64解密第一次
	var aesData []byte
	aesData, err = base64.StdEncoding.DecodeString(*encMsg)
	if err != nil {
		return nil, err
	}
	// 由AES解密第二次
	var block cipher.Block
	block, err = aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, aesKey[:blockSize])
	origData := make([]byte, len(aesData))
	blockMode.CryptBlocks(origData, aesData)
	origData = PKCS7UnPadding(origData)

	// 拆分数据
	var msgLen int32
	binary.Read(bytes.NewBuffer(origData[16:20]), binary.BigEndian, &msgLen)
	testAppId := string(origData[20+msgLen:])
	if testAppId != appId {
		return nil, errors.New("appId not match!")
	}
	return origData[20 : 20+msgLen], nil
}

func GenRedirectPage(url string) string {
	return fmt.Sprintf("<html><head><meta http-equiv='refresh' content='0;url=%s'/></head></html>", url)
}

func GenRedirectPageGet(url string, keys []string, vals []string) string {
	if len(keys) != len(vals) {
		return ""
	}
	htmlDat := fmt.Sprintf("<html><head><meta http-equiv='refresh' content='0;url=%s'", url)
	for i := 0; i < len(keys); i++ {
		if i == 0 {
			htmlDat += "?"
		} else {
			htmlDat += "&"
		}
		htmlDat += fmt.Sprintf("%s=%s", keys[i], vals[i])
	}
	htmlDat += "/></head></html>"
	return htmlDat
}

func GenRedirectPagePost(url string, keys []string, vals []string) string {
	if len(keys) != len(vals) {
		return ""
	}
	htmlDat := fmt.Sprintf("<html><head></head><body><form name='form1' action='%s'>", url)
	for i := 0; i < len(keys); i++ {
		htmlDat += fmt.Sprintf("<input type='hidden' name='%s' value='%s'/>", keys[i], vals[i])
	}
	htmlDat += "</form><script>form1.submit();</script></body></html>"
	return htmlDat
}
func GenRedirectPayPost(url string) string {
	htmlDat := fmt.Sprintf("<html><head></head><body><form name='form1' action='%s' method='post'>", url)
	htmlDat += "</form><script>form1.submit();</script></body></html>"
	return htmlDat
}

func RsaSignWithSha1Encrypt(data string, pubvKey string) string {
	//fmt.Println(pubvKey)
	pubStrByte, _ := base64.StdEncoding.DecodeString(pubvKey)
	publicKey, err := x509.ParsePKIXPublicKey([]byte(pubStrByte))
	if err != nil {
		fmt.Println(err)
	}
	sign, err1 := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), []byte(data))
	if err1 != nil {
		fmt.Println(err1)
	}
	signStr := base64.StdEncoding.EncodeToString(sign)
	return signStr
}





type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}
func (x *ecbEncrypter) BlockSize() int {
	return x.blockSize
}
func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

//明文补码算法
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//明文减码算法
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AES ECB模式的加密解密
type AesTool struct {
	//128 192  256位的其中一个 长度 对应分别是 16 24  32字节长度
	Key       []byte
	BlockSize int
}

func NewAesTool(key []byte, blockSize int) *AesTool {
	return &AesTool{Key: key, BlockSize: blockSize}
}

func (this *AesTool) padding(src []byte) []byte {
	//填充个数
	paddingCount := aes.BlockSize - len(src)%aes.BlockSize
	if paddingCount == 0 {
		return src
	} else {
		//填充数据
		return append(src, bytes.Repeat([]byte{byte(0)}, paddingCount)...)
	}
}

//unpadding
func (this *AesTool) unPadding(src []byte) []byte {
	for i := len(src) - 1; ; i-- {
		if src[i] != 0 {
			return src[:i+1]
		}
	}
	return nil
}

func (this *AesTool) Encrypt(src []byte) ([]byte, error) {
	//key只能是 16 24 32长度
	block, err := aes.NewCipher([]byte(this.Key))
	if err != nil {
		return nil, err
	}
	//padding
	src = this.padding(src)
	//返回加密结果
	encryptData := make([]byte, len(src))
	//存储每次加密的数据
	tmpData := make([]byte, this.BlockSize)

	//分组分块加密
	for index := 0; index < len(src); index += this.BlockSize {
		block.Encrypt(tmpData, src[index:index+this.BlockSize])
		copy(encryptData, tmpData)
	}
	return encryptData, nil
}
func (this *AesTool) Decrypt(src []byte) ([]byte, error) {
	//key只能是 16 24 32长度
	block, err := aes.NewCipher([]byte(this.Key))
	if err != nil {
		return nil, err
	}
	//返回加密结果
	decryptData := make([]byte, len(src))
	//存储每次加密的数据
	tmpData := make([]byte, this.BlockSize)

	//分组分块加密
	for index := 0; index < len(src); index += this.BlockSize {
		block.Decrypt(tmpData, src[index:index+this.BlockSize])
		copy(decryptData, tmpData)
	}
	return PKCS5UnPadding(decryptData), nil
}

func AesECBDecrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	ecb := NewECBDecrypter(block)
	retData := make([]byte, len(data))
	ecb.CryptBlocks(retData, data)
	// 解PKCS7填充
	retData = PKCS7UnPadding(retData)
	return retData, nil
}


func AesECBEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	ecb := NewECBEncrypter(block)
	// 加PKCS7填充
	content := PKCS7Padding(data, block.BlockSize())
	encryptData := make([]byte, len(content))
	// 生成加密数据
	ecb.CryptBlocks(encryptData, content)
	return encryptData, nil
}