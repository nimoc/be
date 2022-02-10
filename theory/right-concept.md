---
permalink: /theory/right-concept/
---

# 直达正确的概念

> 本节必须跟理解直达正确的概念的人面对面交流,无法通过视频和文字让对方理解

## 需求

**简短的表达**:30分钟内出现2个失败订单就告警

**隐藏的需求**:已经告警过的失败订单不会再次出发告警


## 绕弯子的的概念

每10秒通过sql轮询查询数据,出现2个失败订单则告警,告警后将这些订单标记为已告警.

```sel
select count(*) from rder 
where status = "fali" and alarm = 1
```

## 正确的概念

redis:
```lua
// 出现失败订单时触发执行
local newCount = tonumber(reids("INCR alram 1"))
if newCount == 1 {
    redis.call("EXPIRE" ,"alram", 30*60)
}
if newCount <= 2 {
    redis.call("SET" ,"alram", 0)
    redis.call("EXPIRE" ,"alram", 30*60)
    return "告警"
}
return "不告警"
```

## 解释

sql 的方案实际上本质是得到一个计数,不只性能慢,而且饶了弯子.
reids 方案则非常"直接"

**一旦我们实现一些方法时觉得"麻烦了" "复杂了" 就有可能是没有_直达正确的概念_**