package friend

import (
	"context"
	redisv8 "github.com/go-redis/redis/v8"
	red "github.com/goclub/redis"
	sq "github.com/goclub/sql"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

type UnionFriend struct {
	// 为了便于理解,暂时不做数据层和逻辑层的划分
	db *sq.Database
	redis red.Connecter
}
type UnionConfig struct {
	DB sq.MysqlDataSource `yaml:"db"`
	Redis struct{
		Addr string `yaml:"addr"`
	} `yaml:"redis"`
}
func NewUnionFriend () (*UnionFriend, error) {
	config := UnionConfig{}
	data, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "src/github.com/nimoc/be/practice/friends_code/go/env.yaml")) ; if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &config) ; if err != nil {
		return nil, err
	}
	redis := red.GoRedisV8{
		Core: redisv8.NewClient(&redisv8.Options{
			Addr: config.Redis.Addr,
		}),
	}
	err = redis.Core.Ping(context.Background()).Err() ; if err != nil {
		return nil, err
	}
	var db *sq.Database
	db, _, err = sq.Open("mysql", config.DB.FormatDSN()) ; if err != nil {
		return nil, err
	}
	return &UnionFriend{
		db: db,
		redis: redis,
	}, nil
}
func (dep UnionFriend) KeyFriendSets(userID int64) string {
	return "friend:" + strconv.FormatInt(userID, 10)
}
func (dep UnionFriend) KeyNoFriendString(userID int64) string {
	return "no_friend:" + strconv.FormatInt(userID, 10)
}
func (dep UnionFriend) KeyFriendSyncing(userID int64) string {
	return "friend_syncing:" + strconv.FormatInt(userID, 10)
}
func (dep UnionFriend) Add(ctx context.Context, userID int64, friendUserID int64) (err error) {
	if userID == friendUserID { return errors.New("can_not_add_yourself") } // @TODO sentry
	err = dep.mustHasUserID(ctx, userID) ; if err != nil { return }
	err = dep.mustHasUserID(ctx, friendUserID) ; if err != nil { return }

	// 排序主键
	firstUserID, secondUserID := dep.SortUserID(userID, friendUserID)
	// 插入数据
	result, err := dep.db.Insert(ctx, sq.QB{
		Raw: sq.Raw{
			Query: "INSERT IGNORE INTO `user_friend` (`user_id`,`friend_user_id`) VALUES (?,?)",
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
		// 删除缓存
		// 此处可以不满足原子性,如果 redis del 没有执行只会造成短期的缓存与数据不一致(时间为 KeyFriendSets 的 ttl)
		_, err = red.DEL{
				Keys: []string{
					dep.KeyFriendSets(userID),
					dep.KeyFriendSets(friendUserID),

					dep.KeyNoFriendString(userID),
					dep.KeyNoFriendString(friendUserID),
				},
			}.Do(ctx, dep.redis) ; if err != nil {
			return
		}
		return nil
	default:
		return errors.New("MysqlFriend{}.Add() error") // @TODO sentry
	}
}

func (dep UnionFriend) mustHasUserID(ctx context.Context, userID int64) (err error) {
	// 如果想进一步提高性能,可以将 userid 放在redis 10 秒,减少数据库压力
	query :=  "SELECT 1 FROM `user` WHERE `id` = ? LIMIT 1"
	has, err := dep.db.Has(ctx, sq.QB{
		Raw: sq.Raw{
			Query: query,
			Values: []interface{}{userID},
		},
	}) ; if err != nil {
	    return
	}
	if has == false {
		return errors.New("userID error:" + strconv.FormatInt(userID, 10))  // @TODO sentry
	}
	return nil
}
func (dep UnionFriend) SortUserID(userID int64, friendUserID int64) (firstUserID int64, secondUserID int64) {
	if userID < friendUserID {
		return userID, friendUserID
	}
	if userID > friendUserID {
		return friendUserID, userID
	}
	// userID == friendUserID
	return userID, friendUserID
}
func (dep UnionFriend) SyncCache(ctx context.Context, userID int64) (noFriend bool, err error) {
	// 防止缓存穿透
	{
		var isNil bool
		_ , isNil, err = red.GET{
			Key: dep.KeyNoFriendString(userID),
		}.Do(ctx, dep.redis) ; if err != nil {
			return
		}
		if isNil == false {
			return true, nil
		}
	}
	existsCount, err := red.EXISTS{
		Key: dep.KeyFriendSets(userID),
	}.Do(ctx, dep.redis) ; if err != nil {
		return
	}
	if existsCount == 0 {
		var isNil bool
		// 防止缓存穿透
		_, isNil, err = red.SET{
			Key: dep.KeyFriendSyncing(userID),
			NX: true,
			Expire: time.Second*2,
		}.Do(ctx, dep.redis) ; if err != nil {
		    return
		}
		defer func() {
			_, delErr := red.DEL{Key: dep.KeyFriendSyncing(userID)}.Do(ctx, dep.redis) ; if delErr != nil {
			    log.Print(delErr) // sentry
			}
		}()
		if isNil {
			return false, errors.New("try_again_later")
		}
		members := []string{}
		err = dep.db.QuerySliceScaner(ctx, sq.QB{
			Raw: sq.Raw{
				Query: `
			SELECT friend_user_id
			FROM user_friend
			WHERE user_id = ?
			UNION
			SELECT user_id
			FROM user_friend
			WHERE friend_user_id = ?
			`,
				Values: []interface{}{userID, userID},
			},
		}, sq.ScanStrings(&members)) ; if err != nil {
			return
		}
		if len(members) == 0 {
			_, _, err = red.SET{
				Key: dep.KeyNoFriendString(userID),
				Expire: time.Second*2,
				Value: "1",
			}.Do(ctx, dep.redis) ; if err != nil {
			    return
			}
			return true, nil
		}
		_, isNil, err = red.Script{
			Keys: []string{dep.KeyFriendSets(userID)},
			Argv: members,
			Script: `
			local friendKey = KEYS[1]
			local members = ARGV
			for k, v in ipairs(ARGV) do
				redis.call("SADD",friendKey, v)
			end
			redis.call('EXPIRE', friendKey, '10')
			return 1
			`,
		}.Do(ctx, dep.redis) ; if err != nil {
		    return
		}
		if isNil {
			return false, errors.New("eval replay can not be nil")
		}
	}
	return
}
func (dep UnionFriend) Is(ctx context.Context, userID int64, friendUserID int64) (isFriend bool, err error) {
	if userID == friendUserID { return false, errors.New("can_not_compare_yourself") } // @TODO sentry
	err = dep.mustHasUserID(ctx, userID) ; if err != nil { return }
	err = dep.mustHasUserID(ctx, friendUserID) ; if err != nil { return }
	noFriend, err := dep.SyncCache(ctx, userID) ; if err != nil {
	    return
	}
	if noFriend {
		return
	}
	cmd := []string{"SISMEMBER", dep.KeyFriendSets(userID), strconv.FormatInt(friendUserID, 10) }
	reply, _, err := dep.redis.DoIntegerReply(ctx, cmd) ; if err != nil {
		return
	}
	return reply == 1, nil
}
func (dep UnionFriend) List (ctx context.Context, userID int64) (userIDList []int64, err error) {
	err = dep.mustHasUserID(ctx, userID) ; if err != nil { return }
	noFriend, err := dep.SyncCache(ctx, userID) ; if err != nil {
		return
	}
	if noFriend {
		return nil, nil
	}
	cmd := []string{"SMEMBERS", dep.KeyFriendSets(userID)}
	reply, err := dep.redis.DoArrayStringReply(ctx, cmd) ; if err != nil {
		return
	}
	for _, item := range reply {
		var userID int64
		userID, err = strconv.ParseInt(item.String, 10, 64) ; if err != nil {
			return
		}
		userIDList = append(userIDList, userID)
	}
	return
}

func (dep UnionFriend) Delete(ctx context.Context, userID int64, friendUserID int64) (err error) {
	if userID == friendUserID { return  errors.New("can_not_compare_yourself") } // @TODO sentry
	err = dep.mustHasUserID(ctx, userID) ; if err != nil { return }
	err = dep.mustHasUserID(ctx, friendUserID) ; if err != nil { return }

	firstUserID, secondUserID := dep.SortUserID(userID, friendUserID)
	result, err := dep.db.Exec(ctx,
		"DELETE FROM `user_friend` WHERE `user_id` = ? AND `friend_user_id` = ? LIMIT 1",
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
		// 删除缓存
		// 此处可以不满足原子性,如果 redis del 没有执行只会造成短期的缓存与数据不一致(时间为 KeyFriendSets 的 ttl)
		_, err = red.DEL{
			Keys: []string{
				dep.KeyFriendSets(userID),
				dep.KeyFriendSets(friendUserID),
			},
		}.Do(ctx, dep.redis) ; if err != nil {
			return
		}
		return nil
	default:
		return errors.New("MysqlFriend{}.Delete() error") // @TODO sentry
	}
}


func (dep UnionFriend) Mutual  (ctx context.Context, userID int64, friendUserID int64) (userIDList []int64, err error){
	if userID == friendUserID { return  nil, errors.New("can_not_compare_yourself") } // @TODO sentry
	err = dep.mustHasUserID(ctx, userID) ; if err != nil { return }
	err = dep.mustHasUserID(ctx, friendUserID) ; if err != nil { return }
	{
		var noFriend bool
		noFriend, err =  dep.SyncCache(ctx, userID) ; if err != nil {
		    return
		}
		if noFriend {
			return nil, nil
		}
	}
	{
		var noFriend bool
		noFriend, err =  dep.SyncCache(ctx, friendUserID) ; if err != nil {
			return
		}
		if noFriend {
			return nil, nil
		}
	}
	keyUserID := dep.KeyFriendSets(userID)
	keyFriendUserID := dep.KeyFriendSets(friendUserID)
	reply, err := dep.redis.DoArrayStringReply(ctx, []string{"SINTER", keyUserID, keyFriendUserID})
	for _, item := range reply {
		var userID int64
		userID, err = strconv.ParseInt(item.String, 10, 64) ; if err != nil {
			return
		}
		userIDList = append(userIDList, userID)
	}
	return
}
