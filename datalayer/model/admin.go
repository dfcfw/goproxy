package model

type Admin struct {
	JobNumber string `json:"job_number" gorm:"column:job_number;primaryKey;comment:工号"`
}

func (Admin) TableName() string {
	return "admin"
}
