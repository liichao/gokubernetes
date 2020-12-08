package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// EtcdJSONParse etcd json 解析
type EtcdJSONParse struct {
	CN    string   `json:"CN"`
	Hosts []string `json:"hosts"`
	Key   struct {
		Algo string `json:"algo"`
		Size int    `json:"size"`
	} `json:"key"`
	Names []struct {
		C  string `json:"C"`
		ST string `json:"ST"`
		L  string `json:"L"`
		O  string `json:"O"`
		OU string `json:"OU"`
	} `json:"names"`
}

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
	log.Info(stdout.String())
	log.Error(stderr.String())
	return true
}

// PrintHelpInfo 输出帮助信息
func PrintHelpInfo() {
	// 输出帮助命令
	fmt.Printf("Hi Welcome to use One Key Install kubernests \n")
	fmt.Printf("Usage: go-install-kubernetes [PARATEMERS]...\n")
	fmt.Printf(" \n")
	fmt.Printf("  First parameter\n")
	fmt.Printf("    Example: go-install-kubernetes system 172.16.0.1-254 123456 ipvs ntpserverip \n")
	fmt.Printf("    - system: prepare system environment and install vim,net-tools etc \n")
	fmt.Printf("    - chrony: remove ntp serever install chrony and change the chrony configuration\n")
	fmt.Printf("    - cert:   create k8s and etcd cert file save to /tmp/k8s/cert/\n")
	fmt.Printf("  Two parameter\n")
	fmt.Printf("    - IP addresses        172.16.0.1-10 \n")
	fmt.Printf("  Three parameter\n")
	fmt.Printf("    - password            system password \n")
	fmt.Printf("  Four parameter\n")
	fmt.Printf("    - ipvs:               proxy mode \n")
	fmt.Printf("  Five parameter\n")
	fmt.Printf("    - NTP server IP:      ntp.aliyun.com \n")
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
	log.Info("Write " + strconv.Itoa(n4) + " bytes successed")
	w.Flush()
	f.Close()
	return false
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

// ConfigSystem 配置系统参数
func ConfigSystem(ip, pwd, proxymode string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// 开始更新系统版本与安装相关必要组件
	log.Info("Update system and install other tools commond")
	c.Exec("yum update -y")
	err = c.Exec("yum install vim net-tools wget bash-completion conntrack-tools ipset ipvsadm libseccomp nfs-utils psmisc rsync socat -y")
	if err != nil {
		log.Info(err)
	}
	// 禁用swap
	log.Info(ip + " 开始禁用系统Swap")
	swapoutput, err := c.Output("swapoff -a && sysctl -w vm.swappiness=0")
	if err != nil {
		panic(err)
	}
	log.Info(ip + " " + string(swapoutput[:]) + "禁用系统Swap完成")
	log.Info(ip + " 开始删除fstab swap 相关配置")
	fstaboutput, err := c.Output("sed -i '/swap/d' /etc/fstab")
	if err != nil {
		panic(err)
	}
	log.Info(string(fstaboutput[:]) + "删除fstab swap 相关配置完成")
	// 删除防火墙
	log.Info(ip + "stop remove firewalld")
	err = c.Exec("yum remove firewalld python-firewall firewalld-filesystem -y")
	if err != nil {
		log.Info(err)
	}
	// 关闭selinux
	log.Info(ip + "trunoff selinux")
	err = c.Exec("setenforce 0")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/sysconfig/selinux")
	if err != nil {
		log.Info(err)
	}
	// 优化rsyslog获取journald日志
	log.Info(ip + " optimize journald log")
	err = c.Exec("sed -i 's/$ModLoad imjournal/#$ModLoad imjournal/g' /etc/rsyslog.conf")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("sed -i 's/$IMJournalStateFile/#$IMJournalStateFile/g' /etc/rsyslog.conf")
	if err != nil {
		log.Info(err)
	}
	log.Info(ip + " restart journald service")
	err = c.Exec("systemctl restart ")
	if err != nil {
		log.Info(err)
	}
	// 加载内核模块
	log.Info(ip + "获取内核系统版本,并根据不同版本加载内核模块")
	versionoutput, err := c.Output("uname -a")
	if err != nil {
		panic(err)
	}
	sysversion := strings.Split(strings.Split(string(versionoutput[:]), "-")[1], " ")[1]
	log.Info(ip + " 系统版本Version:" + sysversion)
	sysversionint, err := strconv.Atoi(strings.Split(sysversion, ".")[0])
	if err != nil {
		log.Error("字符串转换成整数失败")
	}
	// 系统内核版本小于4.19的 nf_conntrack模块名为：nf_conntrack_ipv4 大于4.19的名为nf_conntrack_ipv4
	var nfConntrack string
	var k8sSysctl bool
	if sysversionint == 3 {
		nfConntrack = "nf_conntrack_ipv4"
		k8sSysctl = true
	} else {
		nfConntrack = "nf_conntrack"
		k8sSysctl = false
	}
	log.Info(ip + "nf_conntrack模块名为:" + nfConntrack)
	KernelModule := []string{"br_netfilter", "ip_vs", "ip_vs_rr", "ip_vs_wrr", "ip_vs_sh", nfConntrack}
	for _, d := range KernelModule {
		log.Info("Load system kernel module:" + d)
		err = c.Exec("modprobe " + d)
		if err != nil {
			log.Info(err)
		}
		err = c.Exec("lsmod |grep " + d)
		if err != nil {
			log.Info(err)
		}
	}
	// 将10-k8s-modules.conf放到服务器的指定目录/etc/modules-load.d/
	log.Info(ip + " scp 10-k8s-modules.conf to /etc/modules-load.d/")
	err = c.Upload("/tmp/k8s/10-k8s-modules.conf", "/etc/modules-load.d/")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("echo " + nfConntrack + ">> /etc/modules-load.d/10-k8s-modules.conf")
	if err != nil {
		log.Info(err)
	}
	// 将95-k8s-sysctl.conf放到服务器的指定目录/etc/sysctl.d
	log.Info(ip + " scp 95-k8s-sysctl.conf to /etc/sysctl.d/")
	err = c.Upload("/tmp/k8s/95-k8s-sysctl.conf", "/etc/sysctl.d/")
	if err != nil {
		log.Info(err)
	}
	if k8sSysctl {
		err = c.Exec("echo 'net.ipv4.tcp_tw_recycle = 0'>> /etc/sysctl.d/95-k8s-sysctl.conf")
		if err != nil {
			log.Info(err)
		}
	}
	if strings.EqualFold(proxymode, "ipvs") {
		err = c.Exec("echo 'net.ipv4.tcp_keepalive_time = 600\nnet.ipv4.tcp_keepalive_intvl = 30\nnet.ipv4.tcp_keepalive_probes = 10'>> /etc/sysctl.d/95-k8s-sysctl.conf")
		if err != nil {
			log.Info(err)
		}
	}
	// 将系统参数生效
	log.Info("sysctl -p /etc/sysctl.d/95-k8s-sysctl.conf")
	err = c.Exec("sysctl -p /etc/sysctl.d/95-k8s-sysctl.conf")
	if err != nil {
		log.Info(err)
	}
	// 启动 systemd-modules-load 服务并配置开机启动
	err = c.Exec("systemctl restart systemd-modules-load && systemctl enable systemd-modules-load")
	if err != nil {
		log.Info(err)
	}
	// 创建 systemd 配置目录
	err = c.Exec("mkdir -p /etc/systemd/system.conf.d")
	if err != nil {
		log.Error(err)
	}
	log.Info(ip + " scp 30-k8s-ulimits.conf to /etc/systemd/system.conf.d/")
	err = c.Upload("/tmp/k8s/30-k8s-ulimits.conf", "/etc/systemd/system.conf.d/")
	if err != nil {
		log.Info(err)
	}
	// 把SCTP列入内核模块黑名单
	err = c.Exec("mkdir -p /etc/systemd/system.conf.d")
	if err != nil {
		log.Error(err)
	}
	err = c.Upload("/tmp/k8s/sctp.conf", "/etc/modprobe.d/")
	if err != nil {
		log.Info(err)
	}
}

// InstallChrony 安装时钟同步服务器
func InstallChrony(ip, pwd, ntpserver string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// 卸载ntp 并安装 chrony
	log.Info(ip + " remove ntp server and install chrony")
	err = c.Exec("yum remove -y ntp && yum install -y chrony")
	if err != nil {
		log.Error(err)
	}
	err = c.Upload("/tmp/k8s/server-centos.conf", "/etc/chrony.conf")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("sed -i 's/NTPSERVER/" + ntpserver + "/g' /etc/chrony.conf")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("systemctl restart chronyd && systemctl enable chronyd")
	if err != nil {
		log.Info(err)
	}
}

// CreateCert 创建相关证书
func CreateCert(k8spath string) {
	// 	# kubeconfig 配置参数，注意权限根据‘USER_NAME’设置：
	// # 'admin' 表示创建集群管理员（所有）权限的 kubeconfig
	// # 'read' 表示创建只读权限的 kubeconfig
	// CLUSTER_NAME:
	// 	"cluster1"
	// USER_NAME:
	// 	"admin"
	// CONTEXT_NAME:
	// 	"context-{{ CLUSTER_NAME }}-{{ USER_NAME }}"
	userName := "admin"
	clusterName := "cluster1"
	contextName := "context-" + userName + "-" + clusterName
	kubeAPIServer := "https://10.10.77.202:6443"
	log.Info(contextName)
	log.Info("create k8s and etcd cert...")
	// 卸载ntp 并安装 chrony
	log.Info("chmod +x cfssl cfssljson ")
	if !ShellOut("chmod +x " + k8spath + "tools/*") {
		log.Error("chmod error please check cfssl and cfssljson exist")
	}
	// 生成 CA 证书和私钥
	//ShellOut("cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -initca ca-csr.json |  " + k8spath + "tools/cfssljson -bare ca")
	shell := "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -initca ca-csr.json |  " + k8spath + "tools/cfssljson -bare ca"
	log.Info(shell)
	if !ShellOut(shell) {
		log.Error("create CA cert private key faild!!!")
	}
	// 创建admin证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes " + userName + "-csr.json |  " + k8spath + "tools/cfssljson -bare " + userName
	log.Info(shell)
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 删除原有kubeconfig
	shell = "rm -rf /root/.kube/config"
	log.Info("start run " + shell)
	if !ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}
	// 设置集群参数admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/hyperkube1.16.15 kubectl config set-cluster " + clusterName + " --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer
	log.Info("start run " + shell)
	if !ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}
	// 设置客户端认证参数admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/hyperkube1.16.15 kubectl config set-credentials " + userName + " --client-certificate=" + k8spath + "cert/" + userName + ".pem --embed-certs=true --client-key=" + k8spath + "cert/" + userName + "-key.pem"
	log.Info("start run " + shell)
	if !ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}
	// 设置上下文参数admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/hyperkube1.16.15 kubectl config set-context " + contextName + " --cluster=" + clusterName + " --user=" + userName
	log.Info("start run " + shell)
	if !ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}
	// 选择默认上下文admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/hyperkube1.16.15 kubectl config set-context " + contextName
	log.Info("start run " + shell)
	if !ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}

	// 创建 kube-scheduler证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-scheduler-csr.json | " + k8spath + "tools/cfssljson -bare kube-scheduler"
	log.Info("start run " + shell)
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置集群参数kube-scheduler
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-cluster kubernetes --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer + "--kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置认证参数kube-scheduler
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-credentials system:kube-scheduler --client-certificate=" + k8spath + "cert/kube-scheduler.pem --client-key=" + k8spath + "cert/kube-scheduler-key.pem --embed-certs=true --server=" + kubeAPIServer + "--kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置上下文参数kube-scheduler
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-context default --cluster=kubernetes --user=system:kube-scheduler --kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 选择默认上下文kube-scheduler
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config use-context default --kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}

	//创建 kube-proxy证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-proxy-csr.json | " + k8spath + "tools/cfssljson -bare kube-proxy"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置集群参数kube-proxy
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-cluster kubernetes --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer + "--kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置客户端认证参数kube-proxy
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-credentials kube-proxy --client-certificate=" + k8spath + "cert/kube-proxy.pem --client-key=" + k8spath + "cert/kube-proxy-key.pem --embed-certs=true --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置上下文参数kube-proxy
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-context default --cluster=kubernetes --user=kube-proxy --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 选择默认上下文kube-proxy
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config use-context default --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}

	// 创建 kube-controller-manager证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-controller-manager-csr.json | " + k8spath + "tools/cfssljson -bare kube-controller-manager"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置集群参数kube-controller-manager
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-cluster kubernetes --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer + "--kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置认证参数kube-controller-manager
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-credentials system:kube-controller-manager --client-certificate=" + k8spath + "cert/kube-controller-manager.pem --client-key=" + k8spath + "cert/kube-controller-manager-key.pem --embed-certs=true --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置上下文参数kube-controller-manager
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config set-context default --cluster=kubernetes --user=system:kube-controller-manager --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 选择默认上下文kube-controller-manager
	shell = k8spath + "tools/hyperkube1.16.15 kubectl config use-context default --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
}

// InstallEtcd 安装etcd集群
func InstallEtcd(ip, pwd, k8spath, etcdID, etcdnode string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// create etcd path
	if !c.IsExist("/etc/etcd/ssl") {
		err = c.MkdirAll("/etc/etcd/ssl")
		if err != nil {
			log.Error(err)
		}
	}
	if !c.IsExist("/opt/kubernetes/bin/") {
		err = c.MkdirAll("/opt/kubernetes/bin/")
		if err != nil {
			log.Error(err)
		}
	}
	// create etcd data path
	if !c.IsExist("/var/lib/etcd/") {
		err = c.MkdirAll("/var/lib/etcd/")
		if err != nil {
			log.Error(err)
		}
	}
	// scp etcd to desc host
	err = c.Upload(k8spath+"tools/etcd/etcd", "/opt/kubernetes/bin/etcd")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"tools/etcd/etcdctl", "/opt/kubernetes/bin/etcdctl")
	if err != nil {
		log.Info(err)
	}
	// scp cert to desc host
	err = c.Upload(k8spath+"cert/etcd.pem", "/etc/etcd/ssl/")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"cert/etcd-key.pem", "/etc/etcd/ssl/")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"cert/ca.pem", "/etc/etcd/ssl/")
	if err != nil {
		log.Info(err)
	}
	// chmod bin
	err = c.Exec("chmod 755 /opt/kubernetes/bin/*")
	if err != nil {
		log.Error(err)
	}
	// scp etcd.service to desc hosts
	err = c.Upload(k8spath+"service/etcd.service", "/etc/systemd/system/")
	if err != nil {
		log.Info(err)
	}
	// modify the etcd.service config file
	log.Info(etcdID)
	err = c.Exec("sed -i 's/NODE_NAME/etcd" + etcdID + "/g' /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%ETCD_NODES%" + etcdnode + "%g' /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	// config etcd service enable
	err = c.Exec("systemctl enable etcd")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl daemon-reload")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl restart etcd")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl status etcd.service")
	if err != nil {
		log.Error(err)
	}
}

/*
	参数1: 系统配置修改
	参数2: 服务器ip或者网段
	参数3: 服务器密码
	参数4: 启用ipvs
	参数4: 时间同步服务器
*/
func main() {
	//para := make(map[string]string)
	// https://storage.googleapis.com/kubernetes-release/release/v1.16.15/bin/linux/amd64/hyperkube 更改版本号以下载
	// 创建并发
	var wg sync.WaitGroup
	para := make(map[string]interface{})
	// ParaList := []string{"system", "chrony", "kubernetes", "createcert"}
	confLists := []string{"10-k8s-modules.conf", "95-k8s-sysctl.conf", "30-k8s-ulimits.conf", "sctp.conf", "server-centos.conf"}
	certFileLists := []string{"etcd-csr.json", "admin-csr.json", "ca-config.json", "ca-csr.json", "kube-controller-manager-csr.json", "kube-proxy-csr.json", "kube-scheduler-csr.json", "read-csr.json"}
	//toolsFileLists := []string{"cfssl", "cfssljson", "hyperkube1.16.15"}
	toolsFileLists := []string{"cfssl", "cfssljson", "etcd.tar.gz"}
	yamlFileLists := []string{"read-group-rbac.yaml"}
	serviceFileLists := []string{"etcd.service"}
	k8spath := "/tmp/k8s/"
	// 获取命令行参数,并检查参数是否存在
	if len(os.Args) < 2 {
		PrintHelpInfo()
		os.Exit(0)
	}
	// 获取参数
	for _, value := range os.Args {
		if !strings.Contains(value, `=`) {
			// todo print err msg
			continue
		}
		prameters := strings.Split(value, `=`)
		if len(prameters) != 2 {
			// todo print err msg
			continue
		}
		para[prameters[0]] = prameters[1]
	}
	// 将配置文件生成到/tmp/k8s目录中
	log.Info(" Load Kubernetes config file ...")
	for _, file := range confLists {
		filebytes, err := Asset("config/" + file)
		if err != nil {
			panic(err)
		}
		CheckCreateWriteFile(k8spath, file, string(filebytes))
	}
	// 将证书释放到相关目录
	for _, file := range certFileLists {
		filebytes, err := Asset("config/cert/" + file)
		if err != nil {
			panic(err)
		}
		CheckCreateWriteFile(k8spath+"cert/", file, string(filebytes))
	}
	// 将tools文件释放到相关目录
	for _, file := range toolsFileLists {
		filebytes, err := Asset("config/tools/" + file)
		if err != nil {
			panic(err)
		}
		CheckCreateWriteFile(k8spath+"tools/", file, string(filebytes))
	}
	// 将yaml文件释放到相关目录
	for _, file := range yamlFileLists {
		filebytes, err := Asset("config/yaml/" + file)
		if err != nil {
			panic(err)
		}
		CheckCreateWriteFile(k8spath+"yaml/", file, string(filebytes))
	}
	// 将service文件释放到相关目录
	for _, file := range serviceFileLists {
		filebytes, err := Asset("config/service/" + file)
		if err != nil {
			panic(err)
		}
		CheckCreateWriteFile(k8spath+"service/", file, string(filebytes))
	}
	//var hostIp string
	hostIPSplit, hostStartIP, hostStopIP := GetIPDes(para[`ips`].(string))
	// start theard
	threadNum := hostStopIP - hostStartIP + 1
	wg.Add(threadNum)
	switch para[`para`] {
	case `system`:
		log.Info(" Start config system parameters...")
		for ; hostStartIP <= hostStopIP; hostStartIP++ {
			log.Info(hostStartIP)
			//ConfigSystem(para[`ips`].(string), para[`pwd`].(string), para[`proxymode`].(string), &wg)
			hostStartIPstr := strconv.Itoa(hostStartIP)
			go ConfigSystem(hostIPSplit+hostStartIPstr, para[`pwd`].(string), para[`proxymode`].(string), &wg)
		}
		wg.Wait()
		log.Info("all system parameters config done.")
	case `chrony`:
		log.Info(" Start install chrony...")
		// InstallChrony(para[`ips`].(string), para[`pwd`].(string), para[`ntpserver`].(string))
		for ; hostStartIP <= hostStopIP; hostStartIP++ {
			log.Info(hostStartIP)
			hostStartIPstr := strconv.Itoa(hostStartIP)
			go InstallChrony(hostIPSplit+hostStartIPstr, para[`pwd`].(string), para[`ntpserver`].(string), &wg)
		}
		wg.Wait()
		log.Info("all chrony server config done.")
	case `createcert`:
		log.Info(" Start install chrony...")
		CreateCert(k8spath)
	case `etcd`:
		var etcd EtcdJSONParse
		log.Info("Start Install Etcd....")
		// unzip etcd.tar.gz
		shell := "cd " + k8spath + "tools/ && tar -zxf etcd.tar.gz && mv " + k8spath + "tools/etcd-* " + k8spath + "tools/etcd"
		log.Info("start run " + shell)
		if !ShellOut(shell) {
			log.Error("unzip etcd.tar.gz faild!!!")
		}
		// load etcd-csr.json
		jsonFile, _ := os.Open("/tmp/k8s/cert/etcd-csr.json")
		defer jsonFile.Close()
		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			log.Error(err)
		}
		json.Unmarshal(byteValue, &etcd)
		for ; hostStartIP <= hostStopIP; hostStartIP++ {
			hostStartIPstr := strconv.Itoa(hostStartIP)
			etcd.Hosts = append(etcd.Hosts, hostIPSplit+hostStartIPstr)
		}
		byteValue, _ = json.Marshal(etcd)
		// wirte json to etcd-csr.json
		err = ioutil.WriteFile("/tmp/k8s/cert/etcd-csr.json", byteValue, 0644)
		if err != nil {
			log.Error(err)
		}
		// create etcd cert
		shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes etcd-csr.json  |  " + k8spath + "tools/cfssljson -bare etcd"
		log.Info(shell)
		if !ShellOut(shell) {
			log.Error("create ETCD cert faild!!!")
		}
		// copy etcd to hosts
		// for ; hostStartIP <= hostStopIP; hostStartIP++ {
		// 	log.Info(hostStartIP)
		// 	hostStartIPstr := strconv.Itoa(hostStartIP)
		// 	log.Info(hostStartIPstr)
		// 	go InstallEtcd(hostIPSplit+hostStartIPstr, para[`pwd`].(string), k8spath, &wg)
		// }
		// wg.Wait()
		var etcdnode string
		for i := 0; i <= threadNum-1; i++ {
			hostStartIPstr := strconv.Itoa(hostStartIP - i - 1)
			etcdID := strconv.Itoa(i + 1)
			etcdnode = "etcd" + etcdID + "=https://" + hostIPSplit + hostStartIPstr + ":2380," + etcdnode
		}
		etcdnode = etcdnode[0 : len(etcdnode)-1]
		for i := 0; i <= threadNum-1; i++ {
			hostStartIPstr := strconv.Itoa(hostStartIP - i - 1)
			etcdID := strconv.Itoa(i + 1)
			log.Info("etcdID:" + etcdID)
			go InstallEtcd(hostIPSplit+hostStartIPstr, para[`pwd`].(string), k8spath, etcdID, etcdnode, &wg)
		}
		wg.Wait()
		log.Info("ETCD Install Done.")
	}
}
