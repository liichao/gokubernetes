package k8stools

import (
	"sync"

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
func InstallK8sMaster(ip, pwd, k8spath string, ws *sync.WaitGroup) {
	log.Info(ip + pwd)
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	if !c.IsExist("/opt/kubenetes/cfg/") {
		err = c.MkdirAll("/opt/kubenetes/cfg/")
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
	// 拷贝 hyperkube 到主机
	err = c.Upload(k8spath+"tools/hyperkube", "/opt/kubernetes/bin/hyperkube")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"tools/cfssl", "/opt/kubernetes/bin/cfssl")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"tools/cfssl", "/opt/kubernetes/bin/cfssl")
	if err != nil {
		log.Info(err)
	}
	shell := "chmod 0755 /opt/kubernetes/bin/*"
	log.Info(shell)
	if !myTools.ShellOut(shell) {
		log.Error(ip + "chmod 0755 /opt/kubernetes/bin/* 失败!!!")
	}
	// 分发config 可以不分发
	err = c.Upload("/root/.kube/config", "/root/.kube/config")
	if err != nil {
		log.Info(err)
	}
	// 分发证书和config
	// basic-auth.csv apiserver 基础认证（用户名/密码）配置
	certList := []string{"basic-auth.csv", "admin.pem", "admin-key.pem", "ca.pem", "ca-key.pem", "ca-config.json", "kube-proxy.kubeconfig", "kube-controller-manager.kubeconfig", "kube-scheduler.kubeconfig"}
	for _, file := range certList {
		err = c.Upload(k8spath+"cert/"+file, "/opt/kubenetes/cfg/"+file)
		if err != nil {
			log.Info(err)
		}
	}
	// 修改config配置
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /opt/kubenetes/cfg/kubernetes-csr.json")
	if err != nil {
		log.Error(err)
	}
	// 创建 kubernetes 证书和私钥
	shell = "cd /opt/kubenetes/cfg/ && /opt/kubenetes/bin/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kubernetes-csr.json |  /opt/kubenetes/bin/cfssljson -bare kubernetes"
	log.Info(shell)
	if !myTools.ShellOut(shell) {
		log.Error(" 创建 kubernetes 证书和私钥失败!!!")
	}
	// 创建 aggregator proxy证书签名请求
	shell = "cd /opt/kubenetes/cfg/ && /opt/kubenetes/bin/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes aggregator-proxy-csr.json |  /opt/kubenetes/bin/cfssljson -bare aggregator-proxy"
	log.Info(shell)
	if !myTools.ShellOut(shell) {
		log.Error(" aggregator proxy证书签名请求失败!!!")
	}
	// err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /etc/systemd/system/etcd.service")
	// if err != nil {
	// 	log.Error(err)
	// }
}
