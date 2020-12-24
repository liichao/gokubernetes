package k8stools

import (
	"sync"
	"time"

	"github.com/pytool/ssh"
)

// InstallDocker 安装docker
func InstallDocker(ip, pwd, k8spath string, ws *sync.WaitGroup) {
	log.Info(ip + pwd)
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// 检查并创建目录
	if !c.IsExist("/etc/docker") {
		err = c.MkdirAll("/etc/docker")
		if err != nil {
			log.Error(err)
		}
	}
	err = c.Upload(k8spath+"daemon.json", "/etc/docker/")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"docker", "/etc/bash_completion.d/docker")
	if err != nil {
		log.Info(err)
	}
	err = c.Exec("chmod 0644 /etc/bash_completion.d/docker")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("rm -rf /opt/kubernetes/bin/docker")
	if err != nil {
		log.Error(err)
	}
	// 传输docker相关文件到 /opt/kubernetes/bin/docker/目录下
	log.Info(ip + ":" + k8spath + "tools/docker to " + "/opt/kubernetes/bin/")
	err = c.UploadDir(k8spath+"tools/docker", "/opt/kubernetes/bin/")
	if err != nil {
		log.Info(err)
	}
	// 赋权
	err = c.Exec("chmod 755 /opt/kubernetes/bin/docker/*")
	if err != nil {
		log.Error(err)
	}
	// flush-iptables 清空iptables
	err = c.Exec("iptables -P INPUT ACCEPT && iptables -F && iptables -X && iptables -F -t nat && iptables -X -t nat && iptables -F -t raw && iptables -X -t raw && iptables -F -t mangle && iptables -X -t mangle")
	if err != nil {
		log.Error(err)
	}
	// ln -s /opt/kubernetes/bin/docker/docker /usr/bin/docker 创建软连接
	err = c.Exec("ln -s /opt/kubernetes/bin/docker/docker /usr/bin/docker")
	if err != nil {
		log.Error(err)
	}
	// 传输 docker启动文件到 /etc/systemd/system/docker.service
	err = c.Upload(k8spath+"service/docker.service", "/etc/systemd/system/docker.service")
	if err != nil {
		log.Info(err)
	}
	// 配置docker开机启动与启动
	err = c.Exec("systemctl enable docker")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl daemon-reload")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl restart docker")
	if err != nil {
		log.Error(err)
	}
	// 延迟一秒
	time.Sleep(1000000000)
	err = c.Exec("systemctl status docker.service")
	if err != nil {
		log.Error(err)
	}
}

// RemoveDocker 删除etcd集群
func RemoveDocker(ip, pwd string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	log.Info("Disable startup docker service.")
	err = c.Exec("systemctl disable docker")
	if err != nil {
		log.Error(err)
	}
	log.Info("Stop docker service.")
	err = c.Exec("systemctl stop docker")
	if err != nil {
		log.Error(err)
	}
	log.Info("delete /var/lib/docker")
	err = c.Exec("rm -rf /var/lib/docker")
	if err != nil {
		log.Error(err)
	}
	log.Info("delete /etc/docker/")
	err = c.Exec("rm -rf /etc/docker/")
	if err != nil {
		log.Error(err)
	}
	log.Info("delete /etc/systemd/system/docker.service")
	err = c.Exec("rm -rf /etc/systemd/system/docker.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl daemon-reload")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("rm -rf /usr/bin/docker")
	if err != nil {
		log.Error(err)
	}
	log.Info("Remove docker Service Done.")
}

// LoginDockerHarbor 每个node节点都登陆一下harbor仓库  已废弃 写到了ChangeHarborHost函数中
// func LoginDockerHarbor(ip, pwd, harborURL, harborPwd, harborUser string, ws *sync.WaitGroup) {
// 	defer ws.Done()
// 	c, err := ssh.NewClient(ip, "22", "root", pwd)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer c.Close()

// }
