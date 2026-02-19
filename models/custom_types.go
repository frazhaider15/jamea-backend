package models

import "fmt"

type Module string

func (d Module) String() string {
	return string(d)
}

func NewModuleString(s string) (Module, error) {
	status := Module(s)
	switch status {
	case ModuleAms:
	case ModuleTbum:
	case ModulePtbum:
	default:
		return status, fmt.Errorf("invalid module: %v", s)
	}
	return status, nil
}
