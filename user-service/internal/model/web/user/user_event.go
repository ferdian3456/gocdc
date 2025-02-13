package user

import "time"

type AuditEvent struct {
	Id         string     `json:"id"`
	Event      string     `json:"event"`
	Created_at *time.Time `json:"created_at"`
}

type NotificationEvent struct {
	Id         string     `json:"id"`
	Email      string     `json:"email"`
	Event      string     `json:"event"`
	Created_at *time.Time `json:"created_at"`
}

type VerificationEvent struct {
	Id              string `json:"id"`
	Profile_picture string `json:"profile_picture"`
}
