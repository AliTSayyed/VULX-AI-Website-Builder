package services

/*
account service will combine the oauth flow, jwt, and creating / logging back in a user.
If there is an existing email in the db, do not create a new user with the same email, give an error with which provider has used that email
This service is what is directly used by the account handler.
JWT will handle user login, not storing oauth access tokens
*/

type AccountService struct {
	oauthService OauthService
	authService  AuthService
	userService  UserService
}

func NewAccountService(oauthService *OauthService, authService *AuthService, userService *UserService) *AccountService {
	return &AccountService{
		oauthService: *oauthService,
		authService:  *authService,
		userService:  *userService,
	}
}
