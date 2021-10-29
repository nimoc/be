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
CREATE TABLE mass_data (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  batch_id int(10) unsigned NOT NULL,
  content varchar(20) NOT NULL,
  PRIMARY KEY (id),
  KEY bi (batch_id,id),
  KEY ib (id,batch_id)
) ENGINE=InnoDB AUTO_INCREMENT=10000001 DEFAULT CHARSET=utf8mb4;
	`, nil,) ; if err != nil {
						return
					}
	// 插入一千五的数据
	for insertCount := 1; insertCount< 3; insertCount++ {
		// 共插入一千五
		for i := 0; i < 1000; i++ {
			var insertValues [][]interface{}
			// 每次插入1万(根据实际数据库性能/数据库配置/插入列数决定)
			rows := 10000
			for j := 0; j < rows; j++ {
				var content []byte
				content, err = xrand.BytesBySeed(xrand.Seed{}.Alphabet(), 10) ; if err != nil {
					return
				}
				var batchID uint64
				batchID, err = xrand.RangeUint64(1, 40) ; if err != nil {
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
		}
	}
	return
}
