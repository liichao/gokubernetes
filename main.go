package main

import (
	"encoding/json"
	"fmt"
	"go-install-kubernetes/k8stools"
	myTools "go-install-kubernetes/tools"
	"io/ioutil"
	"os"
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
	confLists := []string{"cni-default.conf", "10-k8s-modules.conf", "95-k8s-sysctl.conf", "30-k8s-ulimits.conf", "sctp.conf", "server-centos.conf", "daemon.json", "docker"}
	certFileLists := []string{"kubelet-csr.json", "kubernetes-csr.json", "basic-auth.csv", "aggregator-proxy-csr.json", "etcd-csr.json", "admin-csr.json", "ca-config.json", "ca-csr.json", "kube-controller-manager-csr.json", "kube-proxy-csr.json", "kube-scheduler-csr.json", "read-csr.json"}
	//toolsFileLists := []string{"cfssl", "cfssljson", "hyperkube","etcd.tar.gz"}
	//toolsFileLists := []string{"cfssl", "cfssljson"}
	yamlFileLists := []string{"read-group-rbac.yaml", "basic-auth-rbac.yaml", "kubelet-config.yaml"}
	serviceFileLists := []string{"kubelet.service", "kube-proxy.service", "kube-apiserver.service", "kube-scheduler.service", "kube-controller-manager.service", "etcd.service", "docker.service"}
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
	log.Info(" 将配置文件生成到/tmp/k8s目录中...")
	for _, file := range confLists {
		filebytes, err := Asset("config/" + file)
		if err != nil {
			panic(err)
		}
		myTools.CheckCreateWriteFile(k8spath, file, string(filebytes))
	}
	// 将证书释放到相关目录
	for _, file := range certFileLists {
		filebytes, err := Asset("config/cert/" + file)
		if err != nil {
			panic(err)
		}
		myTools.CheckCreateWriteFile(k8spath+"cert/", file, string(filebytes))
	}
	// 将tools文件释放到相关目录
	// for _, file := range toolsFileLists {
	// 	filebytes, err := Asset("config/tools/" + file)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	myTools.CheckCreateWriteFile(k8spath+"tools/", file, string(filebytes))
	// }
	// 将yaml文件释放到相关目录
	for _, file := range yamlFileLists {
		filebytes, err := Asset("config/yaml/" + file)
		if err != nil {
			panic(err)
		}
		myTools.CheckCreateWriteFile(k8spath+"yaml/", file, string(filebytes))
	}
	// 将service文件释放到相关目录
	for _, file := range serviceFileLists {
		filebytes, err := Asset("config/service/" + file)
		if err != nil {
			panic(err)
		}
		myTools.CheckCreateWriteFile(k8spath+"service/", file, string(filebytes))
	}
	//var hostIp string
	hostIPSplit, hostStartIP, hostStopIP := myTools.GetIPDes(para[`ips`].(string))
	// start theard
	threadNum := hostStopIP - hostStartIP + 1
	wg.Add(threadNum)
	switch para[`para`] {
	case `system`:
		log.Info(" 开始配置系统参数...")
		for ; hostStartIP <= hostStopIP; hostStartIP++ {
			log.Info(hostStartIP)
			//ConfigSystem(para[`ips`].(string), para[`pwd`].(string), para[`proxymode`].(string), &wg)
			hostStartIPstr := strconv.Itoa(hostStartIP)
			go k8stools.ConfigSystem(hostIPSplit+hostStartIPstr, para[`pwd`].(string), para[`proxymode`].(string), &wg)
		}
		wg.Wait()
		log.Info("所有主机均已配置完成.")
	case `chrony`:
		log.Info(" 开始安装并配置chrony服务...")
		// InstallChrony(para[`ips`].(string), para[`pwd`].(string), para[`ntpserver`].(string))
		for ; hostStartIP <= hostStopIP; hostStartIP++ {
			log.Info(hostStartIP)
			hostStartIPstr := strconv.Itoa(hostStartIP)
			go k8stools.InstallChrony(hostIPSplit+hostStartIPstr, para[`pwd`].(string), para[`ntpserver`].(string), &wg)
		}
		wg.Wait()
		log.Info("所有主机安装chrony服务完成")
	case `createcert`:
		log.Info(" 开始创建相关需要的证书...")
		k8stools.CreateCert(k8spath)
	case `etcd`:
		var etcd EtcdJSONParse
		log.Info("开始安装Etcd服务....")
		// unzip etcd.tar.gz
		shell := "cd " + k8spath + "tools/ && tar -zxf etcd.tar.gz && mv " + k8spath + "tools/etcd-* " + k8spath + "tools/etcd"
		log.Info("start run " + shell)
		if !myTools.ShellOut(shell) {
			log.Error("解压 etcd.tar.gz文件失败!!!")
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
		if !myTools.ShellOut(shell) {
			log.Error("create ETCD cert faild!!!")
		}
		// uninstall etcd
		if para[`handle`].(string) == "uninstall" {
			for i := 0; i <= threadNum-1; i++ {
				hostStartIPstr := strconv.Itoa(hostStartIP - i - 1)
				go k8stools.RemoveEtcd(hostIPSplit+hostStartIPstr, para[`pwd`].(string), &wg)
			}
			wg.Wait()
			log.Info("ETCD全部卸载完成.")
		}
		// install etcd
		if para[`handle`].(string) == "install" {
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
				go k8stools.InstallEtcd(hostIPSplit+hostStartIPstr, para[`pwd`].(string), k8spath, etcdID, etcdnode, &wg)
			}
			wg.Wait()
			log.Info("ETCD集群安装完成.")
			log.Info("请运行`/opt/kubernetes/bin/etcdctl --endpoints=https://IP:2379  --cacert=/etc/etcd/ssl/ca.pem  --cert=/etc/etcd/ssl/etcd.pem  --key=/etc/etcd/ssl/etcd-key.pem  endpoint health` 检查etcd集群状态是否正常.")
		}
	case `docker`:
		log.Info(" Start install docker...")
		// check docker.tgz Exists
		if !myTools.Exists("/tmp/k8s/tools/docker-" + para[`version`].(string) + ".tgz") {
			log.Info("docker.tgz压缩包不存在，请下载并上传到/tmp/k8s/tools/目录下")
			log.Info("docker.tgz download url : https://download.docker.com/linux/static/stable/x86_64/docker-" + para[`version`].(string) + ".tgz")
		}
		// install docker
		if para[`handle`].(string) == "install" {
			shell := "cd " + k8spath + "tools/ && tar zxvf docker-" + para[`version`].(string) + ".tgz"
			log.Info("start run " + shell)
			if !myTools.ShellOut(shell) {
				log.Error("unzip docker-" + para[`version`].(string) + ".tgz faild!!!")
			}
			// handle ssh host
			for ; hostStartIP <= hostStopIP; hostStartIP++ {
				hostStartIPstr := strconv.Itoa(hostStartIP)
				log.Info(hostIPSplit + hostStartIPstr)
				go k8stools.InstallDocker(hostIPSplit+hostStartIPstr, para[`pwd`].(string), k8spath, &wg)
			}
			wg.Wait()
			log.Info("安装docker完成.")
		}
		// uninstall docker
		if para[`handle`].(string) == "uninstall" {
			for ; hostStartIP <= hostStopIP; hostStartIP++ {
				hostStartIPstr := strconv.Itoa(hostStartIP)
				log.Info(hostIPSplit + hostStartIPstr)
				go k8stools.RemoveDocker(hostIPSplit+hostStartIPstr, para[`pwd`].(string), &wg)
			}
			wg.Wait()
			log.Info("Docker卸载完成.")
		}
	case `k8s`:
		log.Info(" 开始安装k8smaster三大组件...")
		// 拼接etcd集群字符串
		// etcdlist=https://10.10.76.222:2379,https://10.10.76.223:2379,https://10.10.76.225:2379
		etcdIPSplit, etcdStartIP, etcdStopIP := myTools.GetIPDes(para[`etcd`].(string))
		if para[`handle`].(string) == "install" {
			var etcdNodeList string
			for ; etcdStartIP <= etcdStopIP; etcdStartIP++ {
				etcdStartIPstr := strconv.Itoa(etcdStartIP)
				log.Info(etcdIPSplit + etcdStartIPstr)
				if etcdStartIP == etcdStopIP {
					etcdNodeList = etcdNodeList + "https://" + etcdIPSplit + etcdStartIPstr + ":2379"
				} else {
					etcdNodeList = etcdNodeList + "https://" + etcdIPSplit + etcdStartIPstr + ":2379,"
				}
			}
			log.Info("etcdlist:" + etcdNodeList)
			if para[`handle`].(string) == "install" {
				for ; hostStartIP <= hostStopIP; hostStartIP++ {
					hostStartIPstr := strconv.Itoa(hostStartIP)
					log.Info(hostIPSplit + hostStartIPstr)
					go k8stools.InstallK8sMaster(hostIPSplit+hostStartIPstr, para[`pwd`].(string), k8spath, para[`nodeportrange`].(string), para[`svcIP`].(string), etcdNodeList, para[`clusterIP`].(string), para[`nodeCidrLen`].(string), &wg)
				}
				wg.Wait()
				log.Info("k8s master 三大组件安装完成.")
			}
			// kubectl apply -f basic-auth-rbac.yaml 随便找一台机器执行即可
			c, err := ssh.NewClient(hostIPSplit+strconv.Itoa(hostStartIP), "22", "root", para[`pwd`].(string))
			if err != nil {
				panic(err)
			}
			defer c.Close()
			err = c.Exec("/opt/kubernetes/bin/hyperkube kubectl apply -f /opt/kubernetes/cfg/basic-auth-rbac.yaml")
			if err != nil {
				log.Error(err)
			}
		}
		//删除操作
		if para[`handle`].(string) == "uninstall" {
			for ; hostStartIP <= hostStopIP; hostStartIP++ {
				hostStartIPstr := strconv.Itoa(hostStartIP)
				log.Info(hostIPSplit + hostStartIPstr)
				go k8stools.RemoveK8sMaster(hostIPSplit+hostStartIPstr, para[`pwd`].(string), &wg)
			}
			wg.Wait()
			log.Info("k8s 删除完成.")
		}
	case `node`:
		hostIPSplit, hostStartIP, hostStopIP = myTools.GetIPDes(para[`node`].(string))
		// to-do 这边临时取了第一个masterip为apiserverip
		apiserver := hostIPSplit + strconv.Itoa(hostStartIP)
		if para[`handle`].(string) == "install" {
			
			for ; hostStartIP <= hostStopIP; hostStartIP++ {
				hostStartIPstr := strconv.Itoa(hostStartIP)
				log.Info(hostIPSplit + hostStartIPstr)
				go k8stools.InstallK8sNode(hostIPSplit+hostStartIPstr, para[`pwd`].(string), para[`svcIP`].(string), k8spath, apiserver, para[`maxPods`].(string), para[`clusterIP`].(string), para[`proxyMode`].(string), para[`pauseImage`].(string), &wg)
			}
			wg.Wait()
			log.Info("k8s master 三大组件安装完成.")
		}
	}
}
