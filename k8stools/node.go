package k8stools

import (
	"go-install-kubernetes/tools"
	"strings"
	"sync"

	"github.com/pytool/ssh"
)

// InstallK8sNode 安装k8s node
func InstallK8sNode(ip, pwd, svcIP, k8spath, apiserver, maxPods, clusterIP, proxyMode, pauseImage string, ws *sync.WaitGroup) {
	log.Info(ip + pwd)
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// 更改主机名
	err = c.Exec("echo '" + "node-" + strings.ReplaceAll(ip, ".", "-") + "' > /etc/hostname")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("hostname " + "node-" + strings.ReplaceAll(ip, ".", "-"))
	if err != nil {
		log.Error(err)
	}
	// 创建目录
	kubeListPath := []string{"/opt/kubernetes/cfg/cert/", "/root/.kube/", "/var/lib/kubelet/", "/var/lib/kube-proxy/", "/etc/cni/net.d/", "/opt/kubernetes/bin/"}
	for _, kubepath := range kubeListPath {
		if !c.IsExist(kubepath) {
			err = c.MkdirAll(kubepath)
			if err != nil {
				log.Error(err)
			}
		}
	}
	// 上传cni配置文件
	err = c.Upload(k8spath+"cni-default.conf", "/etc/cni/net.d/")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("sed -i 's%CLUSTER_CIDR%" + clusterIP + "%g' /etc/cni/net.d/cni-default.conf")
	if err != nil {
		log.Error(err)
	}
	// 分发文件
	binFileList := []string{"hyperkube", "cfssl", "cfssljson", "bridge", "host-local", "loopback"}
	for _, binFile := range binFileList {
		fileName := binFile[strings.LastIndex(binFile, `/`)+1:]
		log.Info(fileName)
		err = c.Upload(k8spath+"tools/"+binFile, "/opt/kubernetes/bin/"+fileName)
		if err != nil {
			log.Info(err)
		}
	}
	// 给bin目录赋权
	err = c.Exec("chmod u+x /opt/kubernetes/bin/*")
	if err != nil {
		log.Error(err)
	}
	// 传输config
	err = c.Upload("/root/.kube/config", "/root/.kube/config")
	if err != nil {
		log.Info(err)
	}
	// 修改apiserver地址
	err = c.Exec("sed -i 's%.*server.*%" + "    server: https://" + apiserver + ":6443" + "%g' /root/.kube/config")
	if err != nil {
		log.Error(err)
	}
	// 创建 kubelet 相关证书及 kubelet.kubeconfig
	//kubelet-csr.json
	certListFile := []string{k8spath + "cert/kubelet-csr.json", k8spath + "cert/ca.pem", k8spath + "cert/ca-config.json", k8spath + "cert/ca-key.pem"}
	for _, certFile := range certListFile {
		err = c.Upload(certFile, "/opt/kubernetes/cfg/cert/")
		if err != nil {
			log.Info(err)
		}
	}
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /opt/kubernetes/cfg/cert/kubelet-csr.json")
	if err != nil {
		log.Error(err)
	}
	// 创建 kubelet 证书与私钥
	shell := "cd /opt/kubernetes/cfg/cert && /opt/kubernetes/bin/cfssl gencert -ca=/opt/kubernetes/cfg/cert/ca.pem -ca-key=/opt/kubernetes/cfg/cert/ca-key.pem -config=/opt/kubernetes/cfg/cert/ca-config.json -profile=kubernetes kubelet-csr.json | /opt/kubernetes/bin/cfssljson -bare kubelet"
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 设置集群参数
	shell = "/opt/kubernetes/bin/hyperkube kubectl config set-cluster kubernetes --certificate-authority=/opt/kubernetes/cfg/cert/ca.pem --embed-certs=true --server=https://" + apiserver + ":6443 --kubeconfig=/opt/kubernetes/cfg/kubelet.kubeconfig"
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	//设置客户端认证参数
	shell = "/opt/kubernetes/bin/hyperkube kubectl config set-credentials system:node:" + ip + " --client-certificate=/opt/kubernetes/cfg/cert/kubelet.pem --embed-certs=true --client-key=/opt/kubernetes/cfg/cert/kubelet-key.pem --kubeconfig=/opt/kubernetes/cfg/kubelet.kubeconfig"
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 设置上下文参数
	shell = "/opt/kubernetes/bin/hyperkube kubectl config set-context default --cluster=kubernetes --user=system:node:" + ip + " --kubeconfig=/opt/kubernetes/cfg/kubelet.kubeconfig"
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 选择默认上下文
	shell = "/opt/kubernetes/bin/hyperkube kubectl config use-context default --kubeconfig=/opt/kubernetes/cfg/kubelet.kubeconfig"
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 分发kube-proxy文件与证书
	err = c.Exec("rm -rf  /opt/kubernetes/cfg/cert/kube-proxy*")
	if err != nil {
		log.Error(err)
	}
	nodeCertFile := []string{"kube-proxy.csr", "kube-proxy-csr.json", "kube-proxy-key.pem", "kube-proxy.pem"}
	for _, certFile := range nodeCertFile {
		err = c.Upload(k8spath+"cert/"+certFile, "/opt/kubernetes/cfg/cert/")
		if err != nil {
			log.Info(err)
		}
	}
	err = c.Upload(k8spath+"cert/kube-proxy.kubeconfig", "/opt/kubernetes/cfg/")
	if err != nil {
		log.Info(err)
	}
	//分发kubelet-config.yaml 文件
	err = c.Upload(k8spath+"yaml/kubelet-config.yaml", "/opt/kubernetes/cfg/")
	if err != nil {
		log.Info(err)
	}
	// 修改配置kubelet-config.yaml
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /opt/kubernetes/cfg/kubelet-config.yaml")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's/CLUSTER_DNS_SVC_IP/" + tools.GetIPString(svcIP) + "2" + "/g' /opt/kubernetes/cfg/kubelet-config.yaml")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's/MAX_PODS/" + maxPods + "/g' /opt/kubernetes/cfg/kubelet-config.yaml")
	if err != nil {
		log.Error(err)
	}
	//// 开始释放service文件,并进行配置
	nodeServerFile := []string{"kube-proxy.service", "kubelet.service"}
	for _, serviceFile := range nodeServerFile {
		err = c.Upload(k8spath+"service/"+serviceFile, "/etc/systemd/system/")
		if err != nil {
			log.Info(err)
		}
		err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /etc/systemd/system/" + serviceFile)
		if err != nil {
			log.Error(err)
		}
	}
	err = c.Exec("sed -i 's%CLUSTER_CIDR%" + clusterIP + "%g' /etc/systemd/system/kube-proxy.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's/PROXY_MODE/" + proxyMode + "/g' /etc/systemd/system/kube-proxy.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%SANDBOX_IMAGE%" + pauseImage + "%g' /etc/systemd/system/kubelet.service")
	if err != nil {
		log.Error(err)
	}
	// 替换kube-proxy.kubeconfig中的api地址
	err = c.Exec("sed -i 's%.*server.*%" + "    server: https://" + apiserver + ":6443" + "%g' /opt/kubernetes/cfg/kube-proxy.kubeconfig")
	if err != nil {
		log.Error(err)
	}
	// 配置服务开机启动并启动
	for _, nodeService := range nodeServerFile {
		err = c.Exec("systemctl enable " + nodeService)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl daemon-reload")
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl restart " + nodeService)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl status " + nodeService)
		if err != nil {
			log.Error(err)
		}
	}
}

// RemoveK8sNode 删除kubelet和kube-proxy
func RemoveK8sNode(ip, pwd string, ws *sync.WaitGroup) {
	log.Info(ip + pwd)
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	//// 开始释放service文件,并进行配置
	nodeServerFile := []string{"kube-proxy.service", "kubelet.service"}
	for _, serviceFile := range nodeServerFile {
		err = c.Exec("systemctl disable " + serviceFile)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl stop " + serviceFile)
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("systemctl daemon-reload")
		if err != nil {
			log.Error(err)
		}
		err = c.Exec("rm -rf /etc/systemd/system/" + serviceFile)
		if err != nil {
			log.Error(err)
		}
	}
	err = c.Exec("rm -rf /opt/kubernetes/cfg")
	if err != nil {
		log.Error(err)
	}
}
