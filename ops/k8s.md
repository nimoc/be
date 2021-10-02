# 一天学会 Kubernetes


## 在腾讯云 TKE 安装 KubeSphere <a id="tke-ks-install"></a> 

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