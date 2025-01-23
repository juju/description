// Copyright 2023 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/names/v6"
	jc "github.com/juju/testing/checkers"
	"github.com/rs/xid"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type RemoteSecretsSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&RemoteSecretsSerializationSuite{})

func (s *RemoteSecretsSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "remote secret"
	s.sliceName = "remote-secrets"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importRemoteSecrets(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["remote-secrets"] = []interface{}{}
	}
}

func testRemoteSecretArgs() RemoteSecretArgs {
	id := xid.New().String()
	return RemoteSecretArgs{
		ID:              id,
		SourceUUID:      "model-uuid",
		Consumer:        names.NewApplicationTag("mariadb"),
		Label:           "secret label",
		CurrentRevision: 666,
		LatestRevision:  667,
	}
}

func (s *RemoteSecretsSerializationSuite) TestNewRemoteSecret(c *gc.C) {
	args := testRemoteSecretArgs()
	remoteSecret := newRemoteSecret(args)

	c.Check(remoteSecret.ID(), gc.Equals, args.ID)
	c.Check(remoteSecret.SourceUUID(), gc.Equals, "model-uuid")
	c.Check(remoteSecret.Label(), gc.Equals, "secret label")
	c.Check(remoteSecret.CurrentRevision(), gc.Equals, 666)
	c.Check(remoteSecret.LatestRevision(), gc.Equals, 667)
	consumer, err := remoteSecret.Consumer()
	c.Check(err, jc.ErrorIsNil)
	c.Check(consumer, gc.Equals, names.NewApplicationTag("mariadb"))
}

func (s *RemoteSecretsSerializationSuite) TestRemoteSecretValid(c *gc.C) {
	args := testRemoteSecretArgs()
	remoteSecret := newRemoteSecret(args)
	c.Assert(remoteSecret.Validate(), jc.ErrorIsNil)
}

func (s *RemoteSecretsSerializationSuite) TestInvalidID(c *gc.C) {
	v := newRemoteSecret(RemoteSecretArgs{ID: "invalid"})
	err := v.Validate()
	c.Assert(err, gc.ErrorMatches, `remote secret ID "invalid" not valid`)
}

func (s *RemoteSecretsSerializationSuite) TestRemoteSecretMatches(c *gc.C) {
	args := testRemoteSecretArgs()

	remoteSecret := newRemoteSecret(args)
	out, err := yaml.Marshal(remoteSecret)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.YAMLEquals, remoteSecret)
}

func (s *RemoteSecretsSerializationSuite) exportImport(c *gc.C, remoteSecret_ *remoteSecret, version int) *remoteSecret {
	initial := remoteSecrets{
		Version:        version,
		RemoteSecrets_: []*remoteSecret{remoteSecret_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	remoteSecrets, err := importRemoteSecrets(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(remoteSecrets, gc.HasLen, 1)
	return remoteSecrets[0]
}

func (s *RemoteSecretsSerializationSuite) TestParsingSerializedData(c *gc.C) {
	args := testRemoteSecretArgs()
	original := newRemoteSecret(args)
	remoteSecret := s.exportImport(c, original, 1)
	c.Assert(remoteSecret, jc.DeepEquals, original)
}
