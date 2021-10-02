# 好友关系

## 需求 <a id="spec"></a>

数据结构如下：

```sql
CREATE TABLE `user` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(20) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

参考下面的代码使用你自己熟悉的语言实现

{% hint style="warning" %}
不要着急立即开发，看完**练习**小结再开始开发
{% endhint %}

```lua
// 因为目的是学习和练习，所以不需要写 http 代码
// 也不需要写 session 等代码

// userID 是当前登录用户
// friendUserID 是当前登录用户想要添加的好友用户ID
function add(userID, friendUserID) {
  // return "can_not_add_yourself" // 不能添加自己
  // return "ok" // 添加好友成功
  // return "repeat" // 该用户早已是你的好友，无需添加好友
}
// userID 是当前登录用户,
// 返回当前登录用户的好友ID列表
function list(userID) {
  // return [2,29]
}
// 查看两个用户是不是好友
function is(userID, friendUserID) {
  // return true
  // return false
}
// 解除好友关系
function delete(userID, friendUserID){
  return "ok" // 删除好友成功
}
// 查看共同好友
function mutual(userID, friendUserID){
  return [3,4]  // 1,2和3,4 都是好友
}
```

## 测试 <a id="test"></a>

按顺序调用如下代码

```lua
is(1,2) // false
is(2,1) // false
add(1,2) // ""
add(2,1) // repeat
list(1) // 2
list(2) // 1
is(1,2) // true
is(2,1) // true
add(1,3) // ""
list(1) // 2,3
delete(1,2) //  ""
delete(1,2) //  "not friends"
is(1,2) // false
list(1) // 3
cleartFriendUserData() // 清除关系数据
add(1,2)
add(1,3)
add(1,4)
add(2,3)
add(2,4)
mutual(1,2) // 3,4
// 根据你自己的语言并发执行20 次 add(5,6)
// 然后检查数据库数据是否正常(数据一致)
```

将上面的代码改成你熟悉的编程语言的单元测试，基于单元测试开发

## 练习 <a id="practice"></a>

1. 先只使用mysql实现一遍
2. 然后只用redis实现一遍

你可以先自己思考实现流程然后最多花一个小时实现一遍，便于加深印象。否则会**一学就会，一写就废。**

超过了一个小时/觉得累了/实现完了，就休息一会再来看下面的章节。

{% hint style="warning" %}
先用最快的速度实现
{% endhint %}

## 只用mysql <a id="mysql"></a>

新增一张 `user_friend` 表

```sql
CREATE TABLE `user_friend` (
  `user_id` bigint(11) unsigned NOT NULL,
  `friend_user_id` bigint(11) unsigned NOT NULL,
  PRIMARY KEY (`user_id`,`friend_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
```

{% hint style="info" %}
接下来读写数据时候注意将 user\_id 和 friend\_user\_id 进行排序后操作
{% endhint %}

{% code title="伪代码" %}
```javascript
function sortUserID(aid, bid) {
    if (aid < bid) {
        return {firstUserID: aid, secondUserID: bid}
    }
    if (aid > bid) {
        return {firstUserID: bid, secondUserID: aid}
    }
    return {firstUserID: aid, secondUserID: bid}
}
function add(userID, friendUserID) {
    sortID = sortUserID(userID, friendUserID);
    // 使用 IGNORE INTO 防止重复插入
    sql("INSERT IGNORE INTO `user_friend` (`user_id`,`friend_user_id`) VALUES (?,?)", sortID.fristUserID, sortID.secondUserID)
}
function is(userID, friendUserID) {
    sortID = sortUserID(userID, friendUserID);
    // 用 SELECT 1 取代 SELECT count(*) 来查询单条数据是否存在性能更好
    return sql("SELECT 1 FROM `user_friend` WHERE `user_id` = ? AND `friend_user_id` = ? LIMIT 1", sortID.fristUserID, sortID.secondUserID)
}
function delete(userID, friendUserID) {
    sortID = sortUserID(userID, friendUserID);
    sql("DELETE FROM `user_friend` WHERE `user_id` = ? AND `friend_user_id` = ? LIMIT 1", sortID.fristUserID, sortID.secondUserID)
}
function list(userID) {
    // 利用 UNION 查询2次即可
    return sql(`
        SELECT friend_user_id
        FROM user_friend
        WHERE user_id = ?
        UNION
        SELECT user_id
        FROM user_friend
        WHERE friend_user_id = ?`, userID, userID)
}
function mutual(userID, friendUserID) {
    sortID = sortUserID(userID, friendUserID);
    // 先通过 UNION 查出2个用户的好友列表,然后利用 INNSER JOIN 交集的特性查出数据
    return sql(`SELECT a.user_id FROM
         (
             SELECT user_id
                 FROM user_friend
                 WHERE friend_user_id = ?
             UNION
             SELECT friend_user_id AS user_id
                 FROM user_friend
                 WHERE user_id = ?
         ) AS a
         INNER JOIN
         (
             SELECT user_id
                 FROM user_friend
                 WHERE friend_user_id = ?
             UNION
             SELECT friend_user_id AS user_id
                 FROM user_friend
                 WHERE user_id = ?
         ) AS b
         ON (a.user_id = b.user_id)`,

         sortID.firstUserID,
         sortID.firstUserID,

         sortID.secondUserID,
         sortID.secondUserID,
         )
}
```
{% endcode %}

只使用 sql 实现需要注意的有一下几点

1. 通过 `PRIMARY KEY (user_id,friend_user_id)` 和`INSERT IGNORE INTO` 防止并发add时候导致数据不一致\(多了重复数据\)
2. 实现 `sortUserID` 函数,将2个userid 进行排序后进行读写,如果不这么做,读写性能性能会差一点
3. 通过 `INNER JOIN` 实现交集,共同好友在数学上就是交集计算

各个编程语言实现版本:

1. [go](https://github.com/nimoc/be/blob/master/practice/friends/go/internal/mysql.go)

## 只用 redis <a id="redis"></a>

redis 的 sets 结构实现好友关系非常简单.注意使用 lua 保障多个操作是原子性即可

{% code title="伪代码" %}
```lua
function add(userID, friendUserID) {
    redis.call("SADD, userID, friendUserID)
    redis.call("SADD, friendUserID, userID)
}
function is(userID, friendUserID) {
    redis.call("SISMEMBER, userID, friendUserID)
}
function list(userID) {
    redis.call("SMEMBERS", userID)
}
function delete(userID, friendUserID) {
    redis.call("SREM, userID, friendUserID)
    redis.call("SREM, friendUserID, userID)
}

function mutual(userID, friendUserID) {
    redis.call("SINTER", userID, friendUserID)
}
```
{% endcode %}

注意 add 和 delete 都需要使用 lua 来满足原子性

各个编程语言实现版本:

1. [go](https://github.com/nimoc/be/blob/master/practice/friends/go/internal/redis.go)

## mysql + redis <a id="mysql-redis"></a>

在好友关系的场景下只使用mysql方案和只使用redis方案的优点是对方的缺点.

1. mysql优点: 持久化不会丢数据
2. mysql缺点: 数据量大时mutual 性能慢
3. redis优点: 所有的操作性能都好 \(内容操作\)
4. redis缺点: 数据可能会丢失\(1秒同步落盘导致\)

性能和持久化的差别是由mysql实现和redis实现导致的,这里就不再深入讨论了.

只需要将2中存储结合起来使用就能既做到数据一致性不丢失,又做到高性能.

这里需要先了解理论知识:

[旁路缓存](https://www.bing.com/search?q=%E6%97%81%E8%B7%AF%E7%BC%93%E5%AD%98)

实现时需要注意:

1. 在同步缓存发现数据库也没有数据时使用 redis `set no_friend:{userID} ex 2` 避免缓存穿透
2. 在同步数据时使用 redis `set friend_syncing:{userID} ex 2 nx` 避免缓存击穿\(同步完成要 `del friend_syncing:{userID}`\)
3. 记住"读操作时缓存无数据则同步缓存后返回缓存数据,写操作时清除缓存\(包括缓存击穿/穿透\)"

各个编程语言实现版本:

1. [go](https://github.com/nimoc/be/blob/master/practice/friends/go/internal/union.go)



Github 评论: https://github.com/nimoc/be/discussions/3