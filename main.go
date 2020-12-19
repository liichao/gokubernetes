package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"go-install-kubernetes/k8stools"
	"go-install-kubernetes/tools"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/op/go-logging"
	"github.com/pytool/ssh"
	"github.com/spf13/viper"
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
	// 读取yaml配置文件
	config := viper.New()
	config.AddConfigPath(".")
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	// 获取相关参数
	system := config.Get("system").(bool)
	k8spath := config.Get("k8sPath").(string)
	chrony := config.Get("chrony.enable").(bool)
	ntpServer := config.Get("chrony.ntpServer").(string)
	allIPList := list.New()
	masterIPList := list.New()
	nodeIPList := list.New()
	masterIP := config.Get("master").(string)
	nodeIP := config.Get("node").(string)
	hostIPSplit, hostStartIP, hostStopIP := tools.GetIPDes(masterIP)
	for ; hostStartIP <= hostStopIP; hostStartIP++ {
		hostStartIPstr := strconv.Itoa(hostStartIP)
		allIPList.PushBack(hostIPSplit + hostStartIPstr)
		masterIPList.PushBack(hostIPSplit + hostStartIPstr)
	}
	hostIPSplit, hostStartIP, hostStopIP = tools.GetIPDes(nodeIP)
	for ; hostStartIP <= hostStopIP; hostStartIP++ {
		hostStartIPstr := strconv.Itoa(hostStartIP)
		allIPList.PushBack(hostIPSplit + hostStartIPstr)
		nodeIPList.PushBack(hostIPSplit + hostStartIPstr)
	}
	// 循环所有ip
	// for ip := allIPList.Front(); ip != nil; ip = ip.Next() {
	// 	log.Info(ip.Value)
	// }
	password := config.Get("password").(string)
	proxyMode := config.Get("proxyMode").(string)
	createCert := config.Get("createCert").(bool)
	etcd := config.Get("etcd").(bool)
	docker := config.Get("docker").(bool)
	svcIP := config.Get("svcIP").(string)
	clusterIP := config.Get("clusterIP").(string)
	nodeCidrLen := config.Get("nodeCidrLen").(int)
	nodePortRange := config.Get("nodePortRange").(string)
	maxPods := config.Get("maxPods").(int)
	flannelBackend := config.Get("flannelBackend").(string)
	harborURL := config.Get("harborUrl").(string)
	harborUser := config.Get("harborUser").(string)
	harborPwd := config.Get("harborPwd").(string)
	flanneldImage := config.Get("flanneldImage").(string)
	pauseImage := config.Get("pauseImage").(string)
	etcdInstall := config.Get("etcdInstall").(bool)
	apiServer := config.Get("apiServer").(string)
	kernel := config.Get("kernel").(int)
	dockerInstall := config.Get("dockerInstall").(bool)
	dockerInstalltgz := config.Get("dockerInstalltgz").(string)
	clusterDnsDomain := config.Get("clusterDnsDomain").(string)
	iptablesBinName := config.Get("repair.iptablesBinName").(string)
	log.Info(flanneldImage)
	log.Info(harborPwd)
	log.Info(harborUser)
	log.Info(harborURL)
	log.Info(flannelBackend)

	// 相关文件
	confLists := []string{"cni-default.conf", "10-k8s-modules.conf", "95-k8s-sysctl.conf", "30-k8s-ulimits.conf", "sctp.conf", "server-centos.conf", "daemon.json", "docker"}
	certFileLists := []string{"kubelet-csr.json", "kubernetes-csr.json", "basic-auth.csv", "aggregator-proxy-csr.json", "etcd-csr.json", "admin-csr.json", "ca-config.json", "ca-csr.json", "kube-controller-manager-csr.json", "kube-proxy-csr.json", "kube-scheduler-csr.json", "read-csr.json"}
	yamlFileLists := []string{"coredns.yaml", "kube-flannel-vxlan.yaml", "kube-flannel.yaml", "read-group-rbac.yaml", "basic-auth-rbac.yaml", "kubelet-config.yaml"}
	serviceFileLists := []string{"kubelet.service", "kube-proxy.service", "kube-apiserver.service", "kube-scheduler.service", "kube-controller-manager.service", "etcd.service", "docker.service"}
	toolsFileLists := []string{"cfssl", "cfssljson", "hyperkube", "etcd.tar.gz", "cni-plugins-linux-amd64-v0.9.0.tgz", "flanneld-v0.13.0-amd64.docker", "xtables-multi-iptables-1.6.2"}
	log.Info(certFileLists)
	log.Info(yamlFileLists)
	log.Info(serviceFileLists)
	log.Info(toolsFileLists)
	// 将配置文件生成到k8spath目录中
	log.Info(" 将配置文件生成到" + k8spath + "目录中...")
	for _, file := range confLists {
		filebytes, err := Asset("config/" + file)
		if err != nil {
			panic(err)
		}
		tools.CheckCreateWriteFile(k8spath, file, string(filebytes))
	}
	// 将证书释放到相关目录
	for _, file := range certFileLists {
		filebytes, err := Asset("config/cert/" + file)
		if err != nil {
			panic(err)
		}
		tools.CheckCreateWriteFile(k8spath+"cert/", file, string(filebytes))
	}
	// 将yaml文件释放到相关目录
	for _, file := range yamlFileLists {
		filebytes, err := Asset("config/yaml/" + file)
		if err != nil {
			panic(err)
		}
		tools.CheckCreateWriteFile(k8spath+"yaml/", file, string(filebytes))
	}
	// 将service文件释放到相关目录
	for _, file := range serviceFileLists {
		filebytes, err := Asset("config/service/" + file)
		if err != nil {
			panic(err)
		}
		tools.CheckCreateWriteFile(k8spath+"service/", file, string(filebytes))
	}
	// // 将tools文件释放到相关目录
	// // for _, file := range toolsFileLists {
	// // 	filebytes, err := Asset("config/tools/" + file)
	// // 	if err != nil {
	// // 		panic(err)
	// // 	}
	// // 	tools.CheckCreateWriteFile(k8spath+"tools/", file, string(filebytes))
	// // }
	// 修改docker配置文件，将harbor地址加入到信任地址
	shell := "sed -i 's%HarborURL%https://" + harborURL + "%g' " + k8spath + "daemon.json"
	if !tools.ShellOut(shell) {
		log.Error("替换harborURL失败")
	}
	// 创建并发
	var wg sync.WaitGroup
	log.Info("判断系统参数配置是否启用...")
	if system {
		log.Info("开始配置系统参数...")
		wg.Add(allIPList.Len())
		for ip := allIPList.Front(); ip != nil; ip = ip.Next() {
			log.Info(ip.Value.(string))
			go k8stools.ConfigSystem(ip.Value.(string), password, proxyMode, k8spath, kernel, &wg)
		}
		wg.Wait()
		log.Info("所有主机均已配置完成.")
	}
	log.Info("判断时钟参数配置是否启用...")
	if chrony {
		log.Info(" 开始安装并配置chrony服务...")
		wg.Add(allIPList.Len())
		for ip := allIPList.Front(); ip != nil; ip = ip.Next() {
			log.Info(ip.Value.(string))
			go k8stools.InstallChrony(ip.Value.(string), password, ntpServer, k8spath, &wg)
		}
		wg.Wait()
		log.Info("所有主机安装chrony服务完成")
	}
	log.Info("判断是否要创建证书...")
	if createCert {
		log.Info(" 开始在当前主机创建相关证书...")
		k8stools.CreateCert(k8spath, apiServer)
	}
	log.Info("判断是否安装etcd集群...")
	if etcd {
		wg.Add(masterIPList.Len())
		// install etcd
		if etcdInstall {
			var etcd EtcdJSONParse
			log.Info("开始安装Etcd服务....")
			// unzip etcd.tar.gz
			shell := "cd " + k8spath + "tools/ && tar -zxf etcd.tar.gz && mv " + k8spath + "tools/etcd-* " + k8spath + "tools/etcd"
			log.Info("start run " + shell)
			if !tools.ShellOut(shell) {
				log.Error("解压 etcd.tar.gz文件失败!!!")
			}
			// load etcd-csr.json
			jsonFile, _ := os.Open(k8spath + "cert/etcd-csr.json")
			defer jsonFile.Close()
			byteValue, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				log.Error(err)
			}
			json.Unmarshal(byteValue, &etcd)
			hostIPSplit, hostStartIP, hostStopIP = tools.GetIPDes(masterIP)
			for ; hostStartIP <= hostStopIP; hostStartIP++ {
				hostStartIPstr := strconv.Itoa(hostStartIP)
				etcd.Hosts = append(etcd.Hosts, hostIPSplit+hostStartIPstr)
			}
			byteValue, _ = json.Marshal(etcd)
			// wirte json to etcd-csr.json
			err = ioutil.WriteFile(k8spath+"cert/etcd-csr.json", byteValue, 0644)
			if err != nil {
				log.Error(err)
			}
			// create etcd cert
			shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes etcd-csr.json  |  " + k8spath + "tools/cfssljson -bare etcd"
			log.Info(shell)
			if !tools.ShellOut(shell) {
				log.Error("create ETCD cert faild!!!")
			}
			var etcdnode string
			for i := 0; i <= masterIPList.Len()-1; i++ {
				hostStartIPstr, _ := strconv.Atoi(tools.GetIPLastString(masterIPList.Front().Value.(string)))
				hostStartIPstr = hostStartIPstr + i
				etcdIPThree := tools.GetIPString(masterIPList.Front().Value.(string))
				etcdID := strconv.Itoa(i + 1)
				etcdnode = "etcd" + etcdID + "=https://" + etcdIPThree + strconv.Itoa(hostStartIPstr) + ":2380," + etcdnode
			}
			etcdnode = etcdnode[0 : len(etcdnode)-1]
			for i := 0; i <= masterIPList.Len()-1; i++ {
				hostStartIPstr, _ := strconv.Atoi(tools.GetIPLastString(masterIPList.Front().Value.(string)))
				hostStartIPstr = hostStartIPstr + i
				etcdIPThree := tools.GetIPString(masterIPList.Front().Value.(string))
				etcdID := strconv.Itoa(i + 1)
				log.Info("etcdID:" + etcdID)
				go k8stools.InstallEtcd(etcdIPThree+strconv.Itoa(hostStartIPstr), password, k8spath, etcdID, etcdnode, &wg)
			}
			wg.Wait()
			log.Info("ETCD集群安装完成.")
			log.Info("请运行`/opt/kubernetes/bin/etcdctl --endpoints=https://IP:2379  --cacert=/etc/etcd/ssl/ca.pem  --cert=/etc/etcd/ssl/etcd.pem  --key=/etc/etcd/ssl/etcd-key.pem  endpoint health` 检查etcd集群状态是否正常.")
		} else {
			// 卸载 etcd
			for ip := masterIPList.Front(); ip != nil; ip = ip.Next() {
				log.Info(ip.Value.(string))
				go k8stools.RemoveEtcd(ip.Value.(string), password, &wg)

			}
			wg.Wait()
			log.Info("ETCD全部卸载完成.")
		}
	}
	log.Info("判断给node节点安装docker...")
	if docker {
		wg.Add(nodeIPList.Len())
		// install docker
		if dockerInstall {
			log.Info(" 开始安装docker...")
			// check docker.tgz Exists
			if !tools.Exists(k8spath + "tools/" + dockerInstalltgz) {
				log.Info("docker.tgz压缩包不存在，请下载并上传到" + k8spath + "tools/目录下")
				log.Info("docker.tgz download url : https://download.docker.com/linux/static/stable/x86_64/docker-18.09.6.tgz")
			}
			shell := "cd " + k8spath + "tools/ && tar zxvf " + dockerInstalltgz
			log.Info("start run " + shell)
			if !tools.ShellOut(shell) {
				log.Error("unzip " + dockerInstalltgz + "faild!!!")
			}
			for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
				log.Info(ip.Value.(string))
				go k8stools.InstallDocker(ip.Value.(string), password, k8spath, &wg)
			}
			wg.Wait()
			log.Info("安装docker完成.")
		} else {
			// 卸载 docker
			for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
				log.Info(ip.Value.(string))
				go k8stools.RemoveDocker(ip.Value.(string), password, &wg)
			}
			wg.Wait()
			log.Info("Docker卸载完成.")
		}
	}
	log.Info("判断是否安装k8s Master服务...")
	if config.Get("installK8sApi").(bool) {
		wg.Add(masterIPList.Len())
		log.Info("开始安装api相关服务")
		// 拼接etcd集群字符串
		// etcdlist=https://10.10.76.222:2379,https://10.10.76.223:2379,https://10.10.76.225:2379
		etcdIPSplit, etcdStartIP, etcdStopIP := tools.GetIPDes(masterIP)
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
		for ip := masterIPList.Front(); ip != nil; ip = ip.Next() {
			log.Info(ip.Value.(string))
			go k8stools.InstallK8sMaster(ip.Value.(string), password, k8spath, nodePortRange, svcIP, etcdNodeList, clusterIP, strconv.Itoa(nodeCidrLen), &wg)
		}
		wg.Wait()
		log.Info("k8s master 三大组件安装完成.")
		// kubectl apply -f basic-auth-rbac.yaml 随便找一台机器执行即可
		c, err := ssh.NewClient(masterIPList.Front().Value.(string), "22", "root", password)
		if err != nil {
			panic(err)
		}
		defer c.Close()
		err = c.Exec("/opt/kubernetes/bin/hyperkube kubectl apply -f /opt/kubernetes/cfg/basic-auth-rbac.yaml")
		if err != nil {
			log.Error(err)
		}
	}
	log.Info("判断是否删除远程api服务...")
	if config.Get("removeK8sApi").(bool) {
		wg.Add(masterIPList.Len())
		for ip := masterIPList.Front(); ip != nil; ip = ip.Next() {
			log.Info(ip.Value.(string))
			go k8stools.RemoveK8sMaster(ip.Value.(string), password, &wg)
		}
		wg.Wait()
		log.Info("k8s 删除完成.")
	}
	log.Info("判断是否安装k8s Node服务...")
	if config.Get("installK8sNode").(bool) {
		log.Info("开始安装node相关服务")
		wg.Add(nodeIPList.Len())
		for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
			go k8stools.InstallK8sNode(ip.Value.(string), password, svcIP, k8spath, apiServer, strconv.Itoa(maxPods), clusterIP, proxyMode, pauseImage, clusterDnsDomain, &wg)
		}
		wg.Wait()
		log.Info("k8s node kubelet kube-proxy组件安装完成.")
	}
	log.Info("判断是否卸载k8s Node服务...")
	if config.Get("removeK8sNode").(bool) {
		wg.Add(nodeIPList.Len())
		for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
			go k8stools.RemoveK8sNode(ip.Value.(string), password, &wg)
		}
		wg.Wait()
		log.Info("k8s node kubelet kube-proxy组件卸载完成.")
	}
	log.Info("判断是否要修复iptbales版本过低导致--random-fully问题...")
	if config.Get("repair.iptables").(bool) {
		log.Warning("修复iptbales,需要yum安装依赖，请联网或者配置本地全量包")
		// 判断iptables文件是否释放到k8spath/tools目录下
		if !tools.Exists(k8spath + "tools/" + iptablesBinName) {
			log.Info("iptables安装包不存在，请上传到" + k8spath + "tools/目录下")
			log.Info("编译参考URL : https://zhizhebuyan.com/2020/12/19/%E8%A7%A3%E5%86%B3k8s-node%E8%8A%82%E7%82%B9kubelet-network-linux-go-141-Not-using-in-the-MASQUERADE-rule-for-iptables-because-the-local-version-of-iptables-does-not-support-it-%E6%8A%A5%E9%94%99%E9%97%AE%E9%A2%98/")
		}
		wg.Add(nodeIPList.Len())
		for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
			go k8stools.RepirIptables(ip.Value.(string), password, k8spath, iptablesBinName, &wg)
		}
		wg.Wait()
		log.Info("k8s node kubelet kube-proxy的iptables修复完成.")
	}
	// 安装 flanneld
	if config.Get("network.install").(bool) {
		log.Info("安装网络插件")
		// 创建registry secret,在masterList的第一台执行即可或者直接在这台机器执行也可
		shell := k8spath + "tools/hyperkube kubectl create secret docker-registry myregistrykey --docker-server=" + config.Get("harborUrl").(string) + " --docker-username=" + config.Get("harborUser").(string) + " --docker-password=" + config.Get("harborPwd").(string) + " --docker-email=dzero@dero.com -n kube-system"
		if !tools.ShellOut(shell) {
			log.Error("创建kube-system registry secret失败!!!")
		}
		shell = k8spath + "tools/hyperkube kubectl create secret docker-registry myregistrykey --docker-server=" + config.Get("harborUrl").(string) + " --docker-username=" + config.Get("harborUser").(string) + " --docker-password=" + config.Get("harborPwd").(string) + " --docker-email=dzero@dero.com"
		if !tools.ShellOut(shell) {
			log.Error("创建registry secret失败!!!")
		}
		//配置仓库相关信息，将harborIP 地址与域名写死到node的hosts，如果不是内网可以忽略
		if config.Get("harborIP.private").(bool) {
			wg.Add(nodeIPList.Len())
			for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
				go k8stools.ChangeHarborHost(ip.Value.(string), password, config.Get("harborUrl").(string), config.Get("harborIP.IP").(string), &wg)
			}
			wg.Wait()
			log.Info("所有node的hosts都修改完成.")
		}
		// 判断是否要载入images
		if config.Get("network.loadImage").(bool) {
			// 判断flannel镜像文件是否存在
			if !tools.Exists(k8spath + "tools/" + config.Get("flanneldImageOffline").(string)) {
				log.Warning(config.Get("flanneldImageOffline").(string) + "镜像包不存在，请上传到" + k8spath + "tools/目录下")
				log.Warning("下载URL : https://github.com/coreos/flannel/releases")
			}
			// 判断pause镜像文件是否存在
			if !tools.Exists(k8spath + "tools/" + config.Get("pauseImageOffline").(string)) {
				log.Info(config.Get("pauseImageOffline").(string) + "镜像包不存在，请上传到" + k8spath + "tools/目录下")
			}
			// 将镜像包上传到nodeList的第一台之后载入并推送到docker仓库
			c, err := ssh.NewClient(nodeIPList.Front().Value.(string), "22", "root", password)
			if err != nil {
				panic(err)
			}
			// 当命令执行完成后关闭
			defer c.Close()

			imagesOffline := []string{config.Get("pauseImageOffline").(string), config.Get("flanneldImageOffline").(string)}
			for _, images := range imagesOffline {
				// 上传镜像到/tmp目录
				err = c.Upload(k8spath+"tools/"+images, "/tmp/")
				if err != nil {
					log.Info(err)
					log.Error(config.Get("pauseImageOffline").(string) + " 上传" + images + "镜像失败")
				}
				// 载入镜像
				err = c.Exec("docker load -i /tmp/" + images)
				if err != nil {
					log.Error(err)
					log.Error(config.Get("pauseImageOffline").(string) + " 载入" + images + "镜像失败")
				}
			}
			// 修改镜像tag 并推送到docker 仓库
			shell = "docker tag quay.io/coreos/" + strings.Split(config.Get("flanneldImageOffline").(string), ".doc")[0] + " " + flanneldImage
			log.Info(shell)
			err = c.Exec(shell)
			if err != nil {
				log.Error(err)
			}
			// 与仓库建立连接
			shell = "docker login -u " + harborUser + " -p " + harborPwd + " " + harborURL
			log.Info(shell)
			err = c.Exec(shell)
			if err != nil {
				log.Error(err)
			}
			err = c.Exec("docker push " + flanneldImage)
			if err != nil {
				log.Error(err)
			}
			err = c.Exec("docker push " + pauseImage)
			if err != nil {
				log.Error(err)
			}
		}
		wg.Add(nodeIPList.Len())
		for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
			go k8stools.InstallK8sNetwork(ip.Value.(string), password, k8spath, flannelBackend, &wg)
		}
		wg.Wait()
		log.Info("k8s node kubelet kube-proxy组件安装完成.")
		log.Info("开始安装flanneld " + flannelBackend)
		if flannelBackend == "vxlan" {
			shell = "sed -i 's%flanneld_image%" + flanneldImage + "%g' " + k8spath + "yaml/kube-flannel-vxlan.yaml"
			if !tools.ShellOut(shell) {
				log.Error("替换镜像地址失败")
			}
			shell = "sed -i 's%CLUSTER_CIDR%" + clusterIP + "%g' " + k8spath + "yaml/kube-flannel-vxlan.yaml"
			if !tools.ShellOut(shell) {
				log.Error("替换flanneld的IP段失败!!!")
			}
			shell = "sed -i 's%FLANNEL_BACKEND%" + flannelBackend + "%g' " + k8spath + "yaml/kube-flannel-vxlan.yaml"
			if !tools.ShellOut(shell) {
				log.Error("替换flanneld的模式失败!!!")
			}
			shell = k8spath + "tools/hyperkube kubectl apply -f " + k8spath + "yaml/kube-flannel-vxlan.yaml"
			if !tools.ShellOut(shell) {
				log.Error("创建 flannled失败!!!")
			}
		} else {
			shell = "sed -i 's%flanneld_image%" + flanneldImage + "%g' " + k8spath + "yaml/kube-flannel.yaml"
			if !tools.ShellOut(shell) {
				log.Error("替换镜像地址失败")
			}
			shell = "sed -i 's%CLUSTER_CIDR%" + clusterIP + "%g' " + k8spath + "yaml/kube-flannel.yaml"
			if !tools.ShellOut(shell) {
				log.Error("替换flanneld的IP段失败!!!")
			}
			shell = "sed -i 's%FLANNEL_BACKEND%" + flannelBackend + "%g' " + k8spath + "yaml/kube-flannel.yaml"
			if !tools.ShellOut(shell) {
				log.Error("替换flanneld的模式失败!!!")
			}
			shell = k8spath + "tools/hyperkube kubectl apply -f " + k8spath + "yaml/kube-flannel.yaml"
			if !tools.ShellOut(shell) {
				log.Error("创建 flannled失败!!!")
			}
		}
	}
	// 安装其他相关插件
	// 判断是否安装其它插件
	if config.Get("other.install").(bool) {
		// 判断DNS是否安装
		if config.Get("other.dns.install").(bool) {
			coreDNSImages := config.Get("other.dns.coreDnsImages").(string)
			//判断是否需要载入镜像
			if config.Get("other.dns.loadImage").(bool) {
				cordnsTar := config.Get("other.dns.coredns_offline").(string)
				// 判断coredns_1.8.0.tar镜像文件是否存在
				if !tools.Exists(k8spath + "tools/" + cordnsTar) {
					log.Info(cordnsTar + "镜像包不存在，请上传到" + k8spath + "tools/目录下")
				}
				// 建立ssh将镜像包上传到nodeList的第一台之后载入并推送到docker仓库
				c, err := ssh.NewClient(nodeIPList.Front().Value.(string), "22", "root", password)
				if err != nil {
					panic(err)
				}
				// 当命令执行完成后关闭
				defer c.Close()
				// 上传cordns tar到node的第一台机器
				err = c.Upload(k8spath+"tools/"+cordnsTar, "/tmp/")
				if err != nil {
					log.Info(err)
					log.Error(" 上传" + cordnsTar + "镜像失败")
				}
				// 载入镜像
				err = c.Exec("docker load -i /tmp/" + cordnsTar)
				if err != nil {
					log.Error(err)
					log.Error(" 载入" + cordnsTar + "镜像失败")
				}
				// 修改镜像tag 并推送到docker 仓库
				shell = "docker tag coredns/" + strings.Split(coreDNSImages, "base/")[1] + " " + coreDNSImages
				log.Info(shell)
				err = c.Exec(shell)
				if err != nil {
					log.Error(err)
				}
				// 与仓库建立连接
				shell = "docker login -u " + harborUser + " -p " + harborPwd + " " + harborURL
				log.Info(shell)
				err = c.Exec(shell)
				if err != nil {
					log.Error(err)
				}
				//推送到仓库
				err = c.Exec("docker push " + coreDNSImages)
				if err != nil {
					log.Error(err)
				}
			}
			shell := "sed -i 's%CLUSTER_DNS_SVC_IP%" + config.Get("other.dns.dnsIP").(string) + "%g' " + k8spath + "yaml/coredns.yaml"
			log.Info("替换dnsIP  " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换dnsIP失败!!!")
			}
			shell = "sed -i 's%CLUSTER_DNS_DOMAIN%" + config.Get("clusterDnsDomain").(string) + "%g' " + k8spath + "yaml/coredns.yaml"
			log.Info("替换dns name  " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换dns name 失败!!!")
			}
			// 替换镜像地址
			shell = "sed -i 's%coreDnsImages%" + coreDNSImages + "%g' " + k8spath + "yaml/coredns.yaml"
			log.Info("替换镜像地址 " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换镜像地址失败!!!")
			}
			// 创建CoreDNS
			shell = k8spath + "tools/hyperkube kubectl apply -f " + k8spath + "yaml/coredns.yaml"
			log.Info("创建CoreDNS " + shell)
			if !tools.ShellOut(shell) {
				log.Error("创建CoreDNS!!!")
			}
			time.Sleep(time.Second * 10)
			// 输出创建结果
			shell = k8spath + "tools/hyperkube kubectl get pod -o wide -n kube-system"
			log.Info("查看CoreDNS创建状态 " + shell)
		}
	}
}
