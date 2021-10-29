package friend

import (
	"context"
	red "github.com/goclub/redis"
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnionFriends(t *testing.T) {
	f, err := NewUnionFriend() ; assert.NoError(t, err)
	ctx := context.Background()
	FriendTest(t, f, FriendTestCycle{
		ClearUserData: func() {
			// 清空用户
			_, err = f.db.ClearTestData(ctx, sq.QB{
				Raw: sq.Raw{
					Query: "DELETE FROM `user` WHERE `id` < 10",
				},
			}) ; assert.NoError(t, err)
		},
		ClearFriendUserData: func() {
			// 清空关系
			for i:=0;i<10 ;i++ {
				_, err = red.DEL{
					Keys: []string{
						f.KeyFriendSets(int64(i)),
						f.KeyNoFriendString(int64(i)),
					},
				}.Do(ctx, f.redis)
				assert.NoError(t, err)
			}
			// 清空关系
			_, err = f.db.ClearTestData(ctx, sq.QB{
				Raw: sq.Raw{
					Query: "DELETE FROM `user_friend` WHERE `user_id` < 10",
				},
			}) ; assert.NoError(t, err)
		},
		PrepareData: func() {
			_, err = f.db.Insert(ctx, sq.QB{
				Table: sq.Table("user", nil, nil),
				InsertMultiple: sq.InsertMultiple{
					Column: []sq.Column{"id", "name"},
					Values: [][]interface{}{
						{1, "a"},
						{2, "b"},
						{3, "c"},
						{4, "d"},
						{5, "e"},
						{6, "f"},
					},
				},
			}) ; assert.NoError(t, err)
		},
	})
}