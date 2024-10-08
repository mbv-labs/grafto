package models

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mbvlabs/grafto/pkg/validation"
)

type userStorage interface {
	InsertUser(ctx context.Context, arg User, hashedPassword string) (User, error)
	QueryUserByEmail(ctx context.Context, mail string) (User, error)
	QueryUserByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdateUser(ctx context.Context, arg User) (User, error)
	UpdateUserPassword(
		ctx context.Context,
		userID uuid.UUID,
		newPassword string,
		updatedAt time.Time,
	) error
	VerifyUserEmail(
		ctx context.Context,
		updatedAt time.Time,
		email string,
	) error
}

type authService interface {
	HashAndPepperPassword(password string) (string, error)
}

type UserService struct {
	storage userStorage
	authSvc authService
}

func NewUserService(storage userStorage, authSvc authService) UserService {
	return UserService{storage, authSvc}
}

func (us UserService) ByEmail(ctx context.Context, email string) (User, error) {
	user, err := us.storage.QueryUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (us UserService) New(
	ctx context.Context,
	data CreateUserData,
) (User, error) {
	if err := validation.ValidateStruct(data, CreateUserValidations(data.ConfirmPassword)); err != nil {
		return User{}, errors.Join(ErrFailValidation, err)
	}

	_, err := us.storage.QueryUserByEmail(ctx, data.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.ErrorContext(ctx, "could not query user by email", "error", err)
		return User{}, err
	}
	if err == nil {
		return User{}, ErrUserAlreadyExists
	}

	hashedPassword, err := us.authSvc.HashAndPepperPassword(data.Password)
	if err != nil {
		slog.ErrorContext(ctx, "error hashing and peppering password", "error", err)
		return User{}, err
	}

	newUser, err := us.storage.InsertUser(ctx, User{
		ID:        uuid.New(),
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		Name:      data.Name,
		Email:     data.Email,
	}, hashedPassword)
	if err != nil {
		slog.ErrorContext(ctx, "could not insert user", "error", err)
		return User{}, err
	}

	return newUser, nil
}

func (us UserService) Update(
	ctx context.Context,
	data UpdateUserData,
) (User, error) {
	if err := validation.ValidateStruct(data, UpdateUserValidations()); err != nil {
		return User{}, errors.Join(ErrFailValidation, err)
	}

	updatedUser, err := us.storage.UpdateUser(ctx, User{
		ID:        data.ID,
		UpdatedAt: data.UpdatedAt,
		Name:      data.Name,
		Email:     data.Email,
	})
	if err != nil {
		slog.ErrorContext(ctx, "could not insert user", "error", err)
		return User{}, err
	}

	return updatedUser, nil
}

func (us UserService) ChangePassword(ctx context.Context, data ChangeUserPasswordData) error {
	if err := validation.ValidateStruct(data, UpdateUserValidations()); err != nil {
		return errors.Join(ErrFailValidation, err)
	}

	hashedPassword, err := us.authSvc.HashAndPepperPassword(data.Password)
	if err != nil {
		return err
	}

	return us.storage.UpdateUserPassword(ctx, data.ID, hashedPassword, data.UpdatedAt)
}

func (us UserService) VerifyEmail(ctx context.Context, email string) error {
	err := us.storage.VerifyUserEmail(ctx, time.Now(), email)
	if err != nil {
		slog.ErrorContext(ctx, "could not verify user email", "error", err)
		return err
	}

	return nil
}
