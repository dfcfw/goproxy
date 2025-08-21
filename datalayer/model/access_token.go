package model

import "time"

type AccessToken struct {
	ID        int64     `json:"id,string,omitzero"  gorm:"column:id;primaryKey;autoIncrement;comment:ID"`
	Name      string    `json:"name"                gorm:"column:name;size:20;not null;uniqueIndex:uk_job_number_name;comment:名字"`
	JobNumber string    `json:"job_number"          gorm:"column:job_number;size:10;not null;uniqueIndex:uk_job_number_name;comment:工号"`
	Token     string    `json:"token,omitzero"      gorm:"column:token;size:100;not null;unique;comment:Token"`
	ExpiredAt time.Time `json:"expired_at,omitzero" gorm:"column:expired_at;comment:过期时间"`
}

func (AccessToken) TableName() string {
	return "access_token"
}
