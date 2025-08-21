package model

type User struct {
	JobNumber string `json:"job_number" gorm:"column:job_number;primaryKey;comment:工号"`
	Admin     bool   `json:"admin"      gorm:"column:admin;comment:是否管理员"`
}

func (User) TableName() string {
	return "user"
}
