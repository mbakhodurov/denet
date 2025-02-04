package models

import "time"

type User struct {
	Id          int64     `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Points      int64     `json:"points"`
	Referral_id int64     `json:"referral_id"`
	Created_at  time.Time `json:"created_at"`
}
