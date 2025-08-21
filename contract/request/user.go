package request

type UserUpsert struct {
	JobNumber string `json:"job_number"`
	Admin     bool   `json:"admin"`
}
