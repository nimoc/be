---
permalink: /theory/performance-indicators/
---
# 性能指标 

## RT

**R**esponse **t**ime

响应时间

从服务器接收到请求到响应请求的时间称之为响应时间,RT反应了系统的快慢.

## QPS和TPS

**Q**ueries **P**er **S**econd
 
每秒查询数

**T**ransactions **P**er **S**econd

每秒事务数

例如 `/user` 接口返回余额和头像.

使用如下2个SQL查询数据

```sql
SELECT `balance` FROM finance WHERE `user_id` = ?
```

```sql
SELECT `avatar` FROM user_info WHERE `user_id` = ?
```

每秒能响应多少个 `/user` 请求称之为 TPS
每秒能完成多少次余额 `balance` 称之为 QPS

一个"事务"`(T)`中可以只包含一个"请求"`(Q)`,也可以有多个.

可以使用压力测试来得到TPS或QPS的结果,

