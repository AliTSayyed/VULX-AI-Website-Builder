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

type User struct {
	id   uuid.UUID
	name string
}

func NewUser(name string) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name can not be blank")
	}

	return &User{
		id:   uuid.New(),
		name: name,
	}, nil
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
