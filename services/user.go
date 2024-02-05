package services

import (
	"context"
	"log"

	"time"

	"github.com/go-playground/validator/v10"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

type userDatabase interface {
	InsertUser(ctx context.Context, arg database.InsertUserParams) (database.User, error)
	DoesMailExists(ctx context.Context, mail string) (bool, error)
	QueryUserByMail(ctx context.Context, mail string) (database.User, error)
	QueryUser(ctx context.Context, id uuid.UUID) (database.User, error)
	UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
}

type newUserValidation struct {
	ConfirmPassword string `validate:"required,gte=8"`
	Name            string `validate:"required,gte=2"`
	Mail            string `validate:"required,email"`
	MailRegistered  bool   `validate:"required,ne=true"` // TODO: why does this fail with 'required'?
	Password        string `validate:"required,gte=8"`
}

func passwordMatchValidation(sl validator.StructLevel) {
	data := sl.Current().Interface().(newUserValidation)

	if data.ConfirmPassword != data.Password {
		sl.ReportError(data.ConfirmPassword, "", "ConfirmPassword", "", "confirm password must match password")
	}
}

func NewUser(
	ctx context.Context, data entity.NewUser, db userDatabase, v *validator.Validate, passwordPepper string) (entity.User, error) {
	mailAlreadyRegistered, err := db.DoesMailExists(ctx, data.Mail)
	if err != nil {
		telemetry.Logger.Error("could not check if email exists", "error", err)
		return entity.User{}, err
	}
	log.Print("############################")
	log.Print(mailAlreadyRegistered)

	v.RegisterStructValidation(passwordMatchValidation, newUserValidation{})

	newUserData := newUserValidation{
		ConfirmPassword: data.ConfirmPassword,
		Name:            data.Name,
		Mail:            data.Mail,
		MailRegistered:  mailAlreadyRegistered,
		Password:        data.Password,
	}

	if err := v.Struct(newUserData); err != nil {
		return entity.User{}, err
	}

	hashedPassword, err := hashAndPepperPassword(newUserData.Password, passwordPepper)
	if err != nil {
		telemetry.Logger.Error("error hashing and peppering password", "error", err)
		return entity.User{}, err
	}

	user, err := db.InsertUser(ctx, database.InsertUserParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		UpdatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Name:      newUserData.Name,
		Mail:      newUserData.Mail,
		Password:  hashedPassword,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return entity.User{}, err
	}

	return entity.User{
		ID:        user.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(user.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(user.UpdatedAt),
		Name:      user.Name,
		Mail:      user.Mail,
	}, nil
}

type updateUserValidation struct {
	ConfirmPassword string `validate:"required,gte=8"`
	Password        string `validate:"required,gte=8"`
	Name            string `validate:"required,gte=2"`
	Mail            string `validate:"required,email"`
}

func resetPasswordMatchValidation(sl validator.StructLevel) {
	data := sl.Current().Interface().(updateUserValidation)

	if data.ConfirmPassword != data.Password {
		sl.ReportError(data.ConfirmPassword, "", "ConfirmPassword", "", "confirm password must match password")
	}
}

func UpdateUser(
	ctx context.Context, data entity.UpdateUser, db userDatabase, v *validator.Validate, passwordPepper string) (entity.User, error) {

	v.RegisterStructValidation(resetPasswordMatchValidation, updateUserValidation{})

	validatedData := updateUserValidation{
		ConfirmPassword: data.ConfirmPassword,
		Password:        data.Password,
		Name:            data.Name,
		Mail:            data.Mail,
	}

	if err := v.Struct(validatedData); err != nil {
		return entity.User{}, err
	}

	hashedPassword, err := hashAndPepperPassword(validatedData.Password, passwordPepper)
	if err != nil {
		telemetry.Logger.Error("error hashing and peppering password", "error", err)
		return entity.User{}, err
	}
	telemetry.Logger.Info("this is id", "id", data.ID)

	updatedUser, err := db.UpdateUser(ctx, database.UpdateUserParams{
		UpdatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Name:      data.Name,
		Mail:      data.Mail,
		Password:  hashedPassword,
		ID:        data.ID,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return entity.User{}, err
	}

	return entity.User{
		ID:        updatedUser.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(updatedUser.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(updatedUser.UpdatedAt),
		Name:      updatedUser.Name,
		Mail:      updatedUser.Mail,
	}, nil
}
