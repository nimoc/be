package friend

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRedisFriends(t *testing.T) {
	f, err := NewRedisFriend() ; assert.NoError(t, err)
	ctx := context.Background()
	FriendTest(t, f, FriendTestCycle{
		ClearUserData: func() {

		},
		ClearFriendUserData: func() {
			// 清空关系
			for i:=0;i<10 ;i++ {
				_, _, err := f.redis.DoIntegerReply(ctx, []string{"DEL", f.KeyFriendSets(int64(i))})
				assert.NoError(t, err)
			}
		},
		PrepareData: func() {

		},
	})
}