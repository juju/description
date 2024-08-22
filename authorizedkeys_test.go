// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"slices"
	"strings"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type AuthorizedKeysSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&AuthorizedKeysSerializationSuite{})

func (s *AuthorizedKeysSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "users-authorized-keys"
	s.sliceName = "users-authorized-keys"
	s.importFunc = func(m map[string]any) (any, error) {
		return importAuthorizedKeys(m)
	}
	s.testFields = func(m map[string]any) {
		m["users-authorized-keys"] = []any{}
	}
}

// TestNewUserAuthorizedKeys is testing that given a set of arguments we get
// back a [userAuthorizedKeys] struct that contains the information passed in
// via args.
func (s *AuthorizedKeysSerializationSuite) TestNewUserAuthorizedKeys(c *gc.C) {
	args := UserAuthorizedKeysArgs{
		Username: "tlm",
		AuthorizedKeys: []string{
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC existing1",
			"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe existing2",
		},
	}

	uak := newUserAuthorizedKeys(args)
	c.Check(uak.Username(), gc.Equals, "tlm")
	c.Check(uak.AuthorizedKeys(), jc.DeepEquals, []string{
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC existing1",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe existing2",
	})
}

// TestParsingSerializedData is asserting that we can marshal and unmarshal user
// authorized keys to from the go types to yaml and get back the data we expect.
// We give a two user example below to demonstrate a more complex scenario.
func (s *AuthorizedKeysSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := authorizedKeys{
		Version: 1,
		UserAuthorizedKeys_: []*userAuthorizedKeys{
			newUserAuthorizedKeys(UserAuthorizedKeysArgs{
				Username: "tlm",
				AuthorizedKeys: []string{
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC existing1",
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe existing2",
				},
			}),
			newUserAuthorizedKeys(UserAuthorizedKeysArgs{
				Username: "wallyworld",
				AuthorizedKeys: []string{
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC existing1",
					"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe existing2",
				},
			}),
		},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]any
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	usersAuthorizedKeys, err := importAuthorizedKeys(source)
	c.Assert(err, jc.ErrorIsNil)

	slices.SortFunc(usersAuthorizedKeys, func(a, b *userAuthorizedKeys) int {
		return strings.Compare(a.Username(), b.Username())
	})

	c.Check(usersAuthorizedKeys, jc.DeepEquals, []*userAuthorizedKeys{
		newUserAuthorizedKeys(UserAuthorizedKeysArgs{
			Username: "tlm",
			AuthorizedKeys: []string{
				"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC existing1",
				"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe existing2",
			},
		}),
		newUserAuthorizedKeys(UserAuthorizedKeysArgs{
			Username: "wallyworld",
			AuthorizedKeys: []string{
				"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII4GpCvqUUYUJlx6d1kpUO9k/t4VhSYsf0yE0/QTqDzC existing1",
				"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJQJ9wv0uC3yytXM3d2sJJWvZLuISKo7ZHwafHVviwVe existing2",
			},
		}),
	},
	)
}
