package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits           = "0123456789"
	specialChars     = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

type PasswordOptions struct {
	Length      int
	UseUpper    bool
	UseLower    bool
	UseDigits   bool
	UseSpecial  bool
	CustomChars string
}

// DefaultPasswordOptions returns default password options
func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		Length:     16,
		UseUpper:   true,
		UseLower:   true,
		UseDigits:  true,
		UseSpecial: false,
	}
}

// GeneratePassword generates a random password based on given options
func GeneratePassword(opts PasswordOptions) (string, error) {
	if opts.Length <= 0 {
		return "", fmt.Errorf("password length must be positive")
	}

	var chars string
	if opts.CustomChars != "" {
		chars = opts.CustomChars
	} else {
		if opts.UseLower {
			chars += lowercaseLetters
		}
		if opts.UseUpper {
			chars += uppercaseLetters
		}
		if opts.UseDigits {
			chars += digits
		}
		if opts.UseSpecial {
			chars += specialChars
		}
	}

	if chars == "" {
		return "", fmt.Errorf("no character set specified for password generation")
	}

	password := make([]byte, opts.Length)
	charLen := len(chars)
	bigCharLen := big.NewInt(int64(charLen))
	
	for i := 0; i < opts.Length; i++ {
		randomIndex, err := rand.Int(rand.Reader, bigCharLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = chars[randomIndex.Int64()]
	}

	return string(password), nil
}

// GeneratePasswords generates multiple random passwords
func GeneratePasswords(count int, opts PasswordOptions) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("password count must be positive")
	}

	passwords := make([]string, count)
	for i := 0; i < count; i++ {
		password, err := GeneratePassword(opts)
		if err != nil {
			return nil, err
		}
		passwords[i] = password
	}
	return passwords, nil
}
