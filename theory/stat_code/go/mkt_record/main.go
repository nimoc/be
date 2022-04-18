package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goclub/error"
	"github.com/goclub/sql"
)

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
	return 
}