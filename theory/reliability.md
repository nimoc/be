---
permalink: /theory/reliability/
---
    
# 可靠性

## 硬件故障

无论是应用代码还是数据库都要存放在硬件中,即使使用云服务最终也还是在硬件中.

而硬件可能出现错误,例如意外关机,磁盘损坏.虽然这种错误出现的几率非常小.可一旦出现故障轻则导致后端服务无法使用,重则丢失数据.

可以使用使用副本/主备/分片等的方式来部署应用和数据库.

例如:

使用 [k8s](../ops/k8s.md) 将相同的应用代码部署在不同的服务器中,这样即使其中一台服务器出现故障,还有其他机器可以访问.

mysql 使用双机主备同步的方式部署,确保一台机器出现硬件故障可以自动切换连接另外一台机器,并且因为有同步机制的存在不用担心数据会丢失.

在极端情况下可能会出现整个机房出现问题,但这种情况出现的概率极小.

如果要避免这种情况需要实现异地多活等方案,而实际上大部分公司没有必要做到这么高规格的可靠性.

**凡事都是需要付出成本的,不能一味的追求完美而不考虑实际情况**

我建议至少保证数据库是有主从或副本机制,公司服务成本的预算再少后端的底线是出现硬件故障时允许出现服务短时间中断但**绝对不能丢数据**.

## 软件错误

软件错误则可能是自己写的业务bug,也可能是第三方库有bug.或某个底层服务异常(短信/邮件/支付接口).

对于这类错误我们应当通过监控和测试来避免这类错误.

## 人为失误

如果说服务器挂了是天灾,那么运维操作服务器时出错就是人祸了.

像数据库链接配置错误,版本发布时候选错版本,意外删除数据库.都属于疏忽大意导致的错误.

我们应当设计简单严谨的运维操作流程和严谨的权限控制来避免这类错误,并且在任何运维操作完成后都观测一段时间服务的运行情况.

> 在清醒时用认真的态度去执行严谨且简单的流程

**制定一套傻瓜式操作来避免来人为失误.必要时还可以在让同事看着你执行高危操作**

## 检验可靠性

在**硬件层面**可以在测试环境中:

1. 尝试重启数据系统/切换主从/关闭某个分片来观测服务的运行情况.然后观察服务是否有重连机制.
2. 尝试重启代码所在服务器,在重启后观察服务是否自动启动.

在**软件层面**需要在编码则需要通过优雅降级,防御式编程等方法确保服务足够健壮.

在**人为失误层面**则需要反复思考如何通过工作流程和规范来避免人为失误

