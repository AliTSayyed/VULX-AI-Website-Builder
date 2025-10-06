package domain

/*
* A define what a user looks like in our business logic
* validates data, ensures business rules are follwed, ensures proper behavior
* only if incoming data satisfies the domain level requirements can the data be
* sent to the database.
* expose nothing in the domain.User struct, use methods to get values
 */

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrUserFirstNameEmpty = NewError(ErrorTypeInvalid, errors.New("user first name cannot be empty"))
	ErrUserLastNameEmpty  = NewError(ErrorTypeInvalid, errors.New("user last name cannot be empty"))
	ErrUserEmailEmpty     = NewError(ErrorTypeInvalid, errors.New("user email cannot be empty"))
)

type User struct {
	id        uuid.UUID
	firstName string
	lastName  string
	email     string
	credits   int
	isActive  bool
}

type UserFromProvider struct {
	userID     uuid.UUID
	provider   string
	providerID string
}

func NewUser(firstName, lastName, email string) (*User, error) {
	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		return nil, ErrUserFirstNameEmpty
	}

	if lastName == "" {
		return nil, ErrUserLastNameEmpty
	}

	if email == "" {
		return nil, ErrUserEmailEmpty
	}

	return &User{
		id:        uuid.New(),
		firstName: firstName,
		lastName:  lastName,
		email:     email,
	}, nil
}

func RestoreUser(id uuid.UUID, firstName string, lastName string, email string, credits int, is_active bool) *User {
	return &User{
		id:        id,
		firstName: firstName,
		lastName:  lastName,
		email:     email,
		credits:   credits,
		isActive:  is_active,
	}
}

func RestoreUserFromProvider(userID uuid.UUID, providerName LoginProvider, providerID string) {
	return &UserFromProvider{
		userID:     userID,
		provider:   providerName,
		providerID: providerID,
	}
}

func (u *User) ID() uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return u.id
}

func (u *User) FirstName() string {
	if u == nil {
		return ""
	}
	return u.firstName
}

func (u *User) LastName() string {
	if u == nil {
		return ""
	}
	return u.lastName
}

func (u *User) Email() string {
	if u == nil {
		return ""
	}
	return u.email
}

func (u *User) Credits() int {
	if u == nil {
		return 0
	}
	return u.credits
}

func (u *User) IsActive() bool {
	return u.isActive
}
