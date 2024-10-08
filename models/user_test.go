package models_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserValidations(t *testing.T) {
	tests := map[string]struct {
		data     models.CreateUserData
		expected []error
	}{
		"should create a new user without failing validation": {
			data: models.CreateUserData{
				ID:              uuid.New(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Name:            "Jon Snow",
				Email:           "knowseverything@gmail.com",
				Password:        "yrgritte",
				ConfirmPassword: "yrgritte",
			},
			expected: nil,
		},
		"should return fail validation with errors:'ErrIsRequired, ErrIsRequired, ErrValueTooShort, ErrInvalidEmail, ErrPasswordDontMatchConfirm'": {
			data: models.CreateUserData{
				UpdatedAt:       time.Now(),
				Name:            "J",
				Email:           "knowseverythinggmail.com",
				Password:        "yrgritte",
				ConfirmPassword: "sansa",
			},
			expected: []error{
				validation.ErrIsRequired,
				validation.ErrIsRequired,
				validation.ErrValueTooShort,
				validation.ErrInvalidEmail,
				validation.ErrPasswordDontMatchConfirm,
			},
		},
		"should return fail validaiton with errors: 'ErrIsRequired, ErrIsRequired, ErrValueTooLong, ErrInvalidEmail, ErrPasswordDontMatchConfirm'": {
			data: models.CreateUserData{
				UpdatedAt:       time.Now(),
				Name:            "Jon Snoooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooow",
				Email:           "knowseverythinggmail.com",
				Password:        "yrgritte",
				ConfirmPassword: "sansa",
			},
			expected: []error{
				validation.ErrIsRequired,
				validation.ErrIsRequired,
				validation.ErrValueTooLong,
				validation.ErrInvalidEmail,
				validation.ErrPasswordDontMatchConfirm,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(
				test.data,
				models.CreateUserValidations(test.data.ConfirmPassword),
			)
			if test.expected == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expected,
						actualErrors,
					),
				)
			}

			if test.expected != nil {
				var valiErrors validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrors); !ok {
					t.Fail()
				}

				assert.Equal(
					t,
					test.expected,
					valiErrors.Unwrap(),
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expected,
						actualErrors,
					),
				)
			}
		})
	}
}

func TestUpdateUserValidations(t *testing.T) {
	tests := map[string]struct {
		data     models.UpdateUserData
		expected []error
	}{
		"should update a new user without failing validation": {
			data: models.UpdateUserData{
				ID:        uuid.New(),
				UpdatedAt: time.Now(),
				Name:      "Jon Snow",
				Email:     "knowseverything@gmail.com",
			},
			expected: nil,
		},
		"should return fail validation with errors:'ErrIsRequired'": {
			data: models.UpdateUserData{
				UpdatedAt: time.Now(),
				Name:      "King of the North",
				Email:     "theking@stark.com",
			},
			expected: []error{
				validation.ErrIsRequired,
			},
		},
		"should return fail validation with errors:'ErrIsRequired, ErrValueTooShort, ErrInvalidEmail'": {
			data: models.UpdateUserData{
				UpdatedAt: time.Now(),
				Name:      "J",
				Email:     "thekingstark.com",
			},
			expected: []error{
				validation.ErrIsRequired,
				validation.ErrValueTooShort,
				validation.ErrInvalidEmail,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(
				test.data,
				models.UpdateUserValidations(),
			)
			if test.expected == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expected,
						actualErrors,
					),
				)
			}

			if test.expected != nil {
				var valiErrors validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrors); !ok {
					t.Fail()
				}

				assert.Equal(
					t,
					test.expected,
					valiErrors.Unwrap(),
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expected,
						actualErrors,
					),
				)
			}
		})
	}
}
