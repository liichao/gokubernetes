package k8stools

import (
	"sync"

	"github.com/pytool/ssh"
)

// InstallEtcd 安装etcd集群
func InstallEtcd(ip, pwd, k8spath, etcdID, etcdnode string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// 检查并创建ssl目录
	if !c.IsExist("/etc/etcd/ssl") {
		err = c.MkdirAll("/etc/etcd/ssl")
		if err != nil {
			log.Error(err)
		}
	}
	if !c.IsExist("/opt/kubernetes/bin/") {
		err = c.MkdirAll("/opt/kubernetes/bin/")
		if err != nil {
			log.Error(err)
		}
	}
	// 检查并创建etcd目录
	if !c.IsExist("/var/lib/etcd/") {
		err = c.MkdirAll("/var/lib/etcd/")
		if err != nil {
			log.Error(err)
		}
	}
	// 拷贝etcd二进制文件到 相关目录
	err = c.Upload(k8spath+"tools/etcd/etcd", "/opt/kubernetes/bin/etcd")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"tools/etcd/etcdctl", "/opt/kubernetes/bin/etcdctl")
	if err != nil {
		log.Info(err)
	}
	// 拷贝证书到相关目录
	err = c.Upload(k8spath+"cert/etcd.pem", "/etc/etcd/ssl/")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"cert/etcd-key.pem", "/etc/etcd/ssl/")
	if err != nil {
		log.Info(err)
	}
	err = c.Upload(k8spath+"cert/ca.pem", "/etc/etcd/ssl/")
	if err != nil {
		log.Info(err)
	}
	// 给所有二进制文件授权
	err = c.Exec("chmod 755 /opt/kubernetes/bin/*")
	if err != nil {
		log.Error(err)
	}
	// 拷贝etcd.service 启动文件到目录
	err = c.Upload(k8spath+"service/etcd.service", "/etc/systemd/system/")
	if err != nil {
		log.Info(err)
	}
	// 修改etcd启动文件
	log.Info(etcdID)
	err = c.Exec("sed -i 's/NODE_NAME/etcd" + etcdID + "/g' /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's/inventory_hostname/" + ip + "/g' /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("sed -i 's%ETCD_NODES%" + etcdnode + "%g' /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	// 配置etcd开机启动并启动
	err = c.Exec("systemctl enable etcd")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl daemon-reload")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl restart etcd")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl status etcd.service")
	if err != nil {
		log.Error(err)
	}
}

// RemoveEtcd 删除etcd集群
func RemoveEtcd(ip, pwd string, ws *sync.WaitGroup) {
	defer ws.Done()
	c, err := ssh.NewClient(ip, "22", "root", pwd)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	log.Info("Disable startup etcd service.")
	err = c.Exec("systemctl disable etcd")
	if err != nil {
		log.Error(err)
	}
	log.Info("Stop etcd service.")
	err = c.Exec("systemctl stop etcd")
	if err != nil {
		log.Error(err)
	}
	log.Info("delete /var/lib/etcd")
	err = c.Exec("rm -rf /var/lib/etcd")
	if err != nil {
		log.Error(err)
	}
	log.Info("delete /etc/etcd/")
	err = c.Exec("rm -rf /etc/etcd/")
	if err != nil {
		log.Error(err)
	}
	log.Info("delete /etc/systemd/system/etcd.service")
	err = c.Exec("rm -rf /etc/systemd/system/etcd.service")
	if err != nil {
		log.Error(err)
	}
	err = c.Exec("systemctl daemon-reload")
	if err != nil {
		log.Error(err)
	}
	log.Info("Remove Etcd Service Done.")
}
