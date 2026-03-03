package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// --- Mock repository ---

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// --- hashPassword / comparePassword ---

func TestHashPassword(t *testing.T) {
	hash, err := hashPassword("SecurePass1!")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	cost, err := bcrypt.Cost([]byte(hash))
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, cost, 12)
}

func TestComparePassword(t *testing.T) {
	hash, _ := hashPassword("SecurePass1!")
	assert.True(t, comparePassword(hash, "SecurePass1!"))
	assert.False(t, comparePassword(hash, "WrongPassword1!"))
}

// --- validateEmail ---

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@sub.domain.org", true},
		{"invalid-email", false},
		{"@nodomain.com", false},
		{"no@tld", false},
		{"", false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.valid, validateEmail(tc.email), "email: %s", tc.email)
	}
}

// --- validatePassword ---

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		password string
		valid    bool
	}{
		{"SecurePass1!", true},
		{"short1!", false},        // too short
		{"nouppercase1!", false},  // missing uppercase
		{"NODIGITSPECIAL!", false}, // missing digit
		{"NoSpecial1Char", false}, // missing special char
		{"NoDigit!", false},       // missing digit
		{"", false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.valid, validatePassword(tc.password), "password: %q", tc.password)
	}
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(nil, ErrUserNotFound)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	token, err := svc.Register(context.Background(), "user@example.com", "SecurePass1!")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	existing := &User{
		ID:        "user::existing",
		Email:     "user@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(existing, nil)

	_, err := svc.Register(context.Background(), "user@example.com", "SecurePass1!")
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
	repo.AssertExpectations(t)
}

func TestRegister_ValidationError_Email(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "not-an-email", "SecurePass1!")

	var validErr *ValidationError
	assert.True(t, errors.As(err, &validErr))
}

func TestRegister_ValidationError_Password(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "user@example.com", "weak")

	var validErr *ValidationError
	assert.True(t, errors.As(err, &validErr))
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	hash, _ := hashPassword("SecurePass1!")
	user := &User{
		ID:           "user::abc",
		Email:        "user@example.com",
		PasswordHash: hash,
	}
	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)

	token, err := svc.Login(context.Background(), "user@example.com", "SecurePass1!")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	hash, _ := hashPassword("SecurePass1!")
	user := &User{
		ID:           "user::abc",
		Email:        "user@example.com",
		PasswordHash: hash,
	}
	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)
	repo.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)

	_, err := svc.Login(context.Background(), "user@example.com", "WrongPass1!")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 1, user.FailedAttempts)
	repo.AssertExpectations(t)
}

func TestLogin_AccountLocked(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	locked := time.Now().Add(10 * time.Minute)
	user := &User{
		ID:             "user::abc",
		Email:          "user@example.com",
		FailedAttempts: maxFailedAttempts,
		LockedUntil:    &locked,
	}
	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)

	_, err := svc.Login(context.Background(), "user@example.com", "SecurePass1!")
	assert.ErrorIs(t, err, ErrAccountLocked)
	repo.AssertExpectations(t)
}

func TestLogin_LockoutTriggeredOnMaxAttempts(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	hash, _ := hashPassword("SecurePass1!")
	user := &User{
		ID:             "user::abc",
		Email:          "user@example.com",
		PasswordHash:   hash,
		FailedAttempts: maxFailedAttempts - 1,
	}
	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)
	repo.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)

	_, err := svc.Login(context.Background(), "user@example.com", "WrongPass1!")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, maxFailedAttempts, user.FailedAttempts)
	assert.NotNil(t, user.LockedUntil)
	assert.True(t, time.Now().Before(*user.LockedUntil))
	repo.AssertExpectations(t)
}

func TestLogin_SuccessResetsCounter(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	hash, _ := hashPassword("SecurePass1!")
	user := &User{
		ID:             "user::abc",
		Email:          "user@example.com",
		PasswordHash:   hash,
		FailedAttempts: 3,
	}
	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil)
	repo.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)

	token, err := svc.Login(context.Background(), "user@example.com", "SecurePass1!")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, 0, user.FailedAttempts)
	assert.Nil(t, user.LockedUntil)
	repo.AssertExpectations(t)
}

func TestLogin_UnknownEmail(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewService(repo, "test-secret")

	repo.On("FindByEmail", mock.Anything, "ghost@example.com").Return(nil, ErrUserNotFound)

	_, err := svc.Login(context.Background(), "ghost@example.com", "SecurePass1!")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	repo.AssertExpectations(t)
}
