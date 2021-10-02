# 锁

> 锁的目的是防止在并发时的数据竞争导致数据不一致

在能保证数据一致的情况下尽量的不要用锁,使用无锁的方式读写数据能增加并发性能.

## 递增 <a id="incr"></a>

新手经常烦的错误就是使用读取+1再写入来实现递增.例如:

**redis**

```lua
// 错误示例
count = redis("GET count")
newCount = count+1
redis("SET count", newCount)
```

在并发的场景下 读取 `GET count` 和 写入`SET count newCount`之间存在执行间隙,执行间隙会导致一瞬间并发时10次递增操作成功不一定能将0递增到10,可能是1~10之间的任意数字.如果不明白为什么可以[留言讨论](https://github.com/nimoc/be/discussions/2).

**sql**

```javascript
// 错误示例
result = sql(`SELECT value FROM count WHERE id = 1`)
newValue = result.row.value +
sql(`UPDATE count SET value = {newValue} WHERE id = 1`)
```

sql 与 redis 存在同样问题.

redis 和 sql 都提供了递增来解决

```lua
reids("INCR count")
```

```lua
sql(`UPDATE count SET value = value + 1 WHERE id = 1`)
```

## 限制递增最大值 <a id="limit"></a>

工作中会遇到计数限制的需求,由某个用户行为触发一个数字的递增,递增到达一个设定的最大值时不再递增

先列举几种错误的操作:

**redis**

```javascript
// 错误示例
max = 10
count = redis("GET count")
if (count >= max) {
    return
}
redis("INCR count")
```

GET INCR 之间有执行间隙,并发时会导致最终count超出10

```javascript
// 错误示例
newCount = redis("INCR count")
if (newCount > max) {
  reids("DECR count")
}
```

INCR 和 DECR 不是原子性操作,最终count可能超出10

**mysql**

```javascript
// 错误示例
max = 10
result = sql(`SELECT value FROM count WHERE id = 1`)
newValue = result.row.value +
if (newValue >= max) {
  return
}
sql(`UPDATE count SET value = {newValue} WHERE id = 1`)
```

同样存在执行间隙的问题

正确的做法是

**redis**

```javascript
// 使用 redis lua 脚本执行,保障命令直接没有执行间隙
redisEval(`
local count = tonumber(redis.call("GET", "count"))
if (count < 10)
then
    return
end
redis.call("INCR", "count")
`)
```

**mysql**

```javascript
result = sql(`UPDATE count SET value = {newValue} WHERE id = 1 AND value < 10`)
// 获取 result.affected 可以知道是否修改了数据,如果没修改则表示 count 已经达到10
```

## 使用redis锁实现复杂的计数限制 <a id="redis-lock"></a>

接下来由简到难的介绍使用锁来实现计数限制

### 用户24小时只能领取1次奖品 <a id="24h-limit"></a>

> 阅读前确保了解 redis SET NX EX [https://redis.io/commands/set](https://redis.io/commands/set)

```javascript
// redis
userID = 1
key = "prize:" + ":"userID
// 60*60*24 = 一天的秒数
reply = redis("SET", key, "1", "NX", "EX", 60*60*24)
if (reply == nil) {
  return
}
// 继续抽奖逻辑
```

### 用户24小时只能领取2次礼品 <a id="24h-limit-2times"></a>

```javascript
// redis
userID = 1
key = "prize:" + ":"userID
ex = 60*60*24
max = 2
replyInt = redisEval(`
local key = KYES[1]
local ex = ARGV[1]
local max = ARGV[2]

local replyGet = redis.call("GET", key])
local count = 0
if (replyGet)
then
    -- 如果 key 存在 count 为读取的值
    count = tonumber(replyGet)
else
 -- 如果 key 不存在则 count 为 0
    count = 0
end
if (count >= max)
then
    return 0
end
if (!replyGet)
then
    redis.call("SET", key, 0, "EX", ex)
end
redis.call("INCR", key)
return 1
`, {
  KEYS: [key],
  ARGV: [ex, max]
})
if (replyInt == 1) {
  // 发放
} else {
  // 不发放
}
```

当限制次数不是1次时就必须使用redis lua 脚本去执行,保证读取\(GET\)和写入\(INCR\)没有执行间隙

## 使用mysql锁实现扣除抽奖机会 <a id="mysql-lock"></a>

