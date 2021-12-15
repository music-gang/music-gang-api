package util_test

import (
	"testing"

	"github.com/music-gang/music-gang-api/app/util"
)

func TestPassword_CompareHashAndPassword(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		password := "123456"

		hashedPassword, err := util.HashPassword(password)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := util.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ErrNotMatch", func(t *testing.T) {

		password := "123456"

		hashedPassword, err := util.HashPassword(password)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if err := util.CompareHashAndPassword(hashedPassword, []byte("wrong-password")); err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestPasword_IsValidPassword(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		password := "AsecurePassword123!"

		if ok := util.IsValidPassword(password); !ok {
			t.Errorf("Expected true, got false")
		}
	})

	t.Run("ErrTooShort", func(t *testing.T) {

		password := "As1!"

		if ok := util.IsValidPassword(password); ok {
			t.Errorf("Expected false, got true")
		}
	})

	t.Run("ErrTooLong", func(t *testing.T) {

		password := "As1!butTooLooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong"

		if ok := util.IsValidPassword(password); ok {
			t.Errorf("Expected false, got true")
		}
	})

	t.Run("ErrNoUpperCase", func(t *testing.T) {

		password := "asdfghjkl1!"

		if ok := util.IsValidPassword(password); ok {
			t.Errorf("Expected false, got true")
		}
	})

	t.Run("ErrNoLowerCase", func(t *testing.T) {

		password := "ASDFGHJKL1!"

		if ok := util.IsValidPassword(password); ok {
			t.Errorf("Expected false, got true")
		}
	})

	t.Run("ErrNoNumber", func(t *testing.T) {

		password := "ASDFGHJKL!"

		if ok := util.IsValidPassword(password); ok {
			t.Errorf("Expected false, got true")
		}
	})

	t.Run("ErrNoSpecialCharacter", func(t *testing.T) {

		password := "ASDFGHghjkl123"

		if ok := util.IsValidPassword(password); ok {
			t.Errorf("Expected false, got true")
		}
	})
}
