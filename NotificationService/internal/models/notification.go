package models

type Notification struct {
	Id string			`json:"id"`
	Message string		`json:"message"`
	NotificationTime string `json:"timestamp"`
	Delivered bool		`json:"delivered"`
}