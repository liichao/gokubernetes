package tools

import (
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

// 定义日志格式
var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

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
