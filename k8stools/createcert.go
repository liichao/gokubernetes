package k8stools

import myTools "go-install-kubernetes/tools"

// CreateCert 创建相关证书
func CreateCert(k8spath, apiServer string) {
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
	kubeAPIServer := "https://" + apiServer + ":6443"
	log.Info(contextName)
	log.Info("开始创建k8s相关证书...")
	// 卸载ntp 并安装 chrony
	log.Info("给二进制赋权 cfssl cfssljson ")
	if !myTools.ShellOut("chmod +x " + k8spath + "tools/*") {
		log.Error("chmod error please check cfssl and cfssljson exist")
	}
	// 生成 CA 证书和私钥
	//myTools.ShellOut("cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -initca ca-csr.json |  " + k8spath + "tools/cfssljson -bare ca")
	shell := "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -initca ca-csr.json |  " + k8spath + "tools/cfssljson -bare ca"
	if !myTools.ShellOut(shell) {
		log.Error("create CA cert private key faild!!!")
	}
	// 创建admin证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes " + userName + "-csr.json |  " + k8spath + "tools/cfssljson -bare " + userName
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 删除原有kubeconfig
	shell = "rm -rf /root/.kube"
	if !myTools.ShellOut(shell) {
		log.Error("删除/root/.kube失败或者目录不存在!!")
	}
	// 设置集群参数admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/kubectl config set-cluster " + clusterName + " --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer
	log.Info("start run " + shell)
	if !myTools.ShellOut(shell) {
		log.Error("检查路径与kubectl文件是否存在!!!")
	}
	// 设置客户端认证参数admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/kubectl config set-credentials " + userName + " --client-certificate=" + k8spath + "cert/" + userName + ".pem --embed-certs=true --client-key=" + k8spath + "cert/" + userName + "-key.pem"
	log.Info("start run " + shell)
	if !myTools.ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}
	// 设置上下文参数admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/kubectl config set-context " + contextName + " --cluster=" + clusterName + " --user=" + userName
	log.Info("start run " + shell)
	if !myTools.ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}
	// 选择默认上下文admin
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/kubectl config set-context " + contextName
	log.Info("start run " + shell)
	if !myTools.ShellOut(shell) {
		log.Error("del faild or config file not exist!!!")
	}

	// 创建 kube-scheduler证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-scheduler-csr.json | " + k8spath + "tools/cfssljson -bare kube-scheduler"
	log.Info("start run " + shell)
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置集群参数kube-scheduler
	shell = k8spath + "tools/kubectl config set-cluster kubernetes --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer + " --kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置认证参数kube-scheduler
	shell = k8spath + "tools/kubectl config set-credentials system:kube-scheduler --client-certificate=" + k8spath + "cert/kube-scheduler.pem --client-key=" + k8spath + "cert/kube-scheduler-key.pem --embed-certs=true --server=" + kubeAPIServer + " --kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置上下文参数kube-scheduler
	shell = k8spath + "tools/kubectl config set-context default --cluster=kubernetes --user=system:kube-scheduler --kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 选择默认上下文kube-scheduler
	shell = k8spath + "tools/kubectl config use-context default --kubeconfig=" + k8spath + "cert/kube-scheduler.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}

	//创建 kube-proxy证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-proxy-csr.json | " + k8spath + "tools/cfssljson -bare kube-proxy"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置集群参数kube-proxy
	shell = k8spath + "tools/kubectl config set-cluster kubernetes --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer + " --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置客户端认证参数kube-proxy
	shell = k8spath + "tools/kubectl config set-credentials kube-proxy --client-certificate=" + k8spath + "cert/kube-proxy.pem --client-key=" + k8spath + "cert/kube-proxy-key.pem --embed-certs=true --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置上下文参数kube-proxy
	shell = k8spath + "tools/kubectl config set-context default --cluster=kubernetes --user=kube-proxy --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 选择默认上下文kube-proxy
	shell = k8spath + "tools/kubectl config use-context default --kubeconfig=" + k8spath + "cert/kube-proxy.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 创建 kube-controller-manager证书与私钥
	shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-controller-manager-csr.json | " + k8spath + "tools/cfssljson -bare kube-controller-manager"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置集群参数kube-controller-manager
	shell = k8spath + "tools/kubectl config set-cluster kubernetes --certificate-authority=" + k8spath + "cert/ca.pem --embed-certs=true --server=" + kubeAPIServer + " --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置认证参数kube-controller-manager
	shell = k8spath + "tools/kubectl config set-credentials system:kube-controller-manager --client-certificate=" + k8spath + "cert/kube-controller-manager.pem --client-key=" + k8spath + "cert/kube-controller-manager-key.pem --embed-certs=true --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 设置上下文参数kube-controller-manager
	shell = k8spath + "tools/kubectl config set-context default --cluster=kubernetes --user=system:kube-controller-manager --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 选择默认上下文kube-controller-manager
	shell = k8spath + "tools/kubectl config use-context default --kubeconfig=" + k8spath + "cert/kube-controller-manager.kubeconfig"
	if !myTools.ShellOut(shell) {
		log.Error("create admin cert private key faild!!!")
	}
	// 创建flanneld的证书
	// shell = "cd " + k8spath + "cert/ && " + k8spath + "tools/cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes flanneld-csr.json | " + k8spath + "tools/cfssljson -bare flanneld"
	// if !myTools.ShellOut(shell) {
	// 	log.Error("创建flanneld证书失败!!!")
	// }
}
