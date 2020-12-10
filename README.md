# gokubernetes
用golang编写一个 一键部署k8s集群

### 1、安装时钟同步服务器
    go-install-kubernetes para=chrony ips=10.10.77.202-204 pwd=密码 ntpserver=ntp.aliyun.com

### 2、配置系统参数
    go-install-kubernetes para=system ips=10.10.77.202-204 pwd=密码 proxymode=ipvs

### 3、创建相关证书(k8s、etcd)
    go-install-kubernetes para=createcert ips=10.10.77.202-204 pwd=密码

### 4、创建etcd集群
#### 安装
    go-install-kubernetes para=etcd ips=10.10.77.202-204 pwd=密码 handle=install

#### 清理
    go-install-kubernetes para=etcd ips=10.10.77.202-204 pwd=密码 handle=uninstall

描述信息清理etcd集群

    # systemctl stop etcd 
    # systemctl disable etcd
    # rm -rf /var/lib/etcd
    # rm -rf /etc/etcd/
    # rm -rf /etc/systemd/system/etcd.service

### 5、docker
#### 安装
    go-install-kubernetes para=docker ips=10.10.77.202-204 pwd=密码 version=18.09.6 handle=install

#### 卸载
    go-install-kubernetes para=docker ips=10.10.77.202-204 pwd=密码 version=18.09.6 handle=uninstall

描述信息清理docker

    # systemctl stop docker 
    # systemctl disable docker
    # rm -rf /var/lib/docker
    # rm -rf /etc/docker/
    # rm -rf /etc/systemd/system/docker.service
    # rm -rf /usr/bin/docker

- 完成后规划（待完成）
将docker更换container

### 6、apiserver scheduler controller-manager
    go-install-kubernetes para=k8s master=10.10.77.202-204  etcd=10.10.77.202-204 pwd=密码 svcip=10.249.0.0/16 clusterip=172.235.0.0/16 handle=install 
将cfssl和cfssljson拷贝到相关机器上并执行创建证书
note: 还未替换相关apiserver地址  config kube-controller-manager.kubeconfig kube-scheduler.kubeconfig
note: 将所有二进制文件一起拷贝过去
### 未来想法（未完成）
    go-install-kubernetes etcd=10.10.77.202-204 master=10.10.77.202-204 node=10.10.77.205-210 pwd=密码 ntpserver=ntpserver proxymode=ipvs
当proxymode不为ipvs为不启用ipvs