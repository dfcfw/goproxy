package request

type Named struct {
	Name string `json:"name" query:"name"`
}

type JobNumber struct {
	JobNumber string `json:"job_number" query:"job_number"`
}
