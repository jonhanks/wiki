package main

const (
	AnonymousUser = "Anonymous"
)

type UserInfo struct {
	username string
	roles    []string
}

type RequestInfo struct {
	Params map[string]string
	User   *UserInfo
}

func (u *UserInfo) Username() string {
	if u == nil {
		return AnonymousUser
	}
	if u.username == "" {
		return AnonymousUser
	}
	return u.username
}

func (u *UserInfo) String() string {
	return u.Username()
}

func (u *UserInfo) Roles() []string {
	if u.IsAnonymous() {
		return []string{}
	}
	return u.roles
}

func (u *UserInfo) IsAnonymous() bool {
	return u.Username() == AnonymousUser
}

func (u *UserInfo) AddRole(role string) {
	if u.IsAnonymous() {
		return
	}
	for _, existingRole := range u.Roles() {
		if existingRole == role {
			return
		}
	}
	u.roles = append(u.roles, role)
}
