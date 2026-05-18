package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		check    func(t *testing.T, hashedPassword, password string)
	}{
		{
			name:     "OK",
			password: "secret_password_123",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
		{
			name:     "DifferentHashesForSamePassword",
			password: "same_password",
			check: func(t *testing.T, hashedPassword1, password string) {
				hashedPassword2, err := HashPassword(password)
				require.NoError(t, err)
				require.NotEqual(t, hashedPassword1, hashedPassword2)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword1), []byte(password)))
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword2), []byte(password)))
			},
		},
		{
			name:     "ShortPassword",
			password: "a",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
		{
			name:     "LongPassword",
			password: "this_is_a_very_long_password_that_contains_many_characters_and_symbols",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
		{
			name:     "PasswordWithSpecialCharacters",
			password: "P@ssw0rd!#$%&*()_+-=[]{}|;:,.<>?",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
		{
			name:     "PasswordWithUnicodeCharacters",
			password: "密码123456",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
		{
			name:     "PasswordWithSpaces",
			password: "password with spaces",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
		{
			name:     "EmptyPassword",
			password: "",
			check: func(t *testing.T, hashedPassword, password string) {
				require.NotEmpty(t, hashedPassword)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tc.password)
			require.NoError(t, err)
			tc.check(t, hashedPassword, tc.password)
		})
	}
}

func TestCheckPassword(t *testing.T) {
	testCases := []struct {
		name      string
		password  string
		testFunc  func(t *testing.T, password string)
	}{
		{
			name:     "CorrectPassword",
			password: "correct_password",
			testFunc: func(t *testing.T, password string) {
				hashedPwd, err := HashPassword(password)
				require.NoError(t, err)

				err = CheckPassword(password, hashedPwd)
				require.NoError(t, err)
			},
		},
		{
			name:     "IncorrectPassword",
			password: "correct_password",
			testFunc: func(t *testing.T, password string) {
				hashedPwd, err := HashPassword(password)
				require.NoError(t, err)

				err = CheckPassword("wrong_password", hashedPwd)
				require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
			},
		},
		{
			name:     "CaseSensitive",
			password: "MyPassword123",
			testFunc: func(t *testing.T, password string) {
				hashedPwd, err := HashPassword(password)
				require.NoError(t, err)

				err = CheckPassword("mypassword123", hashedPwd)
				require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
			},
		},
		{
			name:     "EmptyPasswordCheck",
			password: "",
			testFunc: func(t *testing.T, password string) {
				hashedPwd, err := HashPassword(password)
				require.NoError(t, err)

				err = CheckPassword("", hashedPwd)
				require.NoError(t, err)
			},
		},
		{
			name:     "CheckAgainstWrongHashedPassword",
			password: "password",
			testFunc: func(t *testing.T, password string) {
				hashedPwd2, err := HashPassword("different_password")
				require.NoError(t, err)

				err = CheckPassword(password, hashedPwd2)
				require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc(t, tc.password)
		})
	}
}

func TestPasswordHashingConsistency(t *testing.T) {
	password := "test_password_123"

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	// Different hashes for same password (bcrypt generates random salt)
	require.NotEqual(t, hash1, hash2)

	// But both hashes should verify against the same password
	require.NoError(t, CheckPassword(password, hash1))
	require.NoError(t, CheckPassword(password, hash2))
}

func TestPasswordLength(t *testing.T) {
	password := "test_password"

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	// Bcrypt hashes have a consistent length (60 characters for bcrypt)
	require.Equal(t, 60, len(hashedPassword))
}

func TestCheckPasswordWithInvalidHash(t *testing.T) {
	password := "test_password"
	invalidHash := "not_a_valid_bcrypt_hash"

	err := CheckPassword(password, invalidHash)
	require.Error(t, err)
	require.NotEqual(t, bcrypt.ErrMismatchedHashAndPassword.Error(), err.Error())
}

func TestPasswordSecurityProperties(t *testing.T) {
	password := "MySecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Hash should not contain the original password
	require.NotContains(t, hash, password)

	// Hash should be a string (bcrypt format starts with $2a$, $2b$, or $2y$)
	require.True(t, len(hash) >= 13)
	require.Contains(t, []string{"$2a$", "$2b$", "$2y$"}, hash[:4])

	// Multiple calls should produce different hashes
	hash2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEqual(t, hash, hash2)

	// Both hashes should validate
	require.NoError(t, CheckPassword(password, hash))
	require.NoError(t, CheckPassword(password, hash2))
}

func TestCheckPasswordWithEmptyHashedPassword(t *testing.T) {
	password := "test_password"
	emptyHash := ""

	err := CheckPassword(password, emptyHash)
	require.Error(t, err)
}
