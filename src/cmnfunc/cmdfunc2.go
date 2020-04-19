package cmnfunc

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

type loghelper struct {
	WHour   int         // 当前正在写的小时数
	LogDir  string      // 日志文件目录
	LogFile *os.File    // 日志文件
	Logger  *log.Logger // logger
}

var Cfg = map[string]string{}

var AchievCfg = make(map[string][]map[string]interface{}, 10)

var VipShopCfg = make(map[string][]map[string]interface{}, 10)
var TaskListCfg = make(map[string][]map[string]interface{}, 10)
var SignInListCfg = make(map[string][]map[string]interface{}, 10)
var PrizeListCfg = make(map[string][]map[string]interface{}, 10)
var VipInfoListCfg = make(map[string][]map[string]interface{}, 10)
var GunInfoListCfg = make(map[string][]map[string]interface{}, 10)
var WeakDayPrizeInfoListCfg = make(map[string][]map[string]interface{}, 10)

var reservedMap map[byte]string
var lh [4]*loghelper

func checkEncodeChar(s string) (int, []int) {
	c := 0
	m := len(s)
	ia := make([]int, m)
	for i := 0; i < m; i++ {
		if _, ok := reservedMap[s[i]]; ok {
			ia[c] = i
			c++
		}
	}
	return c, ia[0:c]
}

func RandBytes(num int) []byte {
	randData := make([]byte, num)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < num; i++ {
		n := 33 + r.Intn(94)
		randData[i] = (byte)(n)
	}
	return randData
}

func RandBytes2(num int) []byte {
	randData := make([]byte, num)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < num; i++ {
		n := r.Intn(62)
		switch {
		case n >= 0 && n <= 9:
			n += 48 // 0-9
		case n >= 10 && n <= 35:
			n += 55 // A-Z
		case n >= 36 && n <= 61:
			n += 61 // a-z
		default:
			n = 0
		}
		randData[i] = (byte)(n)
	}
	return randData
}

// 生成指定长度的session id n为字符个数
func GenSessionId(n int) []byte {
	h := []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	lnh := len(h)
	if lnh >= n {
		return h[lnh-n:]
	}
	s := make([]byte, n)
	copy(s, h)
	copy(s[lnh:], RandBytes2(n-lnh))
	return s
}
func Md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// URL encode
func URLEncode(u string) string {
	c, ia := checkEncodeChar(u)
	if c == 0 {
		return u
	}
	nu := make([]byte, len(u)+c*2)
	for i, j, k := 0, 0, 0; i < len(u); i++ {
		if k < c && ia[k] == i {
			copy(nu[j:j+3], reservedMap[u[i]])
			j += 3
			k++
		} else {
			nu[j] = u[i]
			j++
		}
	}
	return string(nu)
}

//生成指定长度的随机字符串
func CreateCaptcha(l int) string {
	str := "123456789"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//生成指定长度的随机字符串
func CreateStrCaptcha(l int) string {
	str := "123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
