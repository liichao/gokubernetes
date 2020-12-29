package tools

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/op/go-logging"
	"github.com/pytool/ssh"
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
		// log.Warning(err)
		// log.Warning("start create " + filePath + "...")
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	}
	// log.Warning("start write " + fileName + "...")
	f, err3 := os.Create(filePath + fileName) //创建文件
	if err3 != nil {
		log.Warning(err3)
		log.Warning("create file fail," + fileName + "file is exist")
	}
	w := bufio.NewWriter(f) //创建新的 Writer 对象
	_, err3 = w.WriteString(content)
	if err3 != nil {
		log.Warning(err3)
	}
	// log.Info("写入 " + strconv.Itoa(n4) + " 字节成功.")
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
	log.Error(stderr.String())
	return true
}

// LoadImagesChangeTagPushImages 载入镜像修改镜像tag 推送到仓库
/*
ip 执行这个的服务器ip

pwd 密码

k8spath 程序释放目录

harborURL  仓库url

harborUser  仓库账户

harborPwd  仓库密码

imagesOfflineTar docker镜像tar包

ImagesTag 模糊查询关键字，通过该关键字定位到docker镜像并生产新仓库的镜像images
*/
func LoadImagesChangeTagPushImages(ip, pwd, k8spath, harborURL, harborUser, harborPwd, imagesOfflineTar, harborImagesURL, ImagesTag string) bool {
	// 判断pause镜像文件是否存在
	if !Exists(k8spath + "tools/" + imagesOfflineTar) {
		log.Info(imagesOfflineTar + "镜像包不存在，请上传到" + k8spath + "tools/目录下")
	}
	// 将镜像包上传到nodeList的第一台之后载入并推送到docker仓库
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// 上传metricsServerImages tar到node的第一台机器
	err = c.Upload(k8spath+"tools/"+imagesOfflineTar, "/tmp/")
	if err != nil {
		log.Info(err)
		log.Error(" 上传" + imagesOfflineTar + "镜像失败")
		return false
	}
	// 载入镜像
	err = c.Exec("docker load -i /tmp/" + imagesOfflineTar)
	if err != nil {
		log.Error(err)
		log.Error(" 载入" + imagesOfflineTar + "镜像失败")
		return false
	}
	// 更改镜像tag
	shell := "docker tag $(docker images |grep " + ImagesTag + " |head -n 1 |awk '{print $3}')" + " " + harborImagesURL
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
		return false
	}
	// 与仓库建立连接
	shell = "docker login -u " + harborUser + " -p " + harborPwd + " " + harborURL
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	//推送到仓库
	err = c.Exec("docker push " + harborImagesURL)
	if err != nil {
		log.Error(err)
	}
	return true
}

// SendBinAndConfigFile 分发指定文件到指定IP上的指定目录上
/*
ip  执行这个的服务器ip
pwd 密码
sourcePath 源文件
descPath 目的地址 写死/opt/kubernetes/bin/
*/
func SendBinAndConfigFile(ip, pwd, sourcePath, descPath string, listFile []string, ws *sync.WaitGroup) {
	log.Info(ip + pwd + sourcePath + descPath)
	defer ws.Done()
	// 判断源文件是否存在
	if !Exists(sourcePath) {
		log.Info(sourcePath + "不存在，请上传")
	}
	// 开启远程登陆
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// 尝试创建目录 忽略所有报错
	c.MkdirAll(descPath)
	// 上传，并返回结果
	// 上传metricsServerImages tar到node的第一台机器
	for _, file := range listFile {
		err = c.Upload(sourcePath+"tools/"+file, descPath)
		if err != nil {
			log.Info(err)
			log.Error(" 上传" + sourcePath + "tools/" + file + "失败")

		}
	}
	// 赋权
	err = c.Exec("chmod u+x " + descPath + "/*")
	if err != nil {
		log.Error(err)
		log.Error(descPath + "授权失败")
	}
	// 检查文件是否存在
	c.Exec("ls -l " + descPath)
	log.Info(ip + "所有文件都已经上传完成了.可以去/opt/kubernetes/bin目录下查看下")
}
