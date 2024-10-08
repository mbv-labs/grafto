package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/mbvlabs/grafto/pkg/validation"
)

type User struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Email           string
	EmailVerifiedAt time.Time
}

func (u User) IsVerified() bool {
	return !u.EmailVerifiedAt.IsZero()
}

type CreateUserData struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Email           string
	Password        string
	ConfirmPassword string
}

var CreateUserValidations = func(confirm string) map[string][]validation.Rule {
	return map[string][]validation.Rule{
		"ID":        {validation.RequiredRule},
		"CreatedAt": {validation.RequiredRule},
		"Name": {
			validation.RequiredRule,
			validation.MinLengthRule(2),
			validation.MaxLengthRule(25),
		},
		"Email": {validation.RequiredRule, validation.ValidEmailRule},
		"Password": {
			validation.RequiredRule,
			validation.MinLengthRule(6),
			validation.PasswordMatchConfirmRule(confirm),
		},
	}
}

type UpdateUserData struct {
	ID        uuid.UUID
	UpdatedAt time.Time
	Name      string
	Email     string
}

var UpdateUserValidations = func() map[string][]validation.Rule {
	return map[string][]validation.Rule{
		"ID":        {validation.RequiredRule},
		"UpdatedAt": {validation.RequiredRule},
		"Name": {
			validation.RequiredRule,
			validation.MinLengthRule(2),
			validation.MaxLengthRule(25),
		},
		"Email": {validation.RequiredRule, validation.ValidEmailRule},
	}
}

type ChangeUserPasswordData struct {
	ID              uuid.UUID
	UpdatedAt       time.Time
	Password        string
	ConfirmPassword string
}

var ChangeUserPasswordValidations = func(confirm string) map[string][]validation.Rule {
	return map[string][]validation.Rule{
		"ID":        {validation.RequiredRule},
		"UpdatedAt": {validation.RequiredRule},
		"Password": {
			validation.RequiredRule,
			validation.MinLengthRule(6),
			validation.PasswordMatchConfirmRule(confirm),
		},
	}
}
