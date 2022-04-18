package main

import (
	"context"
	goredisv8 "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/error"
	red "github.com/goclub/redis"
	"github.com/goclub/sql"
)

func main() {
	xerr.PrintStack(run())
}

// 如果先取消表的索引插入数据,速度会快很多
func run() (err error) {
	var visit uint8 = 1
	var exposure uint8 = 2
	testData := []struct {
		userID uint32
		mktID uint32
		recordType uint8
	}{
		{1,101,exposure},
		{1,101,visit},
		
		{2,101,exposure},
		{2,102,exposure},
		
		{3,101,exposure},
		{3,102,visit},
		
		{4,102,exposure},
		{4,102,visit},
	}
	for _, datum := range testData {
		err = createMKTRecord(datum.userID, datum.mktID, datum.recordType) ; if err != nil {
		    xerr.PrintStack(err)
		}
	}
	return
}

func createMKTRecord(userID uint32, mktID uint32,recordType uint8) (err error) {
	// db
	// redisClient
	return
}
var db *sq.Database
var redisClient red.GoRedisV8
func init () {
	var err error
	ctx := context.Background()
	coreRedis := goredisv8.NewClient(&goredisv8.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	err = coreRedis.Ping(ctx).Err();
	if err != nil {
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
	}.FormatDSN())
	if err != nil {
		return
	}
}