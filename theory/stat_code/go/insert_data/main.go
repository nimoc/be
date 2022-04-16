package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/error"
	"github.com/goclub/test"
	"log"
	"strconv"
	"time"
)
import "github.com/goclub/sql"

func main() {
	xerr.PrintStack(run())
}
// 如果先取消表的索引插入数据,速度会快很多
func run() (err error) {
	ctx := context.Background()
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
	var values [][]interface{}
	var column []sq.Column
	mktIDSeed := uintSeed(1, 500)     // 500个广告
	userIDSeed := uintSeed(1, 100000) // 十万个用户
	mockRedis := struct {
		IsUV map[string]bool
		IsUE map[string]bool
	}{
		IsUV: map[string]bool{},
		IsUE: map[string]bool{},
	}
	for i := 0; i < 10000; i++ {
		log.Print("insert batch:", i)
		for j := 0; j < 9000; j++ {
			mktID := xtest.PickOne(mktIDSeed)
			userID := xtest.PickOne(userIDSeed)
			kind := xtest.PickOne([]uint8{1,1,1,1,1, 2}) // 1 exposure 2 visit
			createTime := time.
				Date(2022, 1, 1, 0, 0, 0, 0, time.Local).
				AddDate(0, 0, i/100)
			date := createTime.Format("2006-01-02")
			isUV := false
			isUE := false
			// mock redis setnx
			setKey := date + ":" + strconv.FormatUint(uint64(userID), 10)
			if kind == 1 {
				if _, ok := mockRedis.IsUV[setKey]; !ok {
					mockRedis.IsUV[setKey] = true
					isUV = true
				}
			}
			if kind == 2  {
				if _, ok := mockRedis.IsUE[setKey]; !ok {
					mockRedis.IsUE[setKey] = true
					isUE = true
				}
			}
			// 故意在啊循环中一直设置 column,目的是为了在测试环境便于对照 values
			column = []sq.Column{
				"mkt_id", "user_id", "kind", "is_uv", "is_ue", "date", "create_time"}
			values = append(values, []interface{}{
				mktID, userID, kind, isUV, isUE, date, createTime,
			})
		}
		_, err = db.Insert(ctx, sq.QB{
			From: sq.Table("mkt_record", nil, nil),
			InsertMultiple: sq.InsertMultiple{
				Column: column,
				Values: values,
			},
		}) ; if err != nil {
			return
		}
		values = nil
	}
	return
}
func uintSeed(min uint, max uint) (seed []uint) {
	var i uint = 0
	for ; i < max-min; i++ {
		seed = append(seed, min+i)
	}
	return
}
