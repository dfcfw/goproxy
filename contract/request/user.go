package request

type UserUpsert struct {
	JobNumber string `json:"job_number"`
	Name      string `json:"name"`
	Admin     bool   `json:"admin"`
}
