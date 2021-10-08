package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	xerr "github.com/goclub/error"
	sq "github.com/goclub/sql"
	"log"
)

func main() {
	err := run()  ; if err != nil {
		xerr.PrintStack(err)
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
	log.Print("range 查询第一页")
	var firstPage []Data
	{
		batchID := 12
		err = db.QuerySlice(ctx, &firstPage, sq.QB{
			Debug: true,
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			ORDER BY id asc
			LIMIT 10
			`,
				[]interface{}{batchID},
			},
		}) ; if err != nil {
			return
		}
		log.Print("数据: \n", firstPage)
	}
	log.Print("range 查询第二页")
	{
		var secondPage []Data
		var id uint64
		if len(firstPage) != 0 {
			id = firstPage[len(firstPage)-1].ID
		}
		batchID := 12
		err = db.QuerySlice(ctx, &secondPage, sq.QB{
			Debug: true,
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			AND id > ?
			ORDER BY id asc
			LIMIT 10
			`,
				[]interface{}{batchID, id},
			},
		}) ; if err != nil {
			return
		}
		log.Print("数据: \n", secondPage)

	}
	log.Print("range 查询最后几页")
	{
		var lastFewPage []Data
		batchID := 12
		var descFirstPage []Data
		err = db.QuerySlice(ctx, &descFirstPage, sq.QB{
			Debug: true,
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			ORDER BY id DESC
			LIMIT 10
			`,
				[]interface{}{batchID},
			},
		}) ; if err != nil {
			return
		}
		err = db.QuerySlice(ctx, &lastFewPage, sq.QB{
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			AND id > ?
			ORDER BY id ASC
			LIMIT 10
			`,
				[]interface{}{batchID, descFirstPage[len(descFirstPage)-1].ID},
			},
		}) ; if err != nil {
		return
	}
		log.Print("数据: \n", lastFewPage)
	}
	log.Print("使用 limit offset 查询")
	log.Print("offset 查询第一页")
	{
		batchID := 12
		var firstPage  []Data
		err = db.QuerySlice(ctx, &firstPage, sq.QB{
			Debug: true,
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			ORDER BY id ASC
			LIMIT 10 OFFSET 0
			`,
				[]interface{}{batchID},
			},
		}) ; if err != nil {
		return
	}
		log.Print("数据: \n", firstPage)
	}
	log.Print("offset 查询第二页")
	{
		batchID := 12
		var secondPage  []Data
		err = db.QuerySlice(ctx, &secondPage, sq.QB{
			Debug: true,
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			ORDER BY id ASC
			LIMIT 10 OFFSET 9
			`,
				[]interface{}{batchID},
			},
		}) ; if err != nil {
		return
	}
		log.Print("数据: \n", secondPage)
	}
	log.Print("offset 查询最后一页")
	{
		var count uint64
		count, err = db.Count(ctx, sq.QB{
			Raw: sq.Raw{
				`SELECT count(*) FROM mass_data WHERE batch_id=?`,
				[]interface{}{12},
			},
		}) ; if err != nil {
		    return
		}
		batchID := 12
		var lastPage  []Data
		err = db.QuerySlice(ctx, &lastPage, sq.QB{
			Debug: true,
			Raw: sq.Raw{
				`
			SELECT id, batch_id, content FROM mass_data
			USE INDEX(bi)
			WHERE batch_id = ?
			ORDER BY id ASC
			LIMIT 10 OFFSET ?
			`,
				[]interface{}{batchID, count-10},
			},
		}) ; if err != nil {
		return
	}
		log.Print("数据: \n", lastPage)
	}
	return
}

type Data struct {
	ID uint64 `db:"id"`
	BatchID uint64 `db:"batch_id"`
	Content string `db:"content"`
	sq.WithoutSoftDelete
}
func (data Data) TableName() string {
	return "mass_data"
}