---
permalink: /practice/finance/
---

# 财务


## sql表结构

```sql
CREATE TABLE `account_finance` (
  `account_id` bigint(20) unsigned NOT NULL,
  `balance` bigint(20) unsigned NOT NULL DEFAULT '0',
  `cost` bigint(20) unsigned NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 表字段

### 账户ID:

`account_id` 作为主键,避免出现同一个账户出现多个重复数据

> 如果使用自增id作为主键 `id bigint(20) unsigned NOT NULL AUTO_INCREMENT`
> 则必须设置  `UNIQUE KEY account_id (account_id)` 来避免重复数据.


### 余额与支出

余额: `balance`
支出: `cost`
单位: `分`
unsigned: `true`

> 100 = 1元 
> 10  = 0.1元

**单位:**
使用分为单位是为了避免浮点计算的精度问题(编程语言计算 float , redis 递增 float 会出现问题)

**unsigned:**
通过 unsigned 限制不允许出现负余额和负支出能避免代码bug导致的数据错误,

> 如果业务需要实现负余额应该配合欠费表实现,而不是允许 `balance` 出现负数

### 累计余额:

累计余额 = `balance` + `cost`

累计余额不使用字段存储.必须通过计算得出.这样保障了数据结构的简洁,减少了代码复杂度.

## 插入数据

第一种情况是在用户注册的时就插入 account_finance 数据

```js
// 开始事务
sql("BEGIN")

// 插入账户信息
result = sql("INSERT INTO `account` (mobile) VALUES (13111112222)")
accountID = result.lastInsertID

// 插入账户财务信息
sql("INSERT INTO `account_finance` (`account_id`, `balance`, `cost`) VALUES (?, 0, 0)", accountID)

// 提交事务
sql("COMMIT")
```

第二种情况是业务初期没有账户财务表,迭代的过程中出现的 `account_finance` 表.
这种情况下每次读写 `account_finance` 时需要尝试插入

```js
function incomeBalance(accountID, amount) {
    sql("INSERT IGNORE INTO `account_finance` (`account_id`, `balance`, `cost`) VALUES (?, 0, 0)", accountID)
    // TODO: 增加余额
}

// 账户 1 增加1元
incomeBalance(1, 100)
```

`account_id` 是主键,如果重复插入 `account_id = 1` 的数据会导致sql返回错误,是用 `INSERT IGNORE INTO` 可以忽略错误.

参考: [幂等性](../theory/sql.md#idempotent)


## 增加余额

增加余额只需要增加 balance

```sql
UPDATE account_finance
SET 
    balance = balance + 10 
WHERE 
    account_id = 1 
LIMIT 1
```

## 扣除余额


扣除余额需要使用 [CAS](../theory/sql.md#cas) 

```sql
UPDATE account_finance
SET
    balance = balance - 4, cost = cost + 4
WHERE
    account_id = 1 AND
    balance >= 4
LIMIT 1
```

```js
result = sql("UPDATE `account_finance` SET balance = balance - 4, cost = cost + 4 WHERE `account_id` = 1 AND balance >= 4 LIMIT 1")
if (result.rowsAffected == 0) {
    return "余额不足"
}
```

账户1的余额为10的时候
```js
{
    account_id: 1,
    balance: 10,
    cost: 0,
}
```

执行三次 UPDATE

前两次 `rowsAffected = 1` 最后一次 `rowsAffected = 0`

因为 执行第二次后 `balance = 2`,不满足 where 条件中的 `balance >= 4`


### 错误的方法

> 认识错误才能避免错误

**1. 读写存在执行间隙,并发时会多扣**

```js
// 错误的方法!
accountID = 1
deduct = 4
// 读
balance,cost = sql("SELECT balance,cost FROM account_finance WHERE account_id = > LIMIT 1", accountID)
if (balance >= deduct) {
    newBalance = balance - deduct
    newCost = cost + deduct
    // 写
    sql(`
        UPDATE account_finance
        SET
            balance = ?, cost = ?
        WHERE
            account_id = ?
        LIMIT 1
    `, newBalance, newCost, accountID)    
} else {
    return "余额不足"
}
```

*2. version 锁并发时会出现大量失败*

首先增加 version 字段 (int(11))

```js
// 错误的方法!
accountID = 1
deduct = 4
// 读
balance,cost,version = sql("SELECT balance,cost,version FROM account_finance WHERE account_id = > LIMIT 1", accountID)
if (balance >= deduct) {
    newBalance = balance - deduct
    newCost = cost + deduct
    // 写
    sql(`
        UPDATE account_finance
        SET
            balance = ?, cost = ?, version = version + 1
        WHERE
            account_id = ?
            version = ?
        LIMIT 1
    `, newBalance, newCost, accountID, version)
} else {
    return "余额不足"
}
```

使用 version 字段或者 update_time 字段进行 CAS 操作能保障不会多扣钱.
但是在并发时会出现大量的扣款失败.没有 `balance >= ?` 好

**3.out of range 错误**

> 当 `balance` 字段是 `unsigned` 时,执行第三次会出现错误:
> 
> BIGINT UNSIGNED value is out of range in '(be.account_finance.balance - 4)'

```sql
-- 错误的方法!
UPDATE account_finance
SET
    balance = balance - 4, cost = cost + 4
WHERE
    account_id = 1 AND
    balance - 1 >= 0
LIMIT 1
```

最初设计表的时候忘记给 `balance` 加上 `unsigned`, 后续加上 `unsigned` 时会触发错误.

正确的方法是不进行减法运算 `balance >= 4`


**4. 疏忽: >= 写成 >**

账户1的余额为3的时候
```js
{
    account_id: 1,
    balance: 3,
    cost: 0,
}
```

```sql
UPDATE account_finance
SET
    balance = balance - 1, cost = cost + 1
WHERE
    account_id = 1 AND
    -- 这里写错了,写成了 > 1 而不是 >= 1
    balance > 1
LIMIT 1
```

执行3次sql只有前2次成功扣除.最后一次明明余额还有1却无法扣款.

## 财务记录

财务的变更的同时应当保存变更记录,便于排查问题和展示给用户.

### 表结构

    CREATE TABLE `account_finance_record` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `account_id` bigint(20) unsigned NOT NULL,
    `amount` bigint(20) unsigned NOT NULL,
    `type` tinyint(4) unsigned NOT NULL,
    `ref_type` tinyint(4) unsigned NOT NULL,
    `ref_value` char(24) NOT NULL DEFAULT '',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB AUTO_INCREMENT=356749 DEFAULT CHARSET=utf8mb4;


### amount
变更金额,单位分

### type
变更类型

| 值   | 说明        | 对应的数据变化                                |
|-----|-----------|----------------------------------------|
| `1` | 收入 income | balance = balance + ?                  |
| `2` | 支出 expend | balance = balance - ?, cost = cost + ? |
| `3` | 退款 refund | balance = balance + ?, cost = cost - ? |

### ref_type & ref_value
来源信息

> 非必须字段,可根据业务情况决定是否需要.主要用于排查问题追查金额来源


| ref_value | 说明          | ref_value                                   | 场景               |
|-----------|-------------|---------------------------------------------|------------------|
| `1`       | 充值 recharge | `100` 充值订单sql表主键id                          | 用户通过微信支付宝充值金额到账户 |
| `2`       | 抽奖 lottery  | `200` 抽奖记录表主键id                             | 用户参与活动抽奖并中奖      |
| `3`       | 提现 withdraw | `5349b4ddd2781d08c09890f4`  提现订单mongo集合主键id | 用户提现余额到微信或支付宝    |


### 示例

```
// 记录用户 1 收入 2 元. 来源于充值,充值订单id 100
insert({
    account_id: 1,
    type: 1,
    amount: 200,
    ref_type: 1,
    ref_value:intToString(100),
})
```

```
// 记录用户 1 收入 2 元.来源于抽奖,抽奖记录表id 200
insert({
    account_id: 1,
    type: 1,
    amount: 200,
    ref_type: 2,
    ref_value:intToString(200),
})
```

```
// 记录用户 1 支出 2 元. 来源于提现,提现mongo id 5349b4ddd2781d08c09890f4
insert({
    account_id: 1,
    type: 2,
    amount: 200,
    ref_type: 3,
    ref_value:"5349b4ddd2781d08c09890f4",
})
```