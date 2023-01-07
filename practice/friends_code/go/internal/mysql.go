package friend

import (
	"context"
	"errors"
	"github.com/goclub/sql"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)
type MysqlFriend struct {
	// 为了便于理解,暂时不做数据层和逻辑层的划分
	db *sq.Database
}
type MysqlConfig struct {
	DB sq.MysqlDataSource `yaml:"db"`
}
func NewMysqlFriend () (*MysqlFriend, error) {
	config := MysqlConfig{}

	data, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "src/github.com/nimoc/be/practice/friends_code/go/env.yaml")) ; if err != nil {
	    return nil, err
	}
	err = yaml.Unmarshal(data, &config) ; if err != nil {
	    return nil, err
	}
	var db *sq.Database
	db, _, err = sq.Open("mysql", config.DB.FormatDSN()) ; if err != nil {
	    return nil, err
	}
	return &MysqlFriend{
		db: db,
	}, nil
}

func (dep MysqlFriend) Add(ctx context.Context, userID int64, friendUserID int64) (error error) {
	query := "INSERT IGNORE INTO `user_friend` (`user_id`,`friend_user_id`) VALUES (?,?)"
	{ // 可折叠这段代码

		if userID == friendUserID { return errors.New("can_not_add_yourself") } // @TODO sentry

		if userID == 0 { return errors.New("userID error") } // @TODO sentry
		has, err :=  dep.hasUserID(ctx, userID); if err != nil { return }
		if has == false { return errors.New("friendUserID not found") } // @TODO sentry

		if friendUserID == 0 { return errors.New("friendUserID error") } // @TODO sentry
		has, err =  dep.hasUserID(ctx, friendUserID); if err != nil { return }
		if has == false { return errors.New("friendUserID not found") } // @TODO sentry
	}

	// 排序主键
	firstUserID, secondUserID := dep.SortUserID(userID, friendUserID)
	// 插入数据
	result, err := dep.db.Insert(ctx, sq.QB{
		Raw: sq.Raw{
			Query: query,
			Values: []interface{}{firstUserID, secondUserID},
		},
	}) ; if err != nil {
	    return
	}
	affected, err := result.RowsAffected() ; if err != nil {
	    return
	}
	// 根据插入结果判断是否重复
	switch affected {
	case 0:
		return errors.New("repeat")
	case 1:
		return nil
	default:
		return errors.New("MysqlFriend{}.Add() error") // @TODO sentry
	}
}
func (dep MysqlFriend) hasUserID(ctx context.Context, userID int64) (has bool, err error) {
	query :=  "SELECT 1 FROM `user` WHERE `id` = ? LIMIT 1"
	return dep.db.Has(ctx, sq.QB{
		Raw: sq.Raw{
			Query: query,
			Values: []interface{}{userID},
		},
	})
}
// 排序2个id,用于确保 SQL PrimaryKey 约束顺序一致
func (dep MysqlFriend) SortUserID(userID int64, friendUserID int64) (firstUserID int64, secondUserID int64) {
	if userID < friendUserID {
		return userID, friendUserID
	}
	if userID > friendUserID {
		return friendUserID, userID
	}
	// userID == friendUserID
	return userID, friendUserID
}

func (dep MysqlFriend) Is(ctx context.Context, userID int64, friendUserID int64) (isFriend bool, err error) {
	query := "SELECT 1 FROM `user_friend` WHERE `user_id` = ? AND `friend_user_id` = ? LIMIT 1"
	{ // 可折叠这段代码
		if userID == friendUserID { return false, errors.New("can_not_compare_yourself") } // @TODO sentry

		if userID == 0 { return false, errors.New("userID error") } // @TODO sentry
		var has bool
		has, err =  dep.hasUserID(ctx, userID); if err != nil { return }
		if has == false { return false, errors.New("friendUserID not found") } // @TODO sentry


		if friendUserID == 0 { return false, errors.New("friendUserID error") } // @TODO sentry
		has, err =  dep.hasUserID(ctx, friendUserID); if err != nil { return }
		if has == false { return false, errors.New("friendUserID not found") } // @TODO sentry
	}
	// 排序主键
	firstUserID, secondUserID := dep.SortUserID(userID, friendUserID)
	// 查询数据
	return dep.db.Has(ctx, sq.QB{
		Raw: sq.Raw{
			Query: query,
			Values: []interface{}{firstUserID, secondUserID},
		},
	})
}
func (dep MysqlFriend) List (ctx context.Context, userID int64) (userIDList []int64, err error) {
	query := `
			SELECT friend_user_id
			FROM user_friend
			WHERE user_id = ?
			UNION
			SELECT user_id
			FROM user_friend
			WHERE friend_user_id = ?
			`
	{ // 可折叠这段代码
		if userID == 0 { return nil, errors.New("userID error") } // @TODO sentry
		var has bool
		has, err =  dep.hasUserID(ctx, userID); if err != nil { return }
		if has == false { return nil, errors.New("friendUserID not found") } // @TODO sentry
	}
	err = dep.db.QuerySliceScaner(ctx, sq.QB{
		Raw: sq.Raw{
			Query: query,
			Values: []interface{}{userID, userID},
		},
	}, func(rows *sqlx.Rows) error {
		// 扫描id到userIDList
		var id int64
		err := rows.Scan(&id) ; if err != nil {
		    return err
		}
		userIDList = append(userIDList, id)
		return nil
	}) ; if err != nil {
	    return
	}
	return
}


func (dep MysqlFriend) Delete(ctx context.Context, userID int64, friendUserID int64) (err error) {
	query := "DELETE FROM `user_friend` WHERE `user_id` = ? AND `friend_user_id` = ? LIMIT 1"
	{ // 可折叠这段代码
		if userID == friendUserID { return  errors.New("can_not_delete_yourself") } // @TODO sentry

		if userID == 0 { return errors.New("userID error") } // @TODO sentry
		var has bool
		has, err =  dep.hasUserID(ctx, userID); if err != nil { return }
		if has == false { return errors.New("friendUserID not found") } // @TODO sentry


		if friendUserID == 0 { return  errors.New("friendUserID error") } // @TODO sentry
		has, err =  dep.hasUserID(ctx, friendUserID); if err != nil { return }
		if has == false { return  errors.New("friendUserID not found") } // @TODO sentry
	}
	firstUserID, secondUserID := dep.SortUserID(userID, friendUserID)
	result, err := dep.db.Exec(ctx,
		query,
		[]interface{}{firstUserID, secondUserID},
	) ; if err != nil {
	    return
	}
	affected, err := result.RowsAffected() ; if err != nil {
		return
	}
	// 根据删除结果判断是否重复
	switch affected {
	case 0:
		return errors.New("not friends")
	case 1:
		return nil
	default:
		return errors.New("MysqlFriend{}.Delete() error") // @TODO sentry
	}
}
func (dep MysqlFriend) Mutual(ctx context.Context, userID int64, friendUserID int64) (userIDList []int64, err error) {
	{ // 可折叠这段代码
		if userID == friendUserID { return nil, errors.New("can_not_compare_yourself") } // @TODO sentry

		if userID == 0 { return nil, errors.New("userID error") } // @TODO sentry
		var has bool
		has, err =  dep.hasUserID(ctx, userID); if err != nil { return }
		if has == false { return nil, errors.New("friendUserID not found") } // @TODO sentry


		if friendUserID == 0 { return nil, errors.New("friendUserID error") } // @TODO sentry
		has, err =  dep.hasUserID(ctx, friendUserID); if err != nil { return }
		if has == false { return nil, errors.New("friendUserID not found") } // @TODO sentry
	}
	// 排序主键
	firstUserID, secondUserID := dep.SortUserID(userID, friendUserID)
	err = dep.db.QuerySliceScaner(ctx, sq.QB{
		Raw: sq.Raw{
			Query: `
SELECT a.user_id FROM
(
	SELECT user_id
		FROM user_friend
		WHERE friend_user_id = ?
	UNION
	SELECT friend_user_id AS user_id
		FROM user_friend
		WHERE user_id = ?
) AS a
INNER JOIN
(
	SELECT user_id
		FROM user_friend
		WHERE friend_user_id = ?
	UNION
	SELECT friend_user_id AS user_id
		FROM user_friend
		WHERE user_id = ?
) AS b
ON (a.user_id = b.user_id)
			`,
			Values: []interface{}{
				firstUserID,
				firstUserID,

				secondUserID,
				secondUserID,

			},
		},
	}, func(rows *sqlx.Rows) error {
		// 扫描id到userIDList
		var id int64
		err := rows.Scan(&id) ; if err != nil {
			return err
		}
		userIDList = append(userIDList, id)
		return nil
	}) ; if err != nil {
		return
	}
	return
}