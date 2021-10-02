# 新手友好的 Kubernetes 教程

本页面是讲稿

## 初学难点 <a id="difficulty"></a>

1. 安装 

2. yaml创建资源

3. 负载均衡和网络

### 安装<a id="install"></a>

**minikube** k8s本机环境安装时候可能会遇到一些问题,WIndows/Mac 平台或多或少会遇到一些问题.并且学习k8s必须理解主机节点和网络相关知识,导致最终还是需要有一个真实的k8s环境.

**k8s集群环境**的安装也有很多限制要求,也可能折腾了一整天才安装完成.

所以建议使用云平台**一键安装按时付费**的k8s集群环境

### yaml创建资源<a id="yaml-create"></a> 

与 yaml 为主线来学习 k8s 对于新手入门不友好,记不住,写错配置都会产生挫败感.

通过 kubesphere 直观的感受k8s,混个面熟之后再去了解 yaml 配置文件.

### 负载均衡和网络 <a id="lb-net"></a> 

k8s 在网络层面需要云服务商配合使用,很多教程在这方面一带而过,

但是在生产环境使用k8s必须了解这些知识.

*介绍 ingress 应用路由的官方文档*

本教程会基于公有云平台使用负载均衡和应用路由,并使用NAT网关来控制节点统一IP

## 腾讯云安装 TKE<a id="tke-install"></a> 

*登录腾讯云安装 TKE 选择2台最低配置的节点,并提现工作环境最少3台节点每台2核4G.*

### 腾讯云删除 TKE 集群<a id="tke-remove"></a> 

当你不在使用集群时,记得删除集群以避免扣费

## 安装 KubeSphere <a id="ks-install"></a> 

https://kubesphere.io

```shell
# 登录节点(替换ip为你的节点ip)
ssh root@20.205.243.166

# 如果 hub.fastgit.org 不能访问则换成 github.com
# 安装KubeSphere
kubectl apply -f  https://hub.fastgit.org/kubesphere/ks-installer/releases/download/v3.1.1/kubesphere-installer.yaml

# 下载集群配置
wget https://hub.fastgit.org/kubesphere/ks-installer/releases/download/v3.1.1/cluster-configuration.yaml

# 修改集群配置文件，PVC 修改为 10G 的倍数
vim cluster-configuration.yaml

# 应用配置
kubectl apply -f cluster-configuration.yaml

# 查看安装情况
kubectl logs -n kubesphere-system $(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') -f

# 访问管理页面(替换ip为你的节点ip)
http://20.205.243.166:30880
```
```shell
vim cluster-configuration.yaml
//默认值
  common:
    mysqlVolumeSize: 20Gi # MySQL PVC size.
    minioVolumeSize: 20Gi # Minio PVC size.
    etcdVolumeSize: 20Gi  # etcd PVC size.
    openldapVolumeSize: 2Gi   # openldap PVC size.
    redisVolumSize: 2Gi # Redis PVC size.

//修改后的值，PVC 为 10G 的倍数（1倍n倍都可以），其他可拔插组件如果开启也需要调整
  common:
    mysqlVolumeSize: 20Gi # MySQL PVC size.
    minioVolumeSize: 20Gi # Minio PVC size.
    etcdVolumeSize: 20Gi  # etcd PVC size.
    openldapVolumeSize: 10Gi   # openldap PVC size.
    redisVolumSize: 10Gi # Redis PVC size.
```

## 使用 Coding 发布镜像<a id="coding-docker"></a> 

1. *演示Coding* 如何创建代码仓库和发布镜像
2. 将 Docker 秘钥在 KubeSphere 上添加到 k8s中