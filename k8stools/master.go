package k8stools

import (
	"sync"
	"time"

	myTools "go-install-kubernetes/tools"

	"github.com/op/go-logging"
	"github.com/pytool/ssh"
)

// 定义日志格式
var log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// InstallK8sMaster 安装k8s master
func InstallK8sMaster(ip, pwd, k8spath, nodeportrange, svcIP, etcdNodeList, clusterIP, nodeCidrLen string, ws *sync.WaitGroup) {
	log.Info(ip + pwd)
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	if !c.IsExist("/opt/kubernetes/cfg/") {
		err = c.MkdirAll("/opt/kubernetes/cfg/")
		if err != nil {
			log.Error(err)
		}
	}
	if !c.IsExist("/root/.kube/") {
		err = c.MkdirAll("/root/.kube/")
		if err != nil {
			log.Error(err)
		}
	}
	// 拷贝 kubernetes-csr.json 到/opt/kuberntes/cfg目录下
	log.Info(ip + "拷贝kubernetes-csr.json到/opt/kuberntes/cfg/目录")
	err = c.Upload(k8spath+"cert/kubernetes-csr.json", "/opt/kubernetes/cfg/kubernetes-csr.json")
	if err != nil {
		log.Info(err)
	}
	APISERVERSVCIP := myTools.GetIPString(svcIP)
	// 替换证书kubernetes-csr.json中的apiserverSVC地址
	err = c.Exec("sed -i 's/APISERVERSVCIP/" + APISERVERSVCIP + "1/g' /opt/kubernetes/cfg/kubernetes-csr.json")
	if err != nil {
		log.Error(err)
	}
	// 拷贝 hyperkube 到主机
	log.Info(ip + "拷贝hyperkube到/opt/kuberntes/bin/目录")
	err = c.Upload(k8spath+"tools/hyperkube", "/opt/kubernetes/bin/hyperkube")
	if err != nil {
		log.Info(err)
	}
	log.Info(ip + "拷贝cfssl到/opt/kuberntes/bin/目录")
	err = c.Upload(k8spath+"tools/cfssl", "/opt/kubernetes/bin/cfssl")
	if err != nil {
		log.Info(err)
	}
	log.Info(ip + "拷贝cfssljson到/opt/kuberntes/bin/目录")
	err = c.Upload(k8spath+"tools/cfssljson", "/opt/kubernetes/bin/cfssljson")
	if err != nil {
		log.Info(err)
	}
	log.Info(ip + "chmod 0755 /opt/kubernetes/bin/* 赋权")
	shell := "chmod 0755 /opt/kubernetes/bin/*"
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 分发config 可以不分发
	err = c.Upload("/root/.kube/config", "/root/.kube/config")
	if err != nil {
		log.Info(err)
	}
	// 分发证书和config
	// basic-auth.csv apiserver 基础认证（用户名/密码）配置
	certList := []string{"basic-auth.csv", "admin.pem", "admin-key.pem", "ca.pem", "ca-key.pem", "ca-config.json", "kube-proxy.kubeconfig", "kube-controller-manager.kubeconfig", "kube-scheduler.kubeconfig", "aggregator-proxy-csr.json"}
	for _, file := range certList {
		err = c.Upload(k8spath+"cert/"+file, "/opt/kubernetes/cfg/"+file)
		if err != nil {
			log.Info(err)
		}
	}
	// 修改config配置
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /opt/kubernetes/cfg/kubernetes-csr.json")
	if err != nil {
		log.Error(err)
	}
	// 创建 kubernetes 证书和私钥
	shell = "cd /opt/kubernetes/cfg/ && /opt/kubernetes/bin/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kubernetes-csr.json |  /opt/kubernetes/bin/cfssljson -bare kubernetes"
	// log.Info(shell)
	// if !myTools.ShellOut(shell) {
	// 	log.Error("创建 kubernetes 证书和私钥失败!!!")
	// }
	// 将本地执行修改为远程服务器上执行
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 判断文件是否创建 为创建直接退出
	filecheck, err := c.FileExist("/opt/kubernetes/cfg/kubernetes.pem")
	if err != nil || !filecheck {
		log.Info("kubernetes.pem和kubernetes-key.pem 创建完成")
	} else {
		log.Error(err)
		log.Error(filecheck)
		time.Sleep(60 * time.Second)
		log.Info("这条命令在" + ip + "这台机器上执行报错了，" + shell + "请检查，并在一分钟内完成执行手动执行。")
	}

	// 创建 aggregator proxy证书签名请求
	shell = "cd /opt/kubernetes/cfg/ && /opt/kubernetes/bin/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes aggregator-proxy-csr.json |  /opt/kubernetes/bin/cfssljson -bare aggregator-proxy"
	log.Info(ip + shell)
	// if !myTools.ShellOut(shell) {
	// 	log.Error(" aggregator proxy证书签名请求失败!!!")
	// }
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 判断文件是否创建 为创建直接退出
	filecheck, err = c.FileExist("/opt/kubernetes/cfg/aggregator-proxy.pem")
	if err != nil || !filecheck {
		log.Info("aggregator-proxy.aggregator-proxy-key.pem 创建完成")
	} else {
		log.Error(err)
		log.Error(filecheck)
		time.Sleep(60 * time.Second)
		log.Info("这条命令在" + ip + "这台机器上执行报错了，" + shell + "请检查，并在一分钟内完成执行手动执行。")
	}
	// 分发service文件
	service := []string{"kube-apiserver.service", "kube-controller-manager.service", "kube-scheduler.service"}
	for _, file := range service {
		err = c.Upload(k8spath+"service/"+file, "/etc/systemd/system/"+file)
		if err != nil {
			log.Info(err)
		}
	}
	err = c.Upload(k8spath+"yaml/basic-auth-rbac.yaml", "/opt/kubernetes/cfg/basic-auth-rbac.yaml")
	if err != nil {
		log.Info(err)
	}
	// 更改service配置
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /etc/systemd/system/kube-apiserver.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's/NODE_PORT_RANGE/" + nodeportrange + "/g' /etc/systemd/system/kube-apiserver.service")
	if err != nil {
		log.Error(err)
	}
	// 因为特殊字符将/换成%
	err = c.Exec("sed -i 's%SERVICE_CIDR%" + svcIP + "%g' /etc/systemd/system/kube-apiserver.service")
	if err != nil {
		log.Error(err)
	}
	// 因为特殊字符将/换成%
	err = c.Exec("sed -i 's%ETCD_ENDPOINTS%" + etcdNodeList + "%g' /etc/systemd/system/kube-apiserver.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%CLUSTER_CIDR%" + clusterIP + "%g' /etc/systemd/system/kube-controller-manager.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%NODE_CIDR_LEN%" + nodeCidrLen + "%g' /etc/systemd/system/kube-controller-manager.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%.*server.*%" + "    server: https://" + ip + ":6443" + "%g' /opt/kubernetes/cfg/kube-controller-manager.kubeconfig")
	if err != nil {
		log.Error(err)
	}
	shell = "sed -i 's%.*server.*%" + "    server: https://" + ip + ":6443" + "%g' /opt/kubernetes/cfg/kube-controller-manager.kubeconfig"
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%.*server.*%" + "    server: https://" + ip + ":6443" + "%g' /root/.kube/config")
	if err != nil {
		log.Error(err)
	}
	// 配置服务开机启动并启动
	for _, apiservername := range service {
		err = c.Exec("systemctl enable " + apiservername)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl daemon-reload")
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl restart " + apiservername)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl status " + apiservername)
		if err != nil {
			log.Error(err)
		}
	}

}

// RemoveK8sMaster 删除etcd集群
func RemoveK8sMaster(ip, pwd string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	service := []string{"kube-apiserver.service", "kube-controller-manager.service", "kube-scheduler.service"}
	for _, masterService := range service {
		err = c.Exec("systemctl disable " + masterService)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl stop " + masterService)
		if err != nil {
			log.Error(err)
		}
		log.Info("删除service文件")
		err = c.Exec("rm -rf /etc/systemd/system/" + masterService)
		if err != nil {
			log.Error(err)
		}
	}
	log.Info("删除所有配置文件与证书")
	err = c.Exec("rm -rf /opt/kubernetes/cfg")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl daemon-reload")
	if err != nil {
		log.Error(err)
	}
	log.Info("删除k8s master服务完成.")
}
