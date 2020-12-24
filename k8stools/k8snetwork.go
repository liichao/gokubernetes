package k8stools

import (
	"sync"

	"github.com/pytool/ssh"
)

// InstallK8sNetwork 安装k8s node
func InstallK8sNetwork(ip, pwd, k8spath, flannelBackend, harborURL, harborUser, harborPwd, pauseImage string, ws *sync.WaitGroup) {
	log.Info(ip + pwd + flannelBackend)
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// 创建相关目录
	flannelPathList := []string{"/etc/cni/net.d", "/opt/kubernetes/images", "/opt/kubernetes/cfg"}
	for _, path := range flannelPathList {
		if !c.IsExist(path) {
			err = c.MkdirAll(path)
			if err != nil {
				log.Error(err)
			}
		}
	}
	// 上传证书
	// flannelCertList := []string{"flanneld.csr", "flanneld-csr.json", "flanneld-key.pem", "flanneld.pem"}
	// for _, flannelCert := range flannelCertList {
	// 	log.Info(flannelCert)
	// 	err = c.Upload(k8spath+"cert/"+flannelCert, "/opt/kubernetes/cfg/cert/"+flannelCert)
	// 	if err != nil {
	// 		log.Info(err)
	// 	}
	// }
	// 上传二进制文件
	flannelBinList := []string{"bridge", "flannel", "host-local", "loopback", "portmap"}
	for _, flannelBin := range flannelBinList {
		log.Info(flannelBin)
		err = c.Upload(k8spath+"tools/"+flannelBin, "/opt/kubernetes/bin/")
		if err != nil {
			log.Info(err)
		}
	}
	// 赋权
	shell := "chmod 0755 /opt/kubernetes/bin/*"
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// 删除默认cni配置
	shell = "rm -rf /etc/cni/net.d/cni-default.conf"
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	// // 释放service文件
	// err = c.Upload(k8spath+"service/flanneld.service", "/etc/systemd/system/flanneld.service")
	// if err != nil {
	// 	log.Info(err)
	// }
	// 与仓库建立连接
	shell = "docker login -u " + harborUser + " -p " + harborPwd + " " + harborURL
	log.Info(shell)
	err = c.Exec(shell)
	if err != nil {
		log.Error(err)
	}
	shell = "docker pull " + pauseImage
	log.Info(ip + " 执行: " + shell)
	err = c.Exec(shell)
	if err != nil {
		log.Info(err)
	}
}
