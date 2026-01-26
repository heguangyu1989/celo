package utils

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePassword_Default(t *testing.T) {
	opts := DefaultPasswordOptions()
	password, err := GeneratePassword(opts)

	assert.NoError(t, err)
	assert.Len(t, password, 16)
	assert.Contains(t, password, password)
}

func TestGeneratePassword_Length(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"length 8", 8},
		{"length 12", 12},
		{"length 20", 20},
		{"length 32", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultPasswordOptions()
			opts.Length = tt.length
			password, err := GeneratePassword(opts)

			assert.NoError(t, err)
			assert.Len(t, password, tt.length)
		})
	}
}

func TestGeneratePassword_CharacterSets(t *testing.T) {
	tests := []struct {
		name      string
		useUpper  bool
		useLower  bool
		useDigits bool
		validator func(rune) bool
	}{
		{
			"uppercase only",
			true, false, false,
			unicode.IsUpper,
		},
		{
			"lowercase only",
			false, true, false,
			unicode.IsLower,
		},
		{
			"digits only",
			false, false, true,
			unicode.IsDigit,
		},
		{
			"uppercase and lowercase",
			true, true, false,
			func(r rune) bool { return unicode.IsUpper(r) || unicode.IsLower(r) },
		},
		{
			"all types",
			true, true, true,
			func(r rune) bool { return unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsDigit(r) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := PasswordOptions{
				Length:    20,
				UseUpper:  tt.useUpper,
				UseLower:  tt.useLower,
				UseDigits: tt.useDigits,
			}
			password, err := GeneratePassword(opts)

			assert.NoError(t, err)
			for _, char := range password {
				assert.True(t, tt.validator(char), "Character %c does not match expected set", char)
			}
		})
	}
}

func TestGeneratePassword_CustomChars(t *testing.T) {
	customChars := "abc123"
	opts := PasswordOptions{
		Length:      10,
		CustomChars: customChars,
	}
	password, err := GeneratePassword(opts)

	assert.NoError(t, err)
	assert.Len(t, password, 10)
	
	for _, char := range password {
		assert.Contains(t, customChars, string(char))
	}
}

func TestGeneratePassword_InvalidOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    PasswordOptions
		wantErr string
	}{
		{
			"zero length",
			PasswordOptions{Length: 0, UseLower: true},
			"password length must be positive",
		},
		{
			"negative length",
			PasswordOptions{Length: -1, UseLower: true},
			"password length must be positive",
		},
		{
			"no character set",
			PasswordOptions{Length: 10},
			"no character set specified",
		},
		{
			"empty custom chars",
			PasswordOptions{Length: 10, CustomChars: ""},
			"no character set specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GeneratePassword(tt.opts)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGeneratePasswords(t *testing.T) {
	opts := DefaultPasswordOptions()
	passwords, err := GeneratePasswords(5, opts)

	assert.NoError(t, err)
	assert.Len(t, passwords, 5)

	// Check all passwords are different
	seen := make(map[string]bool)
	for _, password := range passwords {
		assert.NotContains(t, seen, password, "Generated duplicate password")
		seen[password] = true
	}
}

func TestGeneratePasswords_InvalidCount(t *testing.T) {
	opts := DefaultPasswordOptions()
	_, err := GeneratePasswords(0, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password count must be positive")
}

func TestGeneratePassword_Randomness(t *testing.T) {
	opts := PasswordOptions{
		Length:   10,
		UseLower: true,
		UseUpper: true,
		UseDigits: true,
	}

	// Generate two passwords and ensure they are different
	password1, err := GeneratePassword(opts)
	assert.NoError(t, err)

	password2, err := GeneratePassword(opts)
	assert.NoError(t, err)

	// It's extremely unlikely that two 10-character passwords with multiple character types will be identical
	assert.NotEqual(t, password1, password2, "Generated identical passwords, randomness may be compromised")
}