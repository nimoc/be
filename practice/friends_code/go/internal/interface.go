package friend

import "context"

type Friend interface {
	Add		(ctx context.Context, userID int64, friendUserID int64) (error error)
	List    (ctx context.Context, userID int64) (userIDList []int64, err error)
	Is      (ctx context.Context, userID int64, friendUserID int64) (isFriend bool, err error)
	Delete  (ctx context.Context, userID int64, friendUserID int64) (err error)
	Mutual  (ctx context.Context, userID int64, friendUserID int64) (userIDList []int64, err error)
}