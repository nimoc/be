---
permalink: /theory/sql/
---

# sql

## 幂等性 <a id="idempotent"></a>


> 在编程中一个幂等操作的特点是其任意多次执行所产生的影响均与一次执行的影响相同。幂等函数，或幂等方法，是指可以使用相同参数重复执行，并能获得相同结果的函数。这些函数不会影响系统状态，也不用担心重复执行会对系统造成改变。例如，“setTrue()”函数就是一个幂等函数,无论多次执行，其结果都是一样的.更复杂的操作幂等保证是利用唯一交易号(流水号)实现。 

知识点:

1.`PRIMARY KEY` 和 `UNIQUE` 的唯一约束
2.`INSERT IGNORE INTO`

### 业务场景: 判断UV

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


### 业务场景: 订单号

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


## 事务锁 <a id="tx-lock"></a>

> 官方文档 https://dev.mysql.com/doc/refman/8.0/en/innodb-locking-reads.html


开启事务后对数据进行 SELECT 查询并不会锁定数据,在并发时会出现数据不一致.

可以简单的讲 `SELECT ... FOR SHARE` 理解为读锁, `SELECT ... FOR UPDATE` 理解为写锁.



