package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	xerr "github.com/goclub/error"
	xrand "github.com/goclub/rand"
	"github.com/goclub/sql"
	"log"
)

func main () {
	err := run()  ; if err != nil {
		xerr.PrintStack(err)
	} else {
		log.Print("prepare_data done")
	}
}
func run () (err error) {
	ctx := context.Background()
	db, dbClose, err := sq.Open("mysql", sq.MysqlDataSource{
		User:     "root",
		// 生产环境账户密码通过读取环境变量或配置文件获取
		Password: "somepass",
		Host:     "localhost",
		Port:     "3306",
		DB:       "be_nimo_run",
		Query:    nil,
	}.FormatDSN()) ; if err != nil {
		return
	}
	defer dbClose()
	// 如果表不存在则创建表
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS mass_data (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  batch_id int(10) unsigned NOT NULL,
  content varchar(20) NOT NULL DEFAULT '',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`, nil,) ; if err != nil {
						return
					}
	// 插入5个批次
	for batchID := 1; batchID< 6; batchID++ {
		// 每个批次插入5百万
		for i := 0; i < 500; i++ {
			var insertValues [][]interface{}
			// 每次插入1万(根据实际数据库性能/数据库配置/插入列数决定)
			rows := 10000
			for j := 0; j < rows; j++ {
				var content []byte
				content, err = xrand.BytesBySeed(xrand.Seed{}.Alphabet(), 10) ; if err != nil {
					return
				}
				insertValues = append(insertValues, []interface{}{
					batchID, content,
				})
			}
			_, err = db.Insert(ctx, sq.QB{
				From: sq.Table("mass_data", nil, nil),
				InsertMultiple: sq.InsertMultiple{
					Column: []sq.Column{"batch_id", "content"},
					Values: insertValues,
				},
			}) ; if err != nil {
				return
			}
			var count uint64
			count, err = db.Count(ctx, sq.QB{From: sq.Table("mass_data", nil, nil)}) ; if err != nil {
			    return
			}
			log.Print("count:", count)
		}
	}
	return
}
