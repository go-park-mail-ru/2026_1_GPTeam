package repository

type AccountUserStatus string

const (
	AccountUserStatusPending  AccountUserStatus = "pending"
	AccountUserStatusAccepted AccountUserStatus = "accepted"
	AccountUserStatusDeclined AccountUserStatus = "declined"
)

type AccountUserDeletedReason string

const (
	AccountUserDeletedReasonKicked AccountUserDeletedReason = "kicked"
	AccountUserDeletedReasonLeft   AccountUserDeletedReason = "left"
)
