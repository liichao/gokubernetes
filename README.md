# gokubernetes
用golang编写一个 一键部署k8s集群，并用来学习golang

## 前期准备
- 系统版本
    Centos 7.6 内核版本4.14
- harbor私有仓库账户密码
- 配置文件中所需要的二进制文件与docker镜像文件
## 配置目录
![整个程序目录](https://github.com/liichao/gokubernetes/blob/main/images/FileTree.jpg)
## 配置文件说明
```yaml
master:
  10.10.77.191-193
node:
  10.10.77.191-196
# 写master中随便一个即可，在安装完成后 会自动替换为ivps service地址
apiServer:
  10.10.77.191
# k8s的serviceip地址段
svcIP:
  10.249.0.0/16
# k8s pod内容器ip地址段
clusterIP:
  172.235.0.0/16
password:
  密码
# 是否安装时间同步服务器
chrony:
  enable:
    true
  # 时钟同步服务器
  ntpServer:
    ntp.aliyun.com
##### 以下内容为私有仓库内容 替换dzero.com 以及10.10.76.185 即可
# harbor地址
harborIP:
  private:
    true
  IP:
    10.10.76.185
harborUrl:
  dzero.com
# harbor账户
harborUser:
  admin
# harbor密码
harborPwd:
  Harbor12345
# flanneld 离线docker镜像
flanneldImageOffline:
  flanneld-v0.13.0-amd64.docker
# flannled 镜像地址
flanneldImage:
  dzero.com/base/flannel:v0.13.0-amd64
# pause 离线docker镜像
pauseImageOffline:
  pause-amd64-3.2.tar
#  pasuse 镜像地址
pauseImage:
  dzero.com/base/pause-amd64:3.2
# 修复iptables版本过低
repair:
  iptables:
    false
  iptablesBinName:
    iptables-1.6.2.tar.bz2
# 集群dns的域名，在kubelet-config.yaml的配置也有这个配置
clusterDnsDomain:
  cluster.local.
# 安装其他增项 默认为安装 
other:
  install:
    true
  # coredns安装
  dns:
    install:
      true
    # 是否载入镜像，当为false时候需要载入镜像，并推送到仓库
    loadImage:
      true
    # dnsIP 为svcip的第二个ip
    dnsIP:
      10.249.0.2
    coreDnsImages:
      dzero.com/base/coredns:1.8.0
    coredns_offline:
      coredns_1.8.0.tar
  metricsserver:
    install:
      true
    loadImage:
      true
    metricsServerImages:
      dzero.com/base/metrics-server-amd64:v0.3.6
    metricsserver_offline:
      metrics-server-amd64_v0.3.6.tar
  # dashboard 安装
  # dashboard v2.x.x 不依赖于heapster
  dashboard:
    install:
      true
    loadImage:
      true
    dashboardImages:
      dzero.com/base/dashboard:v2.1.0
    dashboard_offline:
      dashboard_v2.0.4.tar
    metricsscraperImages:
      dzero.com/base/metrics-scraper:v1.0.6
    metricsscraper_offline:
      metrics-scraper_v1.0.6.tar
##### 以下内容均不需要修改，或按需修改
# 是否启用ipvs apiserver高可用。 如果启用 在集群部署完成后，会将 apiServer的ip 替换成 svcIP网段的第一个地址10.249.0.1:443 
# 从而实现apiserver 高可用 
# 这个组件也可以后面单独启用的
ipvsApiServer:
  true
# 相关配置目录
k8sPath:
  /tmp/k8s/
# 是否配置系统
system:
  true
# 系统内核 3或者4
kernel:
  4
# 是否启用ipvs，默认启用
proxyMode:
  ipvs
# 创建证书，如果第二次运行可以改为no 第一次必须false
createCert:
  true
# 是否安装etcd
etcd:
  true
# 当为true安装etcd集群 当为false时 删除etcd集群
etcdInstall:
  true
# 在node上安装docker
docker:
  true
# 当为true安装docker 当为false时 删除docker
dockerInstall:
  true
# docker的离线二进制文件放到k8spath/tools下
dockerInstalltgz:
  docker-18.09.6.tgz
# k8s api 安装
installK8sApi:
  true
removeK8sApi:
  false
# k8s node 安装
installK8sNode:
  true
# k8s node 删除
removeK8sNode:
  false
# k8s node节点的掩码
nodeCidrLen:
  24
# k8s nodePort允许分配的端口范围
nodePortRange:
  2000-60000
# 单个node 允许运行最多pod数
maxPods:
  88
# cni网络
cnitools:
  cni-plugins-linux-amd64-v0.9.0.tgz
# 安装网络插件
network:
  install:
    true
  loadImage:
    true
# 网络模式  vxlan  host-gw 等
flannelBackend:
  vxlan
```

### 待完成

    紧急处理清理功能
    -   添加各种命令判断功能
    - 离线安装ipset ipvsadm
    - 删除node里面有个危险操作，待修改
    - 将docker更换container
    - 测试1.18和1.19和1.20的k8s安装
    - 后期规划服务网格 一键安装


### 注意事项
如果重新生成证书，请从etcd开始重新生成。以免导致k8sapi服务因证书原因导致无法链接到etcd
