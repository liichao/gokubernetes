# gokubernetes
用golang编写一个 一键部署k8s集群，并用来学习golang

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
    go-install-kubernetes para=k8s master=10.10.77.202-204  etcd=10.10.77.202-204 pwd=密码 svcIP=10.249.0.0/16 clusterIP=172.235.0.0/16 nodeCidrLen=24 nodeportrange=2000-60000 handle=install 
    go-install-kubernetes para=k8s ips=10.10.77.202-204  etcd=10.10.77.202-204 pwd=密码 handle=install 
    临时使用ips 后面更改为master

将cfssl和cfssljson拷贝到相关机器上并执行创建证书
note: 还未替换相关apiserver地址  config kube-controller-manager.kubeconfig kube-scheduler.kubeconfig
note: 将所有二进制文件一起拷贝过去

svcip service_ip 不能与主机网络重合
clusterip  cluster_ip  容器ip，不能与主机网络重合
单个node节点允许分布多少个容器ip网段NODE_CIDR_LEN =24

### 6、安装kubelet kube-proxy
    go-install-kubernetes para=k8snode master=10.10.77.202-204 node=10.10.77.206 pwd=herenit123 svcIP=10.249.0.0/16 proxyMode=ipvs pauseImage=easzlab/pause-amd64:3.2  clusterIP=172.235.0.0/16 maxPods=88 handle=install
    # 设置 dns svc ip
    svnIP 获取第二个地址用以dns
    # node节点最大pod 数
    maxPods=88 
    // ipvs 负载均衡master后期在做
    关于masterip这边临时取了第一个masterip为apiserverip

### 7、安装网络flannel 暂时只支持这个模式
  ./go-install-kubernetes para=network node=10.10.77.208 pwd=herenit123 flannelBackend=vxlan flanneldImage=dzero.com/base/flannel:v0.13.0-amd64 clusterIP=172.235.0.0/16 ips=1.1.1.1 handle=install
默认启用ipvs
flannelBackend vxlan host-gw

----待重新测试一波 总感觉有个bug 还未添加删除

### 8、harbor地址账户密码

### 未来想法（未完成）
harbor地址与账户密码 在创建完成集群后自动创建证书
后期规划服务网格 一键安装
### 注意事项

如果重新生成证书，请从etcd开始重新生成。以免导致k8sapi服务因证书原因导致无法链接到etcd
