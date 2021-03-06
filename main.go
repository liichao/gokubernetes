package main

import (
	"container/list"
	"encoding/json"
	"go-install-kubernetes/k8stools"
	"go-install-kubernetes/tools"
	"io/ioutil"
	"os"
	"strconv"
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

// 程序的开始
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
	clusterDNSDomain := config.Get("clusterDnsDomain").(string)
	iptablesBinName := config.Get("repair.iptablesBinName").(string)
	// 相关文件
	confLists := []string{"cni-default.conf", "10-k8s-modules.conf", "95-k8s-sysctl.conf", "30-k8s-ulimits.conf", "sctp.conf", "server-centos.conf", "daemon.json", "docker"}
	certFileLists := []string{"kubelet-csr.json", "kubernetes-csr.json", "basic-auth.csv", "aggregator-proxy-csr.json", "etcd-csr.json", "admin-csr.json", "ca-config.json", "ca-csr.json", "kube-controller-manager-csr.json", "kube-proxy-csr.json", "kube-scheduler-csr.json", "read-csr.json"}
	yamlFileLists := []string{"read-user-sa-rbac.yaml", "kubernetes-dashboard.yaml", "admin-user-sa-rbac.yaml", "metrics-server.yaml", "coredns.yaml", "kube-flannel-vxlan.yaml", "kube-flannel.yaml", "read-group-rbac.yaml", "basic-auth-rbac.yaml", "kubelet-config.yaml"}
	serviceFileLists := []string{"kubelet.service", "kube-proxy.service", "kube-apiserver.service", "kube-scheduler.service", "kube-controller-manager.service", "etcd.service", "docker.service"}
	toolsFileLists := []string{"cfssl", "cfssljson", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler", "kubelet", "kubectl", "etcd.tar.gz", "cni-plugins-linux-amd64-v0.9.0.tgz", "flanneld-v0.13.0-amd64.docker", "iptables-1.6.2.tar.bz2"}
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
	// 给 k8spath目录下文件赋权
	// 修改docker配置文件，将harbor地址加入到信任地址
	shell = "chmod u+x " + k8spath + "/*"
	if !tools.ShellOut(shell) {
		log.Error("赋权失败...")
	}
	// 判断pause镜像文件是否存在
	if !tools.Exists(k8spath + "tools/" + config.Get("cnitools").(string)) {
		log.Info(config.Get("cnitools").(string) + "镜像包不存在，请上传到" + k8spath + "tools/目录下")
	}
	//解压
	if !tools.ShellOut("cd " + k8spath + "/tools && tar -zxf " + config.Get("cnitools").(string)) {
		log.Error("解压" + config.Get("cnitools").(string) + "失败")
	}
	// 创建并发
	var wg sync.WaitGroup
	// 分发bin二进制文件与所有相关的证书文件等。
	wg.Add(allIPList.Len())
	for ip := allIPList.Front(); ip != nil; ip = ip.Next() {
		log.Info(ip.Value.(string))
		// 主要把核心三个文件分发好 cfssl cfssljson k8s五大中心组件 在确认还会不会出现Text file busy
		FileLists := []string{"cfssl", "cfssljson", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler", "kubelet", "kubectl"}
		go tools.SendBinAndConfigFile(ip.Value.(string), password, k8spath, "/opt/kubernetes/bin/", FileLists, &wg)
	}
	wg.Wait()
	log.Info("所有文件分发完成请确认")
	time.Sleep(10 * time.Second)
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
		log.Info("系统参数配置完成.请确认，如有问题请ctrl+C结束进程，之后修改配置文件system配置为flase")
		time.Sleep(10 * time.Second)
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
		log.Info("安装时间同步服务器完成，如有问题请ctrl+C结束进程，之后修改配置文件chrony配置为flase")
		time.Sleep(10 * time.Second)
	}
	log.Info("判断是否要创建证书...")
	if createCert {
		log.Info(" 开始在当前主机创建相关证书...")
		k8stools.CreateCert(k8spath, apiServer)
		log.Info("创建证书完成，如有问题请ctrl+C结束进程，之后修改配置文件createCert配置为flase")
		time.Sleep(10 * time.Second)
	}
	log.Info("判断是否安装etcd集群...")
	if etcd {
		wg.Add(masterIPList.Len())
		// install etcd
		if etcdInstall {
			var etcd EtcdJSONParse
			log.Info("开始安装Etcd服务....")
			// unzip etcd.tar.gz
			if !tools.Exists(k8spath + "tools/etcd.tar.gz") {
				log.Warning("etcd.tar.gz文件不存在，请上传到" + k8spath + "tools/目录下")
				log.Warning("下载URL : https://github.com/etcd-io/etcd/releases/download/v3.4.14/etcd-v3.4.14-linux-amd64.tar.gz")
			}
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
		log.Info("创建etcd集群，如有问题请ctrl+C结束进程，之后修改配置文件etcdInstall配置为flase")
		time.Sleep(10 * time.Second)
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
		log.Info("创建Docker完成，如有问题请ctrl+C结束进程，之后修改配置文件docker配置为flase")
		time.Sleep(10 * time.Second)
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
		err = c.Exec("/opt/kubernetes/bin/kubectl apply -f /opt/kubernetes/cfg/basic-auth-rbac.yaml")
		if err != nil {
			log.Error(err)
		}
		log.Info("创建k8s api集群完成，如有问题请ctrl+C结束进程，之后修改配置文件installK8sApi配置为flase")
		time.Sleep(10 * time.Second)
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
		log.Info("卸载k8s api集群完成，如有问题请ctrl+C结束进程，之后修改配置文件removeK8sApi配置为flase")
		time.Sleep(10 * time.Second)
	}
	log.Info("判断是否安装k8s Node服务...")
	if config.Get("installK8sNode").(bool) {
		log.Info("开始安装node相关服务")
		wg.Add(nodeIPList.Len())
		for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
			go k8stools.InstallK8sNode(ip.Value.(string), password, svcIP, k8spath, apiServer, strconv.Itoa(maxPods), clusterIP, proxyMode, pauseImage, clusterDNSDomain, &wg)
		}
		wg.Wait()
		log.Info("k8s node kubelet kube-proxy组件安装完成.")
		log.Info("创建k8s node完成，如有问题请ctrl+C结束进程，之后修改配置文件installK8sNode配置为flase")
		time.Sleep(10 * time.Second)
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
		shell := k8spath + "tools/kubectl create secret docker-registry myregistrykey --docker-server=" + config.Get("harborUrl").(string) + " --docker-username=" + config.Get("harborUser").(string) + " --docker-password=" + config.Get("harborPwd").(string) + " --docker-email=dzero@dero.com -n kube-system"
		if !tools.ShellOut(shell) {
			log.Error("创建kube-system registry secret失败!!!")
		}
		shell = k8spath + "tools/kubectl create secret docker-registry myregistrykey --docker-server=" + config.Get("harborUrl").(string) + " --docker-username=" + config.Get("harborUser").(string) + " --docker-password=" + config.Get("harborPwd").(string) + " --docker-email=dzero@dero.com"
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
			// 修改镜像flannel tag 并推送到docker 仓库
			//shell = "docker tag quay.io/coreos/" + strings.Split(config.Get("flanneldImageOffline").(string), ".doc")[0] + " " + flanneldImage
			shell = "docker tag $(docker images |grep flannel |head -n 1 |awk '{print $3}')" + " " + flanneldImage
			log.Info(shell)
			err = c.Exec(shell)
			if err != nil {
				log.Error(err)
			}
			// 修改pause 镜像tag  并推送到仓库
			shell = "docker tag $(docker images |grep pause-amd64 |head -n 1 |awk '{print $3}')" + " " + pauseImage
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
			go k8stools.InstallK8sNetwork(ip.Value.(string), password, k8spath, flannelBackend, harborURL, harborUser, harborPwd, pauseImage, &wg)
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
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/kube-flannel-vxlan.yaml"
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
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/kube-flannel.yaml"
			if !tools.ShellOut(shell) {
				log.Error("创建 flannled失败!!!")
			}
		}
		log.Info("创建flannled完成，如有问题请ctrl+C结束进程，之后修改配置文件network.install配置为flase")
		time.Sleep(10 * time.Second)
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
				// 判断metricsserver_offline镜像文件是否存在
				loadImages := tools.LoadImagesChangeTagPushImages(nodeIPList.Front().Value.(string), password, k8spath, harborURL, harborUser, harborPwd, cordnsTar, coreDNSImages, "coredns")
				if loadImages {
					log.Info(cordnsTar + "镜像导入完成")
				}
			}
			shell := "sed -i 's%CLUSTER_DNS_SVC_IP%" + config.Get("other.dns.dnsIP").(string) + "%g' " + k8spath + "yaml/coredns.yaml"
			log.Info("替换dnsIP  " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换dnsIP失败!!!")
			}
			shell = "sed -i 's%CLUSTER_DNS_DOMAIN%" + clusterDNSDomain + "%g' " + k8spath + "yaml/coredns.yaml"
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
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/coredns.yaml"
			log.Info("创建CoreDNS " + shell)
			if !tools.ShellOut(shell) {
				log.Error("创建CoreDNS!!!")
			}
			time.Sleep(time.Second * 10)
			// 输出创建结果
			shell = k8spath + "tools/kubectl get pod -o wide -n kube-system"
			log.Info("查看CoreDNS创建状态 " + shell)
		}
		// 判断是否安装metricsserver
		if config.Get("other.metricsserver.install").(bool) {
			metricsServerImages := config.Get("other.metricsserver.metricsServerImages").(string)
			metricsserverOffline := config.Get("other.metricsserver.metricsserver_offline").(string)
			// 判断是否需要载入镜像
			if config.Get("other.metricsserver.loadImage").(bool) {
				// 判断metricsserver_offline镜像文件是否存在
				loadImages := tools.LoadImagesChangeTagPushImages(nodeIPList.Front().Value.(string), password, k8spath, harborURL, harborUser, harborPwd, metricsserverOffline, metricsServerImages, "metrics-server")
				if loadImages {
					log.Info(metricsserverOffline + "镜像导入完成")
				}
			}
			// 替换yaml镜像地址
			shell = "sed -i 's%metricsServerImages%" + metricsServerImages + "%g' " + k8spath + "yaml/metrics-server.yaml"
			log.Info("替换镜像地址 " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换镜像地址失败!!!")
			}
			// 创建metrics-server
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/metrics-server.yaml"
			log.Info("创建metrics-server " + shell)
			if !tools.ShellOut(shell) {
				log.Error("创建metrics-server!!!")
			}
		}
		// 判断是否安装dashboard
		if config.Get("other.dashboard.install").(bool) {
			// todo-list
			dashboardImages := config.Get("other.dashboard.dashboardImages").(string)
			dashboardOffline := config.Get("other.dashboard.dashboard_offline").(string)
			metricsscraperImages := config.Get("other.dashboard.metricsscraperImages").(string)
			metricsscraperOffline := config.Get("other.dashboard.metricsscraper_offline").(string)
			// 判断是否需要载入镜像
			if config.Get("other.dashboard.loadImage").(bool) {
				// 判断dashboardOffline镜像文件是否存在
				loadImages := tools.LoadImagesChangeTagPushImages(nodeIPList.Front().Value.(string), password, k8spath, harborURL, harborUser, harborPwd, dashboardOffline, dashboardImages, "dashboard")
				if loadImages {
					log.Info(dashboardOffline + "镜像导入完成")
				}
				// 判断metricsscraperImages镜像文件是否存在
				loadImages = tools.LoadImagesChangeTagPushImages(nodeIPList.Front().Value.(string), password, k8spath, harborURL, harborUser, harborPwd, metricsscraperOffline, metricsscraperImages, "metrics-scraper")
				if loadImages {
					log.Info(metricsscraperOffline + "镜像导入完成")
				}
			}
			// 替换yaml镜像地址
			shell = "sed -i 's%metricsscraperImages%" + metricsscraperImages + "%g' " + k8spath + "yaml/kubernetes-dashboard.yaml"
			log.Info("替换镜像地址 " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换镜像地址失败!!!")
			}
			shell = "sed -i 's%dashboardImages%" + dashboardImages + "%g' " + k8spath + "yaml/kubernetes-dashboard.yaml"
			log.Info("替换镜像地址 " + shell)
			if !tools.ShellOut(shell) {
				log.Error("替换镜像地址失败!!!")
			}
			// 创建dashborad
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/admin-user-sa-rbac.yaml"
			log.Info("创建admin-user-sa-rbac " + shell)
			if !tools.ShellOut(shell) {
				log.Error("创建admin-user-sa-rbac!!!")
			}
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/read-user-sa-rbac.yaml"
			log.Info("创建read-user-sa-rbac " + shell)
			if !tools.ShellOut(shell) {
				log.Error("创建read-user-sa-rbac!!!")
			}
			shell = k8spath + "tools/kubectl apply -f " + k8spath + "yaml/kubernetes-dashboard.yaml"
			log.Info("创建kubernetes-dashboard " + shell)
			if !tools.ShellOut(shell) {
				log.Error("创建kubernetes-dashboard!!!")
			}
		}
	}
	// 将apiServer地址修改成ipvsApiServer地址
	// 替换俩个服务的配置文件并重启 kubelet.kubeconfig kube-proxy.kubeconfig
	if config.Get("ipvsApiServer").(bool) {
		apiServer = apiServer + ":6443"
		ipvsAPIServer := tools.GetIPString(svcIP) + "1" + ":443"
		log.Info(apiServer)
		log.Info(ipvsAPIServer)
		wg.Add(nodeIPList.Len())
		for ip := nodeIPList.Front(); ip != nil; ip = ip.Next() {
			go k8stools.UpdateAPIServerToIpvs(ip.Value.(string), password, apiServer, ipvsAPIServer, &wg)
		}
		wg.Wait()
		log.Info("检查服务")
	}
	log.Info("所有程序都安全完成，请执行systemctl status kubelet 如Not using `--random-fully` in the MASQUERADE rule for iptables because the local version of iptables does not support it 报错，请将config.yaml false修改为true 在运行修复一下")
}
