/*
* A domain defines what a user looks like in our business logic
* validates data, ensures business rules are follwed, ensures proper behavior
* only if incoming data satisfies the domain level requirements can the data be
* sent to the database.
* expose nothing in the domain.User struct, use methods to get values
 */
package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrUserNameEmpty = NewError(ErrorTypeInvalid, errors.New("user name cannot be empty"))

type User struct {
	id   uuid.UUID
	name string
}

func NewUser(name string) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrUserNameEmpty
	}

	return &User{
		id:   uuid.New(),
		name: name,
	}, nil
}

func RestoreUser(id uuid.UUID, name string) *User {
	return &User{
		id:   id,
		name: name,
	}
}

func (u *User) ID() uuid.UUID {
	if u == nil {
		return uuid.Nil
	}
	return u.id
}

func (u *User) Name() string {
	if u == nil {
		return ""
	}
	return u.name
}
