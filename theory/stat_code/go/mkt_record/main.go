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
	ctx := context.Background()
	coreRedis := goredisv8.NewClient(&goredisv8.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	err = coreRedis.Ping(ctx).Err();
	if err != nil {
		err = xerr.WrapPrefix("redis ping fail:", err)
		return
	}
	redisClient := red.NewGoRedisV8(coreRedis)
	// red.GET{Key: "some"}.Do(ctx, redisClient)
	db, dbClose, err := sq.Open("mysql", sq.MysqlDataSource{
		Host:     "127.0.0.1",
		Port:     "3306",
		DB:       "be",
		User:     "root",
		Password: "somepass",
	}.FormatDSN())
	if err != nil {
		return
	}
	defer dbClose()
	return
}