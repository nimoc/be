# 极速包成功 Sentry 本地部署

**为什么极速？**

本地部署有2个难题：

- 大陆连接 docker github debain 即使加速各种代理和源下载速度还是慢
- 要求机器 4核/14GB，太贵了比一些项目所需服务器还高

解决方法：

- 使用各大IDC的香港轻应用服务器跳过网络问题，在试用跨地区镜像恢复到大陆服务器
- 启动 errors-only 模式只需要 2核/8GB （轻应用服务费用大概1000元/一年）


## 1. 下载与环境配置
```
wget https://github.com/getsentry/self-hosted/archive/refs/tags/25.10.0.tar.gz
tar xf 25.10.0.tar.gz
cd self-hosted-25.10.0/
```

在 `.env` 文件中设置（可选）：
```bash
COMPOSE_PROFILES=errors-only
```
此配置仅需 2核CPU/8GB内存，适合试用环境。

## 2. 服务器选择与镜像部署
**推荐方案**：
- 首选香港轻应用服务器（网络通畅）
- 选择预装 Docker CE 的系统或 CentOS
- **关键步骤**：在香港服务器部署完成后，通过镜像功能创建镜像，跨地区复制到大陆轻量应用服务器使用

## 3. 安装
```bash
./install.sh
```

## 4. 域名配置
```bash
vim sentry/config.example.yml
```
修改：
```yaml
system.url-prefix: http://你的域名.com
# 如果没有域名就配置为 htt://你的ip:9000
#system.url-prefix: http://你的域名.com
```
```
vim sentry/sentry.conf.py
```
修改
```
CSRF_TRUSTED_ORIGINS = ["https://example.com", "http://127.0.0.1:9000"]
```

重启

```
docker-compose down && docker-compose up -d
```

## 5. 启动服务
```bash
docker compose up -d
```
访问方式：
- IP:9000
- 域名+CDN解析到IP:9000

## 6. 创建管理员
```bash
docker exec -it sentry-self-hosted-web-1 bash
sentry createuser
```
按提示输入邮箱和密码即可完成部署。
