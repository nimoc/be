---
permalink: /theory/cache/
---

# 缓存

## 旁路缓存

旁路缓存可以提高服务性能.

### 需求

例如数据在 mysql 中,读取数据需要执行SQL

```js
function getNews_1(id) {
    data = sql(`SELECT title, content FROM news WHERE id = 1`)
}
```

### 优先读缓存

当请求量大时mysql的cpu和内存可能会超过100%,导致服务不可用.此时可以利用 redis(高速缓存) 提高性能.

思路:
1. 读数据时优先读cache
2. 如果缓存读不到则读db
4. 如果读到数据则将数据写到cache

```js
function getNews_2(id) {
    // A 读取缓存
    cache = redis(`GET news:${id}`)
    if (cache != nil) {
        // B 如果缓存存在则返回缓存
        return cache
    }
    // C 缓存不存在则读取db
    data, hasData = sql(`SELECT title, content FROM news WHERE id = 1`)

    // D 如果db中无数据则响应无数据
    if (hasData == false) {
        return "暂无数据"
    }
    cacheValue = JSON.stringify(data)
    // C 读取到sql数据后向缓存写入数据.并设置过期时间10分钟
    // SET key value [EX seconds]
    // SET news:1 {"title":"xxx","content": "xxx"} EX 600
    redis(`SET news:${id} ${cacheValue} EX 600`)
}
```

## 缓存击穿

> 数据还没从db同步到缓存时,瞬间出现大量请求会导致大量读db

解决方法: 使用互斥锁(SET NX)控制短时间内只有一个请求能查询数据库

### 缓存穿透

> 当用户请求数据库中没有的数据时会发生缓存穿透

当执行 `getNews_2(100)` 时如果 `id=100` 这个数据不存在于db中所以的请求都会执行到 `// D 如果db中无数据则响应无数据`

执行 `D` 之前必然要执行 `C 读取db` 这就导致高并发时缓存完全失效.

这种情况就叫:缓存穿透

解决方法: 将 "查不到数据" 这个结果缓存起来解决缓存穿透的问题


```js
function getNews_3(id) {
    // A 读取缓存
    cache = redis(`GET news:${id}`)
    if (cache != nil) {
        // B 如果缓存存在则返回缓存
        return cache
    }
    // F 在缓存中查询数据是否不存在
    notFoundReply = redis(`GET news:not_found:${id}`)
    if (notFoundReply != nil) {
        return "暂无数据"
    }
    // C 缓存不存在则读取db
    data, hasData = sql(`SELECT title, content FROM news WHERE id = 1`)

    // D 如果db中无数据则响应无数据
    if (hasData == false) {
        // E 查不到数据时候写入not_found缓存,过期时间10秒
        // SET key value [EX seconds]
        // SET news:not_found:1 1 EX 10
        redis(`SET news:not_found:${id} 1 EX 10`)
        return "暂无数据"
    }
    cacheValue = JSON.stringify(data)
    // C 读取到sql数据后向缓存写入数据.并设置过期时间10分钟
    // SET key value [EX seconds]
    // SET news:1 {"title":"xxx","content": "xxx"} EX 600
    redis(`SET news:${id} ${cacheValue} EX 600`)
}
```

`getNews_3()` 增加了 `// E 查不到数据时候写入not_found缓存` 和 `// F 在缓存中查询数据是否不存在`
