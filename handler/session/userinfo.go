package session

type Userinfo struct {
	JobNumber string `json:"job_number"`     // 工号
	Admin     bool   `json:"admin,omitzero"` // 是否是管理员
}

func (u *Userinfo) ID() string {
	return u.JobNumber
}

var Key = sessionKey{}

type sessionKey struct{}

func (sessionKey) String() string {
	return "session-key"
}

func FromMap(hm map[string]any) *Userinfo {
	if hm == nil {
		return nil
	}
	skey := Key.String()
	val, exists := hm[skey]
	if !exists {
		return nil
	}
	if inf, ok := val.(*Userinfo); ok {
		return inf
	}

	return nil
}
