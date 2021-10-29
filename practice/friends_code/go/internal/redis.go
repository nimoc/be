package friend

import (
	"context"
	redisv8 "github.com/go-redis/redis/v8"
	"github.com/goclub/redis"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)
type RedisFriend struct {
	// 为了便于理解,暂时不做数据层和逻辑层的划分
	redis red.Connecter
}
// 为了便于理解, redis 版不处理错误的 userID 和 friendUserID
type RedisConfig struct {
	Redis struct{
		Addr string `yaml:"addr"`
	} `yaml:"redis"`
}
func NewRedisFriend () (*RedisFriend, error) {
	config := RedisConfig{}

	data, err := ioutil.ReadFile(path.Join(os.Getenv("GOPATH"), "src/github.com/nimoc/backend/practice/friends/go/env.yaml")) ; if err != nil {
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
	return &RedisFriend{
		redis: redis,
	}, nil
}

func (dep RedisFriend) KeyFriendSets(userID int64) string {
	return "friend:" + strconv.FormatInt(userID, 10)
}
func (dep RedisFriend) Add(ctx context.Context, userID int64, friendUserID int64) (err error) {
	keyUserID := dep.KeyFriendSets(userID)
	keyFriendUserID := dep.KeyFriendSets(friendUserID)
	reply, isNil, err := dep.redis.Eval(ctx, red.Script{
		Keys: []string{
			/* 1 */ keyUserID,
			/* 2 */ keyFriendUserID,
		},
		Argv: []string{
			/* 1 */ strconv.FormatInt(
				userID, 10),
			/* 2 */ strconv.FormatInt(
				friendUserID, 10),
		},
		Script:`
		local keyUserID = KEYS[1]
		local keyFriendUserID = KEYS[2]
		local userID = ARGV[1]
		local friendUserID = ARGV[2]
		local reply = {}
		reply[1] = redis.call("SADD", keyUserID, friendUserID)
		reply[2] = redis.call("SADD", keyFriendUserID, userID)
		return reply
		`,
	}) ; if err != nil {
	    return
	}
	if isNil {
		return errors.New("can not be nil")
	}
	arrayIntReply := red.ParseArrayIntegerReply(reply)
	for _, intReply := range arrayIntReply {
		if intReply.Valid && intReply.Int64 == 0 {
			return errors.New("repeat")
		}
	}
	return
}
func (dep RedisFriend) Is(ctx context.Context, userID int64, friendUserID int64) (isFriend bool, err error) {
	cmd := []string{"SISMEMBER", dep.KeyFriendSets(userID), strconv.FormatInt(friendUserID, 10) }
	reply, _, err := dep.redis.DoIntegerReply(ctx, cmd) ; if err != nil {
	    return
	}
	return reply == 1, nil

}
func (dep RedisFriend) List (ctx context.Context, userID int64) (userIDList []int64, err error) {
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
func (dep RedisFriend) Delete  (ctx context.Context, userID int64, friendUserID int64) (err error){
	keyUserID := dep.KeyFriendSets(userID)
	keyFriendUserID := dep.KeyFriendSets(friendUserID)
	reply, isNil, err := dep.redis.Eval(ctx, red.Script{
		Keys: []string{
			/* 1 */ keyUserID,
			/* 2 */ keyFriendUserID,
		},
		Argv: []string{
			/* 1 */ strconv.FormatInt(
				userID, 10),
			/* 2 */ strconv.FormatInt(
				friendUserID, 10),
		},
		Script:`
		local keyUserID = KEYS[1]
		local keyFriendUserID = KEYS[2]
		local userID = ARGV[1]
		local friendUserID = ARGV[2]
		local reply = {}
		reply[1] = redis.call("SREM", keyUserID, friendUserID)
		reply[2] = redis.call("SREM", keyFriendUserID, userID)
		return reply
		`,
	}) ; if err != nil {
		return
	}
	if isNil {
		return errors.New("can not be nil")
	}
	arrayIntReply := red.ParseArrayIntegerReply(reply)
	for _, intReply := range arrayIntReply {
		if intReply.Valid && intReply.Int64 == 0 {
			return errors.New("not friends")
		}
	}
	return
}
func (dep RedisFriend) Mutual  (ctx context.Context, userID int64, friendUserID int64) (userIDList []int64, err error){
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
