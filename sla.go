// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

type sla struct {
	Level_       string `yaml:"level"`
	Credentials_ []byte `yaml:"credentials"`
}

// Level returns the level of the sla.
func (s sla) Level() string {
	return s.Level_
}

// Level returns the Credentials of the sla.
func (s sla) Credentials() []byte {
	return s.Credentials_
}
