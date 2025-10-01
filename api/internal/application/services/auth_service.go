package services

/*
this service creates auth tokens for users to manage sessions.
a user will have an access tokent that will last 7 days
if token expires in < 42 hours, automaitcally create new accesstoken and override old one
if user logs out, store access token in the black list (db)
if user does not interact with app for > 7 days, access token expires and user must login again
*/

type AuthAdapter interface {
	CreateJWT()
	ValidateJWT()
	ParseJWT()
}

// since only using access tokens, just black list tokens in cahce if user logs out
type AuthService struct {
	cache       Cache
	authAdapter AuthAdapter
}

func NewAuthService(cache Cache, authAdapter AuthAdapter) *AuthService {
	return &AuthService{
		cache:       cache,
		authAdapter: authAdapter,
	}
}

func (a *AuthService) CreateSession() {
}

func (a *AuthService) ValidateSession() {
}

func (a *AuthService) Logout() {
}
