package dto

import "fmt"

func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("invalid email ")
	}
	if r.Password == "" {
		return fmt.Errorf("invalid password")
	}
	return nil
}
