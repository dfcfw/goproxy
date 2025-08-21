package request

import "time"

type AccessTokenCreate struct {
	Name      string    `json:"name"`
	ExpiredAt time.Time `json:"expired_at"`
}
