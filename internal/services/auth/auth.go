package auth

import (
	"awesomeProject/internal/domain/model"
	"awesomeProject/internal/lib/jwt"
	"awesomeProject/internal/storage"
	"awesomeProject/internal/storage/sqlite"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTLL     time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExist          = errors.New("user exist")
)

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (model.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (model.App, error)
}

// New return a new instance of Auth service
func New(log *slog.Logger,
	storage *sqlite.Storage,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    storage,
		appProvider:  storage,
		userProvider: storage,
		tokenTLL:     tokenTTL,
	}
}

// Login checks if user with credentials exists in the system
//
// if user exists but pswd is incorrect, returns error
// if not exists, returns error
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int64,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("operation", op),
		slog.String("email", email), //To log user data may be bad idea for some markets. Needs to be careful
	)

	log.Info("attempt to login")
	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid password", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully logged in")

	token, err := jwt.NewToken(user, app, a.tokenTLL)
	if err != nil {
		a.log.Error("failed to create token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// Register register new user and returns userID
// if user exists returns error
func (a *Auth) Register(ctx context.Context, email, password string) (int64, error) {
	const op = "auth.registerUser"

	log := a.log.With(
		slog.String("operation", op),
		slog.String("email", email), //To log user data may be bad idea for some markets. Needs to be careful
	)

	log.Info("register user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			log.Warn("user already exists")

			return 0, fmt.Errorf("%s: %w", op, ErrUserExist)
		}

		log.Error("failed to save user")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("register user")

	return id, nil
}

// IsAdmin checks if user admin
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("operation", op),
		slog.Int64("user_id", userID), //To log user data may be bad idea for some markets. Needs to be careful
	)

	log.Info("Check if user admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user not admin")

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Successfully check if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
