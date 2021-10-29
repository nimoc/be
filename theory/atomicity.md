---
permalink: /theory/atomicity/
---

# 原子性

原子性:一个操作或者多个操作 要么全部执行并且执行的过程不会被任何因素打断，要么就都不执行。

> 让多个操作满足原子性是为了防止**事做了一半**导致数据不一致

## SQL原子性 <a id="vbj4V"></a>

**接口**

```http
POST "/register"

Request: {
    "name":"nimoc",
    "password": "******",
}

Response: {

}
```

**数据结构**

```sql
-- user: id,name

CREATE TABLE `user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(10) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- user_password: user_id, password

CREATE TABLE `user_password` (
  `user_id` int(11) unsigned NOT NULL,
  `password` char(128) NOT NULL DEFAULT '',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
```

**需求**

用户输入用户名和密码进行注册，注册成功将信息保存在 `user` 和 `user_password` 表

**实现**

不满足原子性的伪代码如下:

```lua
function register(request) {
  result = execSQL(
                "INSERT INTO user (name) VALUES (?)",
          request.name,
    )
    userID = result.lastInsertID
  execSQL(
                "INSERT INTO user_password (user_id,password) VALUES (?, ?)",
          userID,
                    request.password,
    )
}
```

上面的代码**不满足原子性**的原因是: 像 `user_password` 表插入数据时可能会因为网络原因/语法错误/mysql宕机等各种原因导致插入失败.此时 `user` 表中新增了一个新用户,但是在 `user_password` 表中没有这个新用户的密码信息.这就**导致了数据不一致**.![](https://cdn.nlark.com/yuque/__puml/fef7ae8871fb33bf619952ac16a4d799.svg)

为了保证数据一致需要让2个 `INSERT` 操作是原子性操作,即要么2个INSERT都不执行或执行失败,要么 INSERT都执行成功.

```lua
function register(request) {
  Begin() // 开启事务
  insertUserResult = execSQL(
                "INSERT INTO user (name) VALUES (?)",
          request.name,
    )
    userID = data.lastInsertID
  insertPasswordResult = execSQL(
                "INSERT INTO user_password (user_id,password) VALUES (?, ?)",
          userID,
                    request.password,
    )
  Commit() // 提交事务
}
```

[搜索:sql事务原子性](https://cn.bing.com/search?q=sql+事务原子性)

## Redis 原子性 <a id="redis"></a>

后端新手常犯的错误是依次使用 redis `get` `set` 命令来实现 uv 的计数。

```lua
function uv(userID) {
    visited = redis("GET", userID) != nil
    if (!visited) {
        // 标记用户访问过
        redis("SET", userID, "1")
        // 递增 uv
        redis("INCR", "uv", 1)
    }
}
function queryUV() {
    return redis("GET", "uv")
}
```

这段代码有两个问题：

1. `SET` 之后 不一定能执行 `INCR`
2. 在高并发或恶意攻击的情况下: A请求执行了 GET 之后，B请求也执行了 GET,他们获取到的结果都是 nil，都执行了 `SET` 和 `INCR` 操作。导致了数据不一致。多产生了一个UV.

可通过 redis lua 脚本让三个操作变成原子性操作。

```lua
function uv(userID) {
    redisEval(`
        local visted = redis.call("GET", KEYS[1]) != nil
        if visted then
            redis.call("SET", KEYS[1])
            redis.call("INCR", KEYS[2])
            return 1
        end
        return 0
    `, userID, "uv")
}
function queryUV() {
    return redis("GET", "uv")
}
```

注意 redis lua 脚本的原子性跟 sql 事务原子性不一样，redis lua 脚本内如果命令执行错误，是不会自动回滚的。你需要确保命令语法不要出现错误，这样就能保证命令一定会执行。

这里使用lua脚本实现UV的统计只是为了说明原子性。  
日常工作中 uv 这种场景用 HyperLogLog 或 Sets 更好。

**脚本保证了3个命令一起执行，消除执行间隙。避免了数据竞争，达到了并发安全。**

注意！这里提到了**执行间隙**这个词,即原子性不只是要保证多个操作都执行,在一些场景下还需要保障多个执行执行没有间隙. redis lua 的原子性就是执行之间没有间隙的,这是由 redis 的实现决定的. （执行间隙会导致**数据竞争**，**数据竞争**会导致**数据不一致**）

## 不是每个场景都需要达到原子性 <a id="6ac6d005"></a>

考虑如下场景：

```lua
// 检查验证码
function checkCaptcha(captcha, sessionID) {
    key = "captcha:" + sessionID
    data = redis("GET", key)
    redis("DEL", key) // 读取 captcha 后立即删除，防止恶意穷举
    if (data == captcha) {
        return true
    }
    return false
}
```

GET 和 DEL 不是原子性操作，但是不会造成数据不一致。因为 如果 GET 执行了但是 DEL 没有执行，不会对数据造成任何改动。

分析不满足原子性时候要**想清楚如果不满足原子性会造成具体的什么BUG**。尝试明确的表述出会造成的 BUG 能减少一些非必要的原子性操作。

## 不同系统之间的原子性 <a id="distributed"></a>

考虑如下场景:

```lua
// 发红包
function sendRedpack(accountID, openid, amount) {
    sql("begin")
    // CAS乐观锁扣除余额
    affected = sql("UPDATE account_finance SET balance = balance - $amount WHERE blance >= $amount AND account_id = $accountID")
    if (affected == 0) {
        sql("rollback")
        return
    }
    sql("INSERT INTO red_pack_record (account_id, openid, amount) VALUES($accountID, $openid, $amount)")
    ok = httpRequest("https://api.mch.weixin.qq.com/mmpaymkttransfers/sendredpack", {...})
    if (ok == false) {
        sql("rollback")
        return
    }
    sql("commit")
}
```

上面代码存在如下问题：

长事务，httpRequest的时间是不可控的，可能会导致事务长时间不结束（事务）。长事务会导致系统能支持的并发量下降。UPDATE 后 account\_finance 中 accountID 这一条数据会被锁定。

虽然 UPDATE 和 INSERT是原子性，但是 httpRequest 与 SQL 操作不是原子性， sql commit 有可能因为网络原因失败。这就导致红包发出去了，但是钱没扣。

这就出现了**数据不一致**的问题。

通过本地任务/消息表可以解决不同系统之间的原子性。

> 本地任务/消息表在后续分布式事务章节再详细介绍。
>
> 边路缓存与数据不一致也是个经典的不同系统之间的数据不一致问题

## 总结 <a id="YQJ6l"></a>

1. 多个操作如果没有一起执行或者一起不执行则不满足原子性，会导致数据不一致 。\(sql 事务保障原子性\)
2. 多个操作有执行间隙会导致数据竞争，从而导致数据不一致。（redis 脚本保障原子性）
3. 不是每个场景都需要保障原子性，要分析多个操作不满足原子性后是否会导致bug。
4. 不同系统之间想要满足原子性需要使用本地\(任务/消息\)记录配合\(重试/补偿\)机制来满足原子性。

## 思考 <a id="think"></a>

1. 你的日常的工作场景中又哪些多个操作是要满足原子性的，哪些多个操作是可以不满足原子性的？

> 欢迎留言评论： [https://github.com/nimoc/be/discussions/1](https://github.com/nimoc/be/discussions/1)

