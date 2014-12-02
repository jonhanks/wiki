package main

const (
	AnonymousUser = "Anonymous"
)

// Hold basic user information
type UserInfo struct {
	username string   // The username
	roles    []string // Basic roles/groups that user has
}

// This is a request context of sorts, bring it all to one place so that
// it bacn be accessed in each handler
type RequestInfo struct {
	Params map[string]string
	User   *UserInfo
	DB     DB
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

// Is this an anoymous user
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
