package tools

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// 定义日志格式
var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// GetLastString 取切片最后一个值
func GetLastString(ss []string) string {
	return ss[len(ss)-1]
}

// GetIPString 获取ip前三个值
func GetIPString(str string) string {
	if len(strings.Split(str, `.`)) != 4 {
		return "ip input Error!"
	}
	return strings.Split(str, `.`)[0] + "." + strings.Split(str, `.`)[1] + "." + strings.Split(str, `.`)[2] + "."
}

// GetIPLastString 获取ip最后一个值
func GetIPLastString(str string) string {
	if len(strings.Split(str, `.`)) != 4 {
		return "ip input Error!"
	}
	return strings.Split(str, `.`)[3]
}

//replaceString将ip中的.更改为-并返回 node-1-1-1-1
func replaceString(str string) string {
	return "node-" + strings.ReplaceAll(str, ".", "-")
}

// GetIPDes 获取IP前面的地址
func GetIPDes(ips string) (string, int, int) {
	if len(strings.Split(ips, `.`)) != 4 {
		return "", 1, 1
	}
	if len(strings.Split(ips, `-`)) != 2 {
		result := strings.Split(ips, `.`)[0] + "." + strings.Split(ips, `.`)[1] + "." + strings.Split(ips, `.`)[2] + "."
		ipnum, err := strconv.Atoi(strings.Split(ips, `.`)[3])
		if err != nil {
			log.Info(err)
		}
		return result, ipnum, ipnum
	}
	result := strings.Split(ips, `.`)[0] + "." + strings.Split(ips, `.`)[1] + "." + strings.Split(ips, `.`)[2] + "."
	hostStartIP, err := strconv.Atoi(strings.Split(strings.Split(ips, ".")[len(strings.Split(ips, "."))-1], `-`)[0])
	if err != nil {
		log.Info(err)
	}
	hostStopIP, err := strconv.Atoi(strings.Split(strings.Split(ips, ".")[len(strings.Split(ips, "."))-1], `-`)[1])
	if err != nil {
		log.Info(err)
	}
	return result, hostStartIP, hostStopIP
}

// CheckCreateWriteFile 检查文件并创建，在写入内容
func CheckCreateWriteFile(filePath, fileName, content string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		log.Warning(err)
		log.Warning("start create " + filePath + "...")
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	}
	log.Warning("start write " + fileName + "...")
	f, err3 := os.Create(filePath + fileName) //创建文件
	if err3 != nil {
		log.Warning(err3)
		log.Warning("create file fail," + fileName + "file is exist")
	}
	w := bufio.NewWriter(f) //创建新的 Writer 对象
	n4, err3 := w.WriteString(content)
	if err3 != nil {
		log.Warning(err3)
	}
	log.Info("写入 " + strconv.Itoa(n4) + " 字节成功.")
	w.Flush()
	f.Close()
	return false
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

// IsContain 判断是否存在数组中
// func IsContain(items []string, item string) bool {
// 	for _, eachItem := range items {
// 		if eachItem == item {
// 			return true
// 		}
// 	}
// 	return false
// }
// ShellToUse 定义shell使用bash
const ShellToUse = "bash"

// ShellOut 执行命令并返回结果
func ShellOut(command string) bool {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return false
	}
	// log.Info(stdout.String())
	// log.Error(stderr.String())
	return true
}
