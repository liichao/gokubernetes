package k8stools

import (
	"strings"
	"sync"

	"github.com/pytool/ssh"
)

// ConfigSystem 配置系统参数
func ConfigSystem(ip, pwd, proxymode, k8spath string, sysversionint int, ws *sync.WaitGroup) {
	log.Info("开始配置" + ip + "地址...")
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// 开始更新系统版本与安装相关必要组件
	log.Info("开始更新系统版本与安装相关必要组件")
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
	err = c.Exec("sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config")
	if err != nil {
		log.Info(err)
	}
	// 优化rsyslog获取journald日志
	log.Info(ip + " 优化rsyslog获取journald日志")
	err = c.Exec("sed -i 's/$ModLoad imjournal/#$ModLoad imjournal/g' /etc/rsyslog.conf")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("sed -i 's/$IMJournalStateFile/#$IMJournalStateFile/g' /etc/rsyslog.conf")
	if err != nil {
		log.Info(err)
	}
	log.Info(ip + " 重启rsyslog服务,以使配置正常.")
	err = c.Exec("systemctl restart rsyslog ")
	if err != nil {
		log.Info(err)
	}
	// // 加载内核模块
	// log.Info(ip + "获取内核系统版本,并根据不同版本加载内核模块")
	// versionoutput, err := c.Output("uname -a")
	// if err != nil {
	// 	panic(err)
	// }
	// log.Info(string(versionoutput[:]))
	// sysversion := strings.Split(strings.Split(string(versionoutput[:]), " ")[2], "1")[1]
	// log.Info(ip + " 系统版本Version:" + sysversion)
	// sysversionint, err := strconv.Atoi(strings.Split(sysversion, ".")[0])
	// if err != nil {
	// 	log.Error("字符串转换成整数失败")
	// }
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
	log.Info(ip + " 拷贝 10-k8s-modules.conf 到 /etc/modules-load.d/")
	err = c.Upload(k8spath+"10-k8s-modules.conf", "/etc/modules-load.d/")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("echo " + nfConntrack + ">> /etc/modules-load.d/10-k8s-modules.conf")
	if err != nil {
		log.Info(err)
	}
	// 将95-k8s-sysctl.conf放到服务器的指定目录/etc/sysctl.d
	log.Info(ip + " 拷贝 95-k8s-sysctl.conf 到 /etc/sysctl.d/")
	err = c.Upload(k8spath+"95-k8s-sysctl.conf", "/etc/sysctl.d/")
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
	log.Info("应用系统参数 /etc/sysctl.d/95-k8s-sysctl.conf")
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
	log.Info(ip + " 拷贝 30-k8s-ulimits.conf 到 /etc/systemd/system.conf.d/")
	err = c.Upload(k8spath+"30-k8s-ulimits.conf", "/etc/systemd/system.conf.d/")
	if err != nil {
		log.Info(err)
	}
	// 把SCTP列入内核模块黑名单
	err = c.Exec("mkdir -p /etc/systemd/system.conf.d")
	if err != nil {
		log.Error(err)
	}
	err = c.Upload(k8spath+"sctp.conf", "/etc/modprobe.d/")
	if err != nil {
		log.Info(err)
	}
}

// InstallChrony 安装时钟同步服务器
func InstallChrony(ip, pwd, ntpserver, k8spath string, ws *sync.WaitGroup) {
	log.Info("开始配置" + ip + "地址...")
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// remote ntp and install chrony
	log.Info(ip + " 删除 ntp server并安装chrony")
	err = c.Exec("yum remove -y ntp && yum install -y chrony")
	if err != nil {
		log.Error(err)
	}
	// 更改时区为上海时区
	err = c.Exec("cp -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime")
	if err != nil {
		log.Error(err)
	}
	err = c.Upload(k8spath+"server-centos.conf", "/etc/chrony.conf")
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

// RepirIptables 将iptables升级到指定版本
func RepirIptables(ip, pwd, k8spath, iptablesBinName string, ws *sync.WaitGroup) {
	log.Info("开始配置" + ip + "地址...")
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	// 分发iptables 安装包
	err = c.Upload(k8spath+"tools/"+iptablesBinName, "/tmp/")
	if err != nil {
		log.Info(err)
	}
	// 安装依赖
	err = c.Exec("yum install -y gcc make libnftnl-devel libmnl-devel autoconf automake libtool bison flex  libnetfilter_conntrack-devel libnetfilter_queue-devel libpcap-devel")
	if err != nil {
		log.Info(err)
	}
	// 对该编译二进制
	// 获取解压文件夹名字
	binName := strings.Split(iptablesBinName, `.tar`)[0]
	err = c.Exec("cd /tmp/ && export LC_ALL=C && tar -xvf " + iptablesBinName + " && cd " + binName + " &&  ./autogen.sh")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("cd /tmp/" + binName + "  && ./configure")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("cd /tmp/" + binName + "  && make -j4")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("cd /tmp/" + binName + "  && make install")
	if err != nil {
		log.Info(err)
	}
	// 创建永久软件
	linkName := []string{"ip6tables", "ip6tables-restore", "ip6tables-save", "iptables", "iptables-restore", "iptables-save"}
	// for _, name := range linkName {
	// 	err = c.Exec("ln -s /usr/local/sbin/xtables-multi /usr/local/sbin/" + name)
	// 	if err != nil {
	// 		log.Info(err)
	// 	}
	// }
	// 应用到正式系统 覆盖掉/sbin/下的相关文件
	for _, name := range linkName {
		err = c.Exec("cp -rf  /usr/local/sbin/" + name + " /sbin/" + name)
		if err != nil {
			log.Info(err)
		}
	}
	// 重启kubelet和 kube-proxy
	err = c.Exec("systemctl restart kube-proxy")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("systemctl restart kubelet")
	if err != nil {
		log.Info(err)
	}
}

// ChangeHarborHost 将harbor ip写入到node的hosts 与域名对应
func ChangeHarborHost(ip, pwd, harborURL, harborIP, harborUser, harborPwd, pauseImage string, ws *sync.WaitGroup) {
	log.Info("开始配置" + ip + "地址...")
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	// 当命令执行完成后关闭
	defer c.Close()
	//  执行 echo  "域名     ip " >> /etc/hosts
	shell := "echo '" + harborIP + "    " + harborURL + "'>> /etc/hosts"
	log.Info(ip + " 执行: " + shell)
	err = c.Exec(shell)
	if err != nil {
		log.Info(err)
	}
	// 与仓库建立连接
	shell = "docker login -u " + harborUser + " -p " + harborPwd + " " + harborURL
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	shell = "docker pull '" + pauseImage
	log.Info(ip + " 执行: " + shell)
	err = c.Exec(shell)
	if err != nil {
		log.Info(err)
	}
}
