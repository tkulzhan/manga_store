package services

type AuthService struct {

}

func NewAuthService() AuthService {
	return AuthService{}
}

func (s AuthService) Register() error {
	return nil
}

func (s AuthService) Login() error {
	return nil
}

func (s AuthService) Logout() error {
	return nil
}