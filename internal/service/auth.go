package service

type StaticAuth struct {
	users map[string]string
}

func NewStaticAuth(users map[string]string) *StaticAuth {
	return &StaticAuth{users: users}
}

func (a *StaticAuth) Authenticate(user, pass string) bool {
	expected, ok := a.users[user]
	return ok && expected == pass
}
