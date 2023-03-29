package utils

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

func ValidatePasswordPattern(password string) error {
	if len(password) < 8 {
		return errors.New("password should be of 8 characters long")
	}

	done, err := regexp.MatchString("([a-z])+", password)
	if err != nil {
		return err
	}

	if !done {
		return errors.New("password should contain at least one lower case character")
	}

	done, err = regexp.MatchString("([A-Z])+", password)
	if err != nil {
		return err
	}

	if !done {
		return errors.New("password should contain at least one upper case character")
	}

	done, err = regexp.MatchString("([0-9])+", password)
	if err != nil {
		return err
	}

	if !done {
		return errors.New("password should contain at least one digit")
	}

	done, err = regexp.MatchString("([!@#$%^&*.?-])+", password)
	if err != nil {
		return err
	}

	if !done {
		return errors.New("password should contain at least one special character")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}
