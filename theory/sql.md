---
permalink: /theory/sql/
---

# sql

## 数据竞争 <a id="data-race"></a>

在产生高并发时会产生数据竞争,数据竞争会导致数据不一致.

递增

```sql
CREATE TABLE `data_race_incr` (
  `user_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `count` int(11) unsigned NOT NULL,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
INSERT INTO `data_race_incr` (`user_id`, `count`)
VALUES
	(100, 0);
```

```js
var userID = 100
var data = sql("SELECT count FROM data_race_incr WHERE user_id = ? LIMIT 1", userID)
var newCount = data.count + 1
sql("UPDATE data_race_incr SET count = ? WHERE user_id = ? LIMIT 1", newCount, userID)
```


A请求和B请求的时间表:

|时间| A操作    | B操作    |
|---|--------|--------|
|t1| SELECT |        |
|t2|        | SELECT |
|t3|        | UPDATE |
|t4|UPDATE|        |

`SELECT` 和 `UPDATE` 之间有执行间隙,如果A请求执行了`SELECT` 但未执行到 `UPDATE`时 B请求已经执行完了 `SELECT` 和 `UPDATE`.
之后当A请求执行完UPDATE时数据库中 userID 100 的 count 是2,而不是2.

使用原子性的操作消除执行间隙就可以避免数据竞争

```sql
UPDATE data_race_incr SET count = count + 1 WHERE user_id = 100 LIMIT 1
```

----

你通过锻炼和思考学会分析代码是否存在数据竞争,并使用本节的知识来避免数据竞争

## 幂等性 <a id="idempotent"></a>


> 在编程中一个幂等操作的特点是其任意多次执行所产生的影响均与一次执行的影响相同。幂等函数，或幂等方法，是指可以使用相同参数重复执行，并能获得相同结果的函数。这些函数不会影响系统状态，也不用担心重复执行会对系统造成改变。例如，“setTrue()”函数就是一个幂等函数,无论多次执行，其结果都是一样的.更复杂的操作幂等保证是利用唯一交易号(流水号)实现。 

知识点:

1.`PRIMARY KEY` 和 `UNIQUE` 的唯一约束
2.`INSERT IGNORE INTO`

###  判断UV

```sql
CREATE TABLE `idempotent_uv` (
  `date` date NOT NULL,
  `news_id` int(11) unsigned NOT NULL,
  `user_id` int(11) unsigned NOT NULL,
  PRIMARY KEY (`date`,`news_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
INSERT IGNORE INTO `idempotent_uv`
(`user_id`, `news_id`, `date`)
VALUES
(1,100, "2022-01-01")
```

运行 `INSERT IGNORE INTO` 后获取 `rows affected`, 如果是1属于UV,如果是0属于PV.

`PRIMARY KEY` 和 `UNIQUE KEY` 能对数据进行唯一约束. `idempotent_uv` 表的 `PRIMARY KEY` 由多个字段组成

插入数据时 sql 会进行唯一性判断,如果主键或UNIQUE重复则会返回错误: `Duplicate entry '2022-01-01-100-1' for key 'PRIMARY'`

使用 `INSERT IGNORE INTO` 能忽略错误,通过检查SQL执行返回的 `rows affected` 来判断数据插入是否成功.

> 对数据一致性要求没有那么高的场景可以使用 redis 判断uv

###  订单号

微信支付和支付宝支付的支付接口都要求每次请求支付时提交一个商户订单号`out_trade_no`,作用是防止重复提交.


```sql
CREATE TABLE `idempotent_order` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` int(10) unsigned NOT NULL,
  `out_trade_no` char(36) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `merchant_id__out_trade_no` (`merchant_id`,`out_trade_no`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4;
```

**幂等插入**

```sql
INSERT IGNORE INTO `idempotent_order` (`merchant_id`, `out_trade_no`)
VALUES
	(100, 'a');
```

运行 `INSERT IGNORE INTO` 后获取 `rows affected`, 如果是1则订单创建成功,如果是0则响应"商户订单号重复订单可能重复提交".

`PRIMARY KEY` 和 `UNIQUE KEY` 能对数据进行唯一约束. `idempotent_order` 表的 `UNIQUE KEY` 由多个字段组成

使用 `INSERT IGNORE INTO` 能忽略错误,通过检查SQL执行返回的 `rows affected` 来判断数据插入是否成功.

---

**幂等更新**

如果你的业务场景中必须先向 `idempotent_order` 插入数据,再更新 `out_trade_no` 字段.

尝试多次执行

```sql
INSERT INTO `idempotent_order` (`merchant_id`, `out_trade_no`)
VALUES
	(100, '');
```

只会产生一条数据


可以修改 `out_trade_no` 字段允许为 NULL,然后再 UPDATE

```sql
-- 准备表
-- out_trade_no 允许为 NULL
CREATE TABLE `idempotent_order` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `merchant_id` int(10) unsigned NOT NULL,
  `out_trade_no` char(36) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `merchant_id` (`merchant_id`,`out_trade_no`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4;
```

```sql
-- 准备数据
INSERT INTO `idempotent_order` (`id`, `merchant_id`, `out_trade_no`)
VALUES
	(1, 100, NULL),
	(2, 100, NULL);
```

```sql
-- 幂等更新
UPDATE `idempotent_order`
SET out_trade_no = "a"
WHERE id = 1
```

## 乐观锁 CAS  <a id="cas"></a>

> CAS 是 compare and swap 的缩写(比较并交换)
> [Wiki百科](https://zh.wikipedia.org/zh-hans/%E6%AF%94%E8%BE%83%E5%B9%B6%E4%BA%A4%E6%8D%A2)

在SQL中比较是 `WHERE`,交换是 `SET`.

### 企业认证审核

```sql
CREATE TABLE `cas_ company _auth` (
  `user_id` int(11) unsigned NOT NULL,
  `status` tinyint(3) unsigned NOT NULL,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
INSERT INTO `cas_company_auth` (`user_id`, `status`)
VALUES
	(100, 1);
```

```sql
UPDATE `cas_company_auth` 
SET status = 2 -- 通过
WHERE status = 1 AND user_id = 100
-- status 1 待审核 2 通过 3 拒绝
```

获取SQL执行返回的`rows affected`,如果返回 `1` 则表示修改成功,如果返回 `0` 则表示修改失败.

### 库存扣减

```sql
CREATE TABLE `cas_ inventory` (
    `sku_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `inventory` int(11) unsigned NOT NULL COMMENT '剩余库存',
    `cost` int(11) unsigned NOT NULL COMMENT '出库数量',
    PRIMARY KEY (`sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
INSERT INTO `cas_inventory` (`sku_id`, `inventory`, `cost`)
VALUES
	(100, 5, 0);
```

业务需求是下单时 减少 `inventory` 增加 `cost`,当 `inventory` 为 0 时不修改数据

````sql
-- 用户下单购买2件 sku_id 为 100 的商品
UPDATE 
	`cas_inventory` 
SET 
	 `inventory` = `inventory` - 2
	,`cost` = `cost` + 2
WHERE 
		sku_id = 100
	AND inventory >= 2
LIMIT 1
````

SQL执行后返回的`rows affected` 为 `1` 则下单成功,为 `0` 则库存不足下单失败.

如果 `WHERE` 条件的库存比较部分是 `inventory - 2 >= 0` 会导致报错 `BIGINT UNSIGNED value is out of range in '(`be`.`cas_inventory`.`inventory` - 2)'`,
 减法运算后 inventory 的值变成了负数, 而 inventory 字段是 `unsigned`.不允许出现负数.


**错误的SQL:**

```sql
-- 错误的SQL
UPDATE 
	`cas_inventory` 
SET 
	 `inventory` = `inventory` - 2
	,`cost` = `cost` + 2
WHERE 
		sku_id = 100
	AND inventory - 2 >= 0
	-- 当 inventory > 2时候会报错
    -- BIGINT UNSIGNED value is out of range in '(`be`.`cas_inventory`.`inventory` - 2)'
LIMIT 1
```

---

乐观锁CAS是高并发利器,实现简单且高性能.应该尽量使用乐观锁而不是[#tx-lock](事务锁).

基于乐观锁还能实现[#queue](高并发队列)

## 事务原子性 <a id="tx-atomicity"></a> 

## 事务锁 <a id="tx-lock"></a>

> 官方文档 https://dev.mysql.com/doc/refman/8.0/en/innodb-locking-reads.html


开启事务后对数据进行 SELECT 查询并不会锁定数据,在并发时会出现数据不一致.

可以简单的讲 `SELECT ... FOR SHARE` 理解为读锁, `SELECT ... FOR UPDATE` 理解为写锁.



## 高并发队列 <a id="queue"></a>
