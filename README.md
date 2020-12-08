# gokubernetes
用golang编写一个 一键部署k8s集群

### 1、安装时钟同步服务器
    go-install-kubernetes para=chrony ips=10.10.77.202-204 pwd=密码 ntpserver=ntpserver

### 2、配置系统参数
    go-install-kubernetes para=system ips=10.10.77.202-204 pwd=密码 proxymode=ipvs

### 3、创建相关证书
    go-install-kubernetes para=createcert ips=10.10.77.202-204 pwd=密码

### 4、创建etcd集群
    go-install-kubernetes para=etcd ips=10.10.77.202-204 pwd=密码

### 未来想法
    go-install-kubernetes etcd=10.10.77.202-204 master=10.10.77.202-204 node=10.10.77.205-210  pwd=密码