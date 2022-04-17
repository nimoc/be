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


新手可能解决统计需求时会使用 SQL group by 直接查询数据.

例如查询某广告最近一周的UV总和:

```sql
select count(*) from mkt_record
where date between "2022-01-01" and "2022-01-07"
and mkt_id = 100
and type   = 1
and is_uv  = 1
```

在每日产生100万数据的表中,虽然这个查询用上了索引 `date__mkt_id__type`,但是查询速度还是很慢.一亿的数据需要耗时3.5s.

例如查询某广告最近一周UV的每日数据:

```sql
select date, count(date) as "uv" from mkt_record
where date between "2022-01-01" and "2022-01-07"
and mkt_id = 100
and type   = 1
and is_uv  = 1
group by date
```

在每日产生100万数据的表中,查询速度1.8秒.



**查询结果**

```
date	uv
2022-01-01	174
2022-01-02	217
2022-01-03	214
2022-01-04	255
2022-01-05	264
2022-01-06	277
2022-01-07	309
```

> 使用SQL在数据量大的详细数据记录表中进行统计分析会很慢,可以采用每日凌晨生成数据日表的方式将查询结果提前准备好.在需要查询的时候直接查询数据日表**
> 这种拆分统计保存日表再查询日表得到统计结果的方式是一种很常见的统计实现技巧.

可以在每日凌晨1点生成昨日广告数据

**曝光**

```sql
select count(*) as "曝光:exposure" from mkt_record
where 
    date = "2022-01-01"
    and mkt_id = 1
    and type = 1
```

***独立曝光*

```sql
select count(*) as "独立曝光:unique_exposure" from mkt_record
where 
    date = "2022-01-01"
    and mkt_id = 10
    and type   = 1
    and is_uv  = 1
```

**访问**

```sql
select count(*) as "访问:visit" from mkt_record
where 
    date = "2022-01-01"
    and mkt_id = 10
    and type = 2
```

**独立访问**

```sql
select count(*) as "独立访问:unique_visit" from mkt_record
where 
    date = "2022-01-01"
    and mkt_id = 10
    and type   = 2
    and is_ue  = 1
```

在每日产生一百万条数据,总数据量一亿,平均每个广告一天产生2000条数据的情况下.
查询利用了索引 `date__mkt_id__type`, 查询速度很快.

然后就可以将查询出的数据存储到广告日统计表中

```sql
CREATE TABLE `mkt_stat_of_date` (
  `date` date NOT NULL,
  `mkt_id` int(11) unsigned NOT NULL,
  `exposure` int(11) unsigned NOT NULL,
  `ue` int(11) unsigned NOT NULL,
  `visit` int(10) unsigned NOT NULL,
  `uv` int(10) unsigned NOT NULL,
  PRIMARY KEY (`date`,`mkt_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
INSERT INTO `mkt_stat_of_date` (`date`, `mkt_id`, `exposure`, `ue`, `visit`, `uv`)
VALUES
	('2022-01-01', 100, 8, 4, 6, 2);
```

主键使用`date,mkt_id`作为复合主键,这样能保证不会出现2份相同日期相同广告ID的数据/

> 插入主键重复数据的数据会报错 Duplicate entry '2022-01-01-100' for key 'PRIMARY'
> 可以使用 `INSERT IGNORE INTO` 插入数据避免报错,并检查SQL执行完成后返回的受影响行数是否为1判断插入是否成功


有了广告日统计表查询最近7天综合和最近7天每日数据


**最近一周某广告uv总和**

```sql
select sum(uv) from mkt_stat_of_date
where date between "2022-01-01" and "2022-01-07"
and mkt_id = 100
```


**最近一周某广告每日uv**

```sql
select date, uv from mkt_stat_of_date
where date between "2022-01-01" and "2022-01-07"
and mkt_id = 100
```

在业务规模流量小,不需要实时展示uv的情况下.
每日凌晨1点查出日统计数据并存储到广告日统计表的方式易于实现且性能高.

## 即时计数统计

一旦每日广告记录数据流达到十万百万时.即使使用了缓存,查询速度也不一定会很快.

提高统计性能常用的方式就是分摊压力,将展示数据统计结果那一刻的查询压力提前消化.

在产生广告曝光访问行为时进行计数递增,这样统计压力就均摊到了每一次的用户请求中.
不用担心每日数据量太多导致统计日表无法生成.

> 你可能在第一次实现统计的时候就使用了计数统计,这很不错.但也请一定将用户行为保存到 `mkt_record` 中吗.
> 因为随着可能会有新的统计需求时需要对原始数据进行分析,也可能是需要将详细数据展示给用户.

mysql 和 redis 都能实现计数递增的功能,方法分别是

**redis**
```lua
HINCRBY uv:counter:${date} ${mktID} 1
```
**mysql**
```sql
update mkt_uv_counter_of_date
set uv = uv + 1
where 
    date       = "2022-01-01"
    and mkt_id = 100
```

接下来我分析一下各自的优劣点:

TOOD...


## 统计精度

TODO... 时间(日/分钟)精度,范围(平台)精度,


## 更多信息

TODO... 地理位置