---
permalink: /theory/stat/
---

# 统计

> 统计指的是对项目产生的数据进行统计分析

我们以一个广告统计例子,由浅入深的介绍统计

需求: 用户访问网站或应用会显示广告,显示广告的行为是曝光 `exposure`,用户点击广告的行为是访问 `visit`.
**一天**内**一个用户**相对于**一个广告**发生的**第一次曝光**属于**独立曝光** `is_ue`.
**一天**内**一个用户**相对于**一个广告**发生的**第一次访问**属于**独立访问** `is_uv`.

要区分独立曝光和独立访问的原因是:业务层面以独立曝光和独立访问进行结算

## 原始记录

> 使用 market mkt 代替 ad 表示广告,因为某些浏览器插件会拦截包含 ad 的HTTP请求和链接

创建记录表:

```sql
CREATE TABLE `mkt_record` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `mkt_id` int(11) unsigned NOT NULL,
    `user_id` int(11) unsigned NOT NULL,
    `type` tinyint(4) unsigned NOT NULL,
    -- 注解:C
    `is_uv` tinyint(4) unsigned NOT NULL COMMENT '独立访问',
    `is_ue` tinyint(4) unsigned NOT NULL COMMENT '独立曝光',
    `date` date NOT NULL,
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    -- 注解:E
    KEY `user_id` (`user_id`),
    -- 注解:F
    KEY `date__mkt_id__type` (`date`,`mkt_id`,`type`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;
```

```js
// 创建广告记录
function createMKTRecord(userID, mktID, type) {
    // userID = 100
    // mktID  = 200
    // type = 曝光 exposure 
    // type = 访问 visit
    
    /* 此处省略入参格式与入参数据真实性校验 */
    
    // 用户每天对一个广告只会产生一次uv 独立访问
    isUV = false
    // 用户每天对一个广告只会产生一次ue 独立曝光
    isUE = false
    date = NewDate() // 2022-01-01
    
    // 注解:A
    // 注解:B
    // 通过 redis 获知本次操作是不是uv
    hsetIsUVReply = reids(`HSETNX mkt:is_uv:${date} ${userID}-${mktID} 1`)
    // HSETNX mkt:is_uv:2022-01-01 100-200 1
    // key: mkt:is_uv:2022-01-01
    // field: 100-200
    // value: 1
    if (hsetIsUVReply == 1) {
        isUV = true
    }
    
    // 通过 redis 获知本次操作是不是ue
    hsetIsUEReply = reids(`HSETNX mkt:is_ue:${date} ${userID}-${mktID} 1`)
    if (hsetIsUEReply == 1) {
        isUE = true
    }
    // 注解:D
    sql(`
    INSERT INTO mkt_record (mkt_id, user_id, type, is_uv, is_ue, date)
    VALUES
	(?, ?, ?, ?, ?, ?,);
    `, mktID, userID, type, isUV, isUE, date)
}
```


### 注解:A

使用 hsetnx 而不是 setnx 的原因是:
1. key的前缀 `mkt:is_ue:${date}` 在一天内不会改变,使用 hashes 比 stirngs 存储更节省空间
2. hsetnx 无需设置key过期时间,使用定时脚本在每日凌晨1点执行 `del mkt:is_ue:${三天前日期}` 即可 (删除三天前的 key 而不是昨天是为了万一出现bug方便排查) 

### 注解:B

使用 redis 判断 uv或者ue 之后进行sql插入操作这个行为不满足原子性,
当 redis 设置成功之后有可能进程中断或者sql网络连接异常.
这种情况下某个用户可能今天无法产生uv或ue,不过不是每个场景都需要满足原子性的以达到数据一致.
大部分业务允许出现极小概率的数据不一致.

> 如果要满足判断uv和插入数据的原子性,可以通过开启sql事务,然后向一张复合主键为 `date,user_id,mkt_id,is_uv` 的表通过 `INSERT IGNORE INTO` 的方式插入数据,
> 如果sql执行返回的受影响行数是1.则is_uv 为 true,否则为 false.
> 然后再向 `mkt_record` 表插入数据.
> 不过这种方式没有redis hsetnx 性能高,性能和数据一致性你可以进行实际业务情况做取舍.
> 多嘴提一句一般情况下此处的redis sql 不满足原子性是极小概率才会出现的.

### 注解:C

`mkt_record` 表有 `is_uv` 和 `is_ue` 字段.如果没有这两个字段在一天结束时也能分析出有多少 uv 和 ue.但是这需要使用 sql 的
查询分组 `GROUP BY user_id` 或者 字段去重`count(distinct user_id)`, 但是这样做性能没有使用 `WHERE is_uv = 1` 查询的方式高. 

所以在创建数据的时候即时计算出冗余字段 `is_uv` `is_ue` 能提高统计性能,最重要的是业务上其他的逻辑可能也需要即时计算出 `is_uv` `is_ue`

### 注解:D

`mkt_record` 表在已经有 `create_time` 字段保存时间的情况下,特意增加了 `date` 字段.目的是为了统计的查询性能.
这是一种增加合理冗余字段并作为索引的数据设计和性能优化的技巧.

### 注解:E

`mkt_record` 表有 `user_id` 索引的原因是因为基于实践经验,在业务上经常会使用  `WHERE user_id = ?` 查找 `record` 表的数据,这样做能提高查询性能.

### 注解:F

`mkt_record` 表 有`date,mkt_id,kind`的复合索引的原因是统计分析sql会使用到 `WHERE date = ? AND mkt_id = ? AND type = ?`,建立索引能提高统计性能.

需要注意的是,索引越多插入数据越慢.向无索引和有索引的表插入时会发现速度不一样.

### 代码实现

1. [Go 插入数据](./stat_code/go/insert_data/main.go)

至此我们完成的广告原始记录的伪代码

## 统计分析每日数据


## 即时计数提高性能


## 统计精度

TODO... 时间(日/分钟)精度,范围(平台)精度,


## 更多信息

TODO... 地理位置