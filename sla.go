// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

// SLA represents the sla for the model.
type SLA interface {
	// Level returns the level of the sla.
	Level() string
	// Credentials returns the credentials of the sla.
	Credentials() string
}

type sla struct {
	Level_       string `yaml:"level"`
	Credentials_ string `yaml:"credentials"`
}

// Level returns the level of the sla.
func (s sla) Level() string {
	return s.Level_
}

// Credentials returns the Credentials of the sla.
func (s sla) Credentials() string {
	return s.Credentials_
}
