package friend

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	sq "github.com/goclub/sql"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewMysqlFriends(t *testing.T) {

	f, err := NewMysqlFriend() ; assert.NoError(t, err)
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
		ConcurrenceCheckData: func() {
			// 并发检查数据库
			firstUserID, secondUserID := f.SortUserID(5, 6)
			count, err := f.db.Count(ctx, sq.QB{
				Raw: sq.Raw{
					Query: "SELECT count(*) FROM `user_friend` WHERE `user_id` = ? AND `friend_user_id` = ?",
					Values: []interface{}{
						firstUserID, secondUserID,
					},
				},
			}) ; assert.NoError(t, err)
			assert.Equal(t, count, uint64(1))
		},
	})
	// 并发数据一致性测试
	{
		{
			isFriend, err := f.Is(ctx, 5, 6)
			assert.NoError(t, err)
			assert.Equal(t, isFriend, false)
		}
		wg := sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := f.Add(ctx, 5, 6)
				if err != nil {
					assert.EqualError(t, err, "repeat")
				}
			}()
		}
		wg.Wait()
	}
}