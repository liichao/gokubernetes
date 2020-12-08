# gokubernetes
用golang编写一个 一键部署k8s集群

### 1、安装时钟同步服务器
    go-install-kubernetes para=chrony ips=10.10.77.202-204 pwd=密码 ntpserver=ntpserver

### 2、配置系统参数
    go-install-kubernetes para=system ips=10.10.77.202-204 pwd=密码 proxymode=ipvs

### 3、创建相关证书
    go-install-kubernetes para=createcert ips=10.10.77.202-204 pwd=密码

### 4、创建etcd集群
#### 安装
    go-install-kubernetes para=etcd ips=10.10.77.202-204 pwd=密码

#### 清理（未完成）
    go-install-kubernetes para=etcd ips=10.10.77.202-204 pwd=密码 handle=remove

描述信息清理etcd集群
    # systemctl stop etcd 
    # systemctl disable etcd
    # rm -rf /var/lib/etcd
    # rm -rf /etc/etcd/
    # rm -rf /etc/systemd/system/etcd.service

### 5、安装docker（未完成）
    go-install-kubernetes para=docker ips=10.10.77.202-204 pwd=密码 harbor=url
- 完成后规划
将docker更换container

### 未来想法（未完成）
    go-install-kubernetes etcd=10.10.77.202-204 master=10.10.77.202-204 node=10.10.77.205-210 pwd=密码 ntpserver=ntpserver proxymode=ipvs
当proxymode不为ipvs为不启用ipvs