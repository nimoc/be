package friend

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type FriendTestCycle struct {
	ClearUserData func()
	ClearFriendUserData func()
	PrepareData func()
	ConcurrenceCheckData func()
}
func FriendTest(t *testing.T, f Friend, cycle FriendTestCycle) {
	ctx := context.Background()
	cycle.ClearUserData()
	cycle.ClearFriendUserData()
	cycle.PrepareData()
	// // f.Is(1,2)
	// {
	// 	isFriend, err := f.Is(ctx, 1,2) ; assert.NoError(t, err)
	// 	assert.Equal(t,isFriend, false)
	// }
	// // f.Is(2,1)
	// {
	// 	isFriend, err := f.Is(ctx, 2,1) ; assert.NoError(t, err)
	// 	assert.Equal(t,isFriend, false)
	// }
	// // f.Add(1,2)
	// // add (2,1)
	// {
	// 	assert.NoError(t, f.Add(ctx, 1,2))
	// 	assert.EqualError(t, f.Add(ctx, 2,1), "repeat")
	// }
	// // f.List(1)
	// {
	// 	userIDList, err := f.List(ctx, 1)  ; assert.NoError(t, err)
	// 	assert.Equal(t, userIDList, []int64{2})
	// }
	// // f.List(2)
	// {
	// 	userIDList, err := f.List(ctx, 2)  ; assert.NoError(t, err)
	// 	assert.Equal(t, userIDList, []int64{1})
	// }
	// // f.Is(1,2)
	// {
	// 	isFriend, err := f.Is(ctx, 1,2) ; assert.NoError(t, err)
	// 	assert.Equal(t,isFriend, true)
	// }
	// // f.Is(2,1)
	// {
	// 	isFriend, err := f.Is(ctx, 2,1) ; assert.NoError(t, err)
	// 	assert.Equal(t,isFriend, true)
	// }
	// // add (1,3)
	// {
	// 	assert.NoError(t, f.Add(ctx, 1,3))
	// }
	// // f.List(1)
	// {
	// 	userIDList, err := f.List(ctx, 1)  ; assert.NoError(t, err)
	// 	assert.Equal(t, userIDList, []int64{2, 3})
	// }
	// // f.List(3)
	// {
	// 	userIDList, err := f.List(ctx, 3)  ; assert.NoError(t, err)
	// 	assert.Equal(t, userIDList, []int64{1})
	// }
	// // f.Delete(1,2)
	// {
	// 	assert.NoError(t, f.Delete(ctx, 1,2))
	// 	assert.EqualError(t, f.Delete(ctx, 1,2), "not friends")
	// }
	// // f.Is(1,2)
	// {
	// 	isFriend, err := f.Is(ctx, 1,2) ; assert.NoError(t, err)
	// 	assert.Equal(t,isFriend, false)
	// }
	// // f.List(1)
	// {
	// 	userIDList, err := f.List(ctx, 1)  ; assert.NoError(t, err)
	// 	assert.Equal(t, userIDList, []int64{3})
	// }
	cycle.ClearFriendUserData()
	// Mutual
	assert.NoError(t, f.Add(ctx, 1,2))
	assert.NoError(t, f.Add(ctx, 1,3))
	assert.NoError(t, f.Add(ctx, 1,4))
	assert.NoError(t, f.Add(ctx, 2,3))
	assert.NoError(t, f.Add(ctx, 2,4))
	userIDList, err := f.Mutual(ctx, 1,2) ; assert.NoError(t, err)
	assert.Equal(t, userIDList, []int64{3,4})
	// 并发数据一致性测试
	{
		{
			isFriend, err := f.Is(ctx, 5, 6)
			assert.NoError(t, err)
			assert.Equal(t, isFriend, false)
		}
		wg := sync.WaitGroup{}
		for i:=0;i<20;i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := f.Add(ctx, 5,6)
				if err != nil {
					assert.EqualError(t, err, "repeat")
				}
			}()
		}
		wg.Wait()
		// 检查关系
		{
			isFriend, err := f.Is(ctx, 5, 6)
			assert.NoError(t, err)
			assert.Equal(t, isFriend, true)
			if cycle.ConcurrenceCheckData != nil {
				cycle.ConcurrenceCheckData()
			}
		}
		// 删除关系
		assert.NoError(t, f.Delete(ctx, 5,6))
		// 检查关系
		{
			isFriend, err := f.Is(ctx, 5, 6)
			assert.NoError(t, err)
			assert.Equal(t, isFriend, false)
		}
	}
}