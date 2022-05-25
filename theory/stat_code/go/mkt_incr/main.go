package main

import (
	"context"
	goredisv8 "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	xerr "github.com/goclub/error"
	red "github.com/goclub/redis"
	sq "github.com/goclub/sql"
	xtime "github.com/goclub/time"
	"strconv"
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
		err = createMKTRecordRedis(datum.userID, datum.mktID, datum.recordType); if err != nil {
			xerr.PrintStack(err)
		}
	}

	// 每日凌晨1点将redis中的数据合计存储到sql中
	yesterday := xtime.FormatChinaDate(time.Now()) // 应查前一天数据，因为示例演示查刚插入的数据
	err = cronCreateMKTRecordOfDay(yesterday); if err != nil {
		return
	}
	return
}

func createMKTRecordRedis(userID uint32, mktID uint32, recordType uint8) (err error) {
	ctx := context.Background()
	date := xtime.FormatChinaDate(time.Now()) // 2022-01-01

	isUV := false
	isUE := false
	evalReply, _, err := redisClient.Eval(ctx, red.Script{
		KEYS:   []string{
			/* 1 mkt:is_uv:${date} */
			"mkt:is_uv:"+date,
			/* 2 ${userID}-${mktID} */
			strconv.FormatUint(uint64(userID), 10) +"-"+ strconv.FormatUint(uint64(mktID), 10),
			/* 3 mkt:uv:${date} */
			`mkt:uv:`+date,
			/* 4 mkt:visit:${date} */
			`mkt:visit:`+date,
			/* 5 mktID */
			strconv.FormatUint(uint64(mktID), 10),
			/* 6 mkt:is_ue:${date} */
			"mkt:is_ue:"+date,
			/* 7 mkt:ue:${date} */
			`mkt:ue:`+date,
			/* 8 mkt:exposure:${date} */
			`mkt:exposure:`+date,
		},
		ARGV:   []string{},
		Script: `
			local isUV = 0
			local isUE = 0

			-- HSETNX mkt:is_uv:${date} ${userID}-${mktID}   
			local hsetReply = redis.call("HSETNX", KEYS[1], KEYS[2], 1)
			if hsetReply == 1 then
				-- HINCRBY mkt:uv:${date} ${mktID} 1
				redis.call("HINCRBY", KEYS[3], KEYS[5], 1)
				isUV = 1
			else
				-- HINCRBY mkt:visit:${date} ${mktID} 1
				redis.call("HINCRBY", KEYS[4], KEYS[5], 1)
				isUV = 0
			end

			-- HSETNX mkt:is_ue:${date} ${userID}-${mktID}   
			local hsetReply = redis.call("HSETNX", KEYS[6], KEYS[2], 1)
			if hsetReply == 1 then
				-- HINCRBY mkt:ue:${date} ${mktID} 1
				redis.call("HINCRBY", KEYS[7], KEYS[5], 1)
				isUE = 1
			else
				-- HINCRBY mkt:exposure:${date} ${mktID} 1
				redis.call("HINCRBY", KEYS[8], KEYS[5], 1)
				isUE = 0
			end

			return {isUV,isUE}
		`,
	}); if err != nil {
		return
	}
	var evalReplyInt []red.OptionInt64
	evalReplyInt, err = evalReply.Int64Slice(); if err != nil {
		return
	}
	if evalReplyInt[0].Int64 == 1 {
		isUV = true
	}
	if evalReplyInt[1].Int64 == 1 {
		isUE = true
	}

	// 为了易于理解省略 ue 判断和递增的代码
	// db 插入数据
	_, err = db.Insert(ctx, sq.QB{
		Raw: sq.Raw{
			Query:  "INSERT INTO `mkt_record` (`mkt_id`, `user_id`, `type`, `is_uv`, `is_ue`, `date`) VALUES (?,?,?,?,?,?)",
			Values: []interface{}{
				mktID, userID, recordType, isUV, isUE, date,
			},
		},
	}); if err != nil {
		return
	}

	return
}

// 每日凌晨1点将redis中的数据通过 HGETALL 读取出来存储到 mysql 的 mkt_stat_of_date 表中. 随后删除redis中对应的数据
func cronCreateMKTRecordOfDay(targetDate string) (err error) {
	ctx := context.Background()
	type stat struct{
		Exposure uint64
		UE       uint64
		Visit    uint64
		UV       uint64
	}
	list := map[uint32]stat{}

	// "mkt:uv:2022-04-19"
	err = cronCreateMKTRecordOfDayCore(ctx, "mkt:uv:" + targetDate, func(mktID uint32, value uint64) {
		if _,ok:=list[mktID];ok==false{
			list[mktID] = stat{}
		}else{
			item := list[mktID]
			item.UV = value
			list[mktID] = item
		}
	}); if err != nil {
		return
	}
	// "mkt:visit:2022-04-19"
	err = cronCreateMKTRecordOfDayCore(ctx, "mkt:visit:" + targetDate, func(mktID uint32, value uint64) {
		if _,ok:=list[mktID];ok==false{
			list[mktID] = stat{}
		}else{
			item := list[mktID]
			item.Visit = value
			list[mktID] = item
		}
	}); if err != nil {
		return
	}
	// "mkt:exposure:2022-04-19"
	err = cronCreateMKTRecordOfDayCore(ctx, "mkt:exposure:" + targetDate, func(mktID uint32, value uint64) {
		if _,ok:=list[mktID];ok==false{
			list[mktID] = stat{}
		}else{
			item := list[mktID]
			item.Exposure = value
			list[mktID] = item
		}
	}); if err != nil {
		return
	}
	// "mkt:ue:2022-04-19"
	err = cronCreateMKTRecordOfDayCore(ctx, "mkt:ue:" + targetDate, func(mktID uint32, value uint64) {
		if _,ok:=list[mktID];ok==false{
			list[mktID] = stat{}
		}else{
			item := list[mktID]
			item.UE = value
			list[mktID] = item
		}
	}); if err != nil {
		return
	}
	// 遍历存储数据
	sqlValueList := [][]interface{}{}
	for mktID, item := range list {
		sqlValueList = append(sqlValueList, []interface{}{
			targetDate, mktID, item.Exposure, item.UE, item.Visit, item.UV,
		})
	}
	_, err = db.Insert(ctx, sq.QB{
		From:                sq.Table("mkt_stat_of_date", nil, nil),
		InsertMultiple:      sq.InsertMultiple{
			Column: []sq.Column{
				`date`, `mkt_id`, `exposure`, `ue`, `visit`, `uv`,
			},
			Values: sqlValueList,
		},
		UseInsertIgnoreInto: true,
	}); if err != nil {
		return
	}
	// 删除3天以前的数据，留3天内数据以便观测
	threeDayBefore := xtime.FormatChinaDate(time.Now().AddDate(0,0,-4))
	_, err = redisClient.DoIntegerReplyWithoutNil(ctx, []string{
		"DEL",
		"mkt:is_ue:"+threeDayBefore,
		"mkt:exposure:"+threeDayBefore,
		"mkt:ue:"+threeDayBefore,
		"mkt:is_uv:"+threeDayBefore,
		"mkt:visit:"+threeDayBefore,
		"mkt:uv:"+threeDayBefore,
	}); if err != nil {
		return
	}
	return
}
func cronCreateMKTRecordOfDayCore(ctx context.Context, redisField string,  handleItem func(mktID uint32, value uint64))(err error){
	var reply []red.OptionInt64
	reply, err = redisClient.DoArrayIntegerReply(ctx, []string{"HGETALL", redisField}); if err != nil {
		return
	}
	var mktID uint32
	for i, optionInt64 := range reply {
		var isKey bool
		if i%2 == 0 { isKey = true }
		if isKey {
			mktID = uint32(optionInt64.Int64)
		}else{
			handleItem(mktID, uint64(optionInt64.Int64))
		}
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
		panic(err)
	}
}
