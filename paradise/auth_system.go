package paradise

type AuthSystem interface {
	CheckUser(user, pass string) bool
}

type AuthManager struct {
	AuthSystem
}

type DefaultAuthSystem struct {
}

func (das DefaultAuthSystem) CheckUser(user, pass string) bool {
	return true
}

func NewDefaultAuthSystem() *AuthManager {
	am := AuthManager{}
	am.AuthSystem = DefaultAuthSystem{}
	return &am
}
