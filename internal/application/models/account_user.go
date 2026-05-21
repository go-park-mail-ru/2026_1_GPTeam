package models

//go:generate easyjson -all models.go
import "time"

type AccountUserModel struct {
	Id        int       `json:"id"`
	AccountId int       `json:"account_id"`
	UserId    int       `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type AccountUserCreateModel struct {
	AccountId int
	UserId    int
	Status    string
}

type AccountUserUpdateModel struct {
	Status *string
}

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRejected InviteStatus = "rejected"
)

type UserSearchResult struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type InviteRequest struct {
	Query string `json:"query"`
}

type InviteResponse struct {
	Users []UserSearchResult
}

type MemberResponse struct {
	Id        int       `json:"id"`
	AccountId int       `json:"account_id"`
	UserId    int       `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	IsOwner   bool      `json:"is_owner"`
}

// PendingInviteView — приглашение с названием счёта для списка «ожидающих».
type PendingInviteView struct {
	Id          int       `json:"id"`
	AccountId   int       `json:"account_id"`
	AccountName string    `json:"account_name"`
	UserId      int       `json:"user_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
