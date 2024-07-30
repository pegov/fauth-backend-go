package model

import "time"

type MassLogoutStatus struct {
	Date   *time.Time `json:"date"`
	Active bool       `json:"active"`
}
