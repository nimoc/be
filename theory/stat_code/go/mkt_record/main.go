package main

import (
	"context"
	goredisv8 "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/error"
	red "github.com/goclub/redis"
	"github.com/goclub/sql"
	xtime "github.com/goclub/time"
	"strconv"
	"strings"
	"time"
)

func main() {
	xerr.PrintStack(run())
}

// 如果先取消表的索引插入数据,速度会快很多
func run() (err error) {
	var visit uint8 = 1
	var exposure uint8 = 2
	testData := []struct {
		userID     uint32
		mktID      uint32
		recordType uint8
	}{
		{1, 101, exposure},
		{1, 101, visit},

		{2, 101, exposure},
		{2, 102, exposure},

		{3, 101, exposure},
		{3, 102, visit},

		{4, 102, exposure},
		{4, 102, visit},
	}
	for _, datum := range testData {
		err = createMKTRecord(datum.userID, datum.mktID, datum.recordType); if err != nil {
			xerr.PrintStack(err)
		}
	}
	return
}

func createMKTRecord(userID uint32, mktID uint32, recordType uint8) (err error) {
	ctx := context.Background()
	// 今天日期
	date := xtime.FormatChinaDate(time.Now()) // 2022-01-01

	// 用户每天对一个广告只会产生一次uv 独立访问
	isUV := false
	// 用户每天对一个广告只会产生一次ue 独立曝光
	isUE := false

	switch recordType {
	case 1: // 访问 visit
		// 通过 redis 获知本次操作是不是uv
		/* 设置成功，返回1。 已经存在且没有操作被执行，返回0。 */
		{
			redisKey := strings.Join([]string{"mkt","is_uv",date},":") // mkt:is_uv:${date}
			redisField := strconv.FormatUint(uint64(userID), 10) +"-"+ strconv.FormatUint(uint64(mktID), 10)  // ${userID}-${mktID}
			var hsetIsUVReply int64
			hsetIsUVReply, _, err = redisClient.DoIntegerReply(ctx, []string{"HSETNX", redisKey, redisField, `1`}); if err != nil {
				return
			}
			if hsetIsUVReply == 1 {
				isUV = true
			}
		}

	case 2: // 曝光 exposure
		// 通过 redis 获知本次操作是不是ue
		{
			redisKey := strings.Join([]string{"mkt","is_ue",date},":") // mkt:is_ue:${date}
			redisField := strconv.FormatUint(uint64(userID), 10) +"-"+ strconv.FormatUint(uint64(mktID), 10)  // ${userID}-${mktID}
			var hsetIsUEReply int64
			hsetIsUEReply, _, err = redisClient.DoIntegerReply(ctx, []string{"HSETNX", redisKey, redisField, `1`});if err != nil {
				return
			}
			if (hsetIsUEReply == 1) {
				isUE = true
			}
		}

	}

	// db 插入数据
	_, err = db.Insert(ctx, sq.QB{
		Raw: sq.Raw{
			Query:  "INSERT INTO `mkt_record` (`mkt_id`, `user_id`, `type`, `is_uv`, `is_ue`, `date`) VALUES (?, ?, ?, ?, ?, ?)",
			Values: []interface{}{
				mktID, userID, recordType, isUV, isUE, date,
			},
		},
	}); if err != nil {
		return
	}
	return
}

var db *sq.Database
var redisClient red.GoRedisV8

func init() {
	var err error
	ctx := context.Background()
	coreRedis := goredisv8.NewClient(&goredisv8.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	err = coreRedis.Ping(ctx).Err(); if err != nil {
		err = xerr.WrapPrefix("redis ping fail:", err)
		panic(err)
	}
	redisClient = red.NewGoRedisV8(coreRedis)
	// red.GET{Key: "some"}.Do(ctx, redisClient)
	db, _, err = sq.Open("mysql", sq.MysqlDataSource{
		Host:     "127.0.0.1",
		Port:     "3306",
		DB:       "be",
		User:     "root",
		Password: "somepass",
	}.FormatDSN()); if err != nil {
		return
	}
}
