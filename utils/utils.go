package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net"
	"os"
	"time"
)

// PrintError 错误输出
func PrintError(msg string) {
	fmt.Printf("\033[1;31m" + msg + " \033[0m\n")
}

// PrintSuccess 正确输出
func PrintSuccess(msg string) {
	fmt.Printf("\033[1;32m" + msg + " \033[0m\n")
}

// FormatYmdHis 返回格式：2021-08-05 00:00:01
func FormatYmdHis(timeObj time.Time) string {
	year := timeObj.Year()
	month := timeObj.Month()
	day := timeObj.Day()
	hour := timeObj.Hour()
	minute := timeObj.Minute()
	second := timeObj.Second()
	//注意：%02d 中的 2 表示宽度，如果整数不够 2 列就补上 0
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// Mkdir 创建目录
func Mkdir(path string, perm os.FileMode) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, perm)
		if err != nil {
			return
		}
		err = os.Chmod(path, perm)
		if err != nil {
			return
		}
	}
	return err
}

// RunDir 前面加上绝对路径
func RunDir(path string, a ...any) string {
	wd, _ := os.Getwd()
	if len(a) > 0 {
		path = fmt.Sprintf(path, a...)
	}
	return fmt.Sprintf("%s%s", wd, path)
}

// ReadFile 读取文件
func ReadFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// WriteFile 保存文件
func WriteFile(path string, content string) error {
	var fileByte = []byte(content)
	return os.WriteFile(path, fileByte, 0666)
}

// AppendToFile 追加文件
func AppendToFile(fileName string, content string) {
	// 以只写的模式，打开文件
	f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
	} else {
		// 查找文件末尾的偏移量
		n, _ := f.Seek(0, io.SeekEnd)
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt([]byte(content), n)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
}

// SliceInsert 向数组插入内容
func SliceInsert(s []string, index int, value string) []string {
	rear := append([]string{}, s[index:]...)
	return append(append(s[:index], value), rear...)
}

// FindIndex 查找数组位置
func FindIndex(tab []string, value string) int {
	for i, v := range tab {
		if v == value {
			return i
		}
	}
	return -1
}

// RandString 生成随机字符串
func RandString(len int) string {
	var r *rand.Rand
	r = rand.New(rand.NewSource(time.Now().Unix()))
	bs := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bs[i] = byte(b)
	}
	return string(bs)
}

// RandNum 生成随机数
func RandNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

// StringMd5 MD5
func StringMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// StringMd52 MD5
func StringMd52(str, pass string) string {
	text := fmt.Sprintf("%s%s", StringMd5(str), pass)
	return StringMd5(text)
}

func IpToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func IntToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

func StringToIP(i string) net.IP {
	return net.ParseIP(i).To4()
}

func Base64Encode(data string) string {
	sEnc := base64.RawURLEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf(sEnc)
}

func Base64Decode(data string) string {
	uDec, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return ""
	}
	return string(uDec)
}

// StringsContains 数组是否包含
func StringsContains(array []string, val string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}

// InArray 元素是否存在数组中
func InArray(item string, items []string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}
