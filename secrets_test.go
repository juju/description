// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/names/v6"
	jc "github.com/juju/testing/checkers"
	"github.com/rs/xid"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type SecretsSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&SecretsSerializationSuite{})

func (s *SecretsSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "secrets"
	s.sliceName = "secrets"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importSecrets(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["secrets"] = []interface{}{}
	}
}

func testSecretArgs() SecretArgs {
	id := xid.New().String()
	created := time.Now().UTC()
	updated := created.Add(time.Hour)
	nextRotate := created.Add(2 * time.Hour)
	return SecretArgs{
		ID:                     id,
		Version:                1,
		Description:            "a secret",
		Label:                  "secret label",
		RotatePolicy:           "hourly",
		AutoPrune:              true,
		Owner:                  names.NewApplicationTag("postgresql"),
		Created:                created,
		Updated:                updated,
		NextRotateTime:         &nextRotate,
		LatestRevisionChecksum: "checksum",
		Revisions:              testSecretRevisionsArgs(),
		ACL:                    testSecretAccessArgs(),
		Consumers:              testSecretConsumerArgs(),
		RemoteConsumers:        testSecretRemoteConsumerArgs(),
	}
}

func testSecretRevisionsArgs() []SecretRevisionArgs {
	created := time.Now().UTC()
	updated := created.Add(time.Hour)
	expire := created.Add(2 * time.Hour)
	valueRef := SecretValueRefArgs{
		BackendID:  "backend-id",
		RevisionID: "rev-id",
	}
	return []SecretRevisionArgs{{
		Number:     1,
		Created:    created,
		Updated:    updated,
		ExpireTime: &expire,
		Content:    map[string]string{"foo": "bar"},
		Obsolete:   true,
	}, {
		Number:        2,
		Created:       created,
		Updated:       updated,
		ValueRef:      &valueRef,
		PendingDelete: true,
	}}
}

func testSecretAccessArgs() map[string]SecretAccessArgs {
	return map[string]SecretAccessArgs{
		"application-postgresql": {
			Scope: "application-postgresql",
			Role:  "manage",
		},
		"unit-mariadb-0": {
			Scope: "relation-mariadb:server#wordpress:db",
			Role:  "view",
		},
	}
}

func testSecretConsumerArgs() []SecretConsumerArgs {
	return []SecretConsumerArgs{{
		Consumer:        names.NewApplicationTag("mariadb"),
		Label:           "label 1",
		CurrentRevision: 1,
	}, {
		Consumer:        names.NewUnitTag("mariadb/0"),
		Label:           "label 2",
		CurrentRevision: 2,
	}}
}

func testSecretRemoteConsumerArgs() []SecretRemoteConsumerArgs {
	id := xid.New().String()
	return []SecretRemoteConsumerArgs{{
		ID:              id,
		Consumer:        names.NewApplicationTag("remote-mariadb"),
		CurrentRevision: 666,
	}}
}

func (s *SecretsSerializationSuite) TestNewSecret(c *gc.C) {
	args := testSecretArgs()
	secret := newSecret(args)

	c.Check(secret.Id(), gc.Equals, args.ID)
	c.Check(secret.Version(), gc.Equals, 1)
	c.Check(secret.Description(), gc.Equals, "a secret")
	c.Check(secret.Label(), gc.Equals, "secret label")
	c.Check(secret.RotatePolicy(), gc.Equals, "hourly")
	c.Check(secret.AutoPrune(), jc.IsTrue)
	c.Check(secret.Created(), jc.DeepEquals, args.Created)
	c.Check(secret.Updated(), jc.DeepEquals, args.Updated)
	c.Check(secret.NextRotateTime(), jc.DeepEquals, args.NextRotateTime)
	c.Check(secret.LatestRevision(), gc.Equals, 2)
	c.Check(secret.LatestRevisionChecksum(), gc.Equals, "checksum")
	c.Check(secret.LatestExpireTime(), jc.DeepEquals, args.Revisions[1].ExpireTime)
	owner, err := secret.Owner()
	c.Check(err, jc.ErrorIsNil)
	c.Check(owner, gc.Equals, names.NewApplicationTag("postgresql"))

	c.Check(secret.Revisions(), gc.HasLen, 2)
	c.Check(secret.Revisions()[0].Number(), gc.Equals, 1)
	c.Check(secret.Revisions()[0].Created(), jc.DeepEquals, args.Revisions[0].Created)
	c.Check(secret.Revisions()[0].Updated(), jc.DeepEquals, args.Revisions[0].Updated)
	c.Check(secret.Revisions()[0].ExpireTime(), jc.DeepEquals, args.Revisions[0].ExpireTime)
	c.Check(secret.Revisions()[0].Content(), gc.DeepEquals, map[string]string{"foo": "bar"})
	c.Check(secret.Revisions()[0].Obsolete(), jc.IsTrue)
	c.Check(secret.Revisions()[0].PendingDelete(), jc.IsFalse)
	c.Check(secret.Revisions()[0].ValueRef(), gc.IsNil)
	c.Check(secret.Revisions()[1].ValueRef().BackendID(), gc.Equals, args.Revisions[1].ValueRef.BackendID)
	c.Check(secret.Revisions()[1].ValueRef().RevisionID(), gc.Equals, args.Revisions[1].ValueRef.RevisionID)
	c.Check(secret.Revisions()[1].Obsolete(), jc.IsFalse)
	c.Check(secret.Revisions()[1].PendingDelete(), jc.IsTrue)
	c.Check(secret.Revisions()[1].Content(), gc.IsNil)
	c.Check(secret.ACL()["application-postgresql"].Role(), gc.Equals, "manage")
	c.Check(secret.ACL()["application-postgresql"].Scope(), gc.Equals, "application-postgresql")
	c.Check(secret.Consumers(), gc.HasLen, 2)
	c.Check(secret.Consumers()[0].Label(), gc.Equals, "label 1")
	c.Check(secret.Consumers()[0].CurrentRevision(), gc.Equals, 1)
	c.Check(secret.Consumers()[0].LatestRevision(), gc.Equals, 2)
	consumer, err := secret.Consumers()[0].Consumer()
	c.Check(err, jc.ErrorIsNil)
	c.Check(consumer, gc.Equals, names.NewApplicationTag("mariadb"))
	c.Check(secret.RemoteConsumers(), gc.HasLen, 1)
	c.Check(secret.RemoteConsumers()[0].CurrentRevision(), gc.Equals, 666)
	c.Check(secret.RemoteConsumers()[0].LatestRevision(), gc.Equals, 2)
	consumer, err = secret.RemoteConsumers()[0].Consumer()
	c.Check(err, jc.ErrorIsNil)
	c.Check(consumer, gc.Equals, names.NewApplicationTag("remote-mariadb"))
}

func (s *SecretsSerializationSuite) TestNewSecretNoRotatePolicy(c *gc.C) {
	args := testSecretArgs()
	args.RotatePolicy = ""
	secret := newSecret(args)

	c.Assert(secret.RotatePolicy(), gc.Equals, "")
}

func (s *SecretsSerializationSuite) TestNewSecretNoAutoPrune(c *gc.C) {
	args := testSecretArgs()
	args.AutoPrune = false
	secret := newSecret(args)

	c.Assert(secret.AutoPrune(), jc.IsFalse)
}

func (s *SecretsSerializationSuite) TestNewSecretNoNextRotateTime(c *gc.C) {
	args := testSecretArgs()
	args.NextRotateTime = nil
	secret := newSecret(args)

	c.Assert(secret.NextRotateTime_, gc.IsNil)
}

func (s *SecretsSerializationSuite) TestSecretValid(c *gc.C) {
	args := testSecretArgs()
	secret := newSecret(args)
	c.Assert(secret.Validate(), jc.ErrorIsNil)
}

func (s *SecretsSerializationSuite) TestInvalidID(c *gc.C) {
	v := newSecret(SecretArgs{ID: "invalid"})
	err := v.Validate()
	c.Assert(err, gc.ErrorMatches, `secret ID "invalid" not valid`)
}

func (s *SecretsSerializationSuite) TestComputedFields(c *gc.C) {
	args := testSecretArgs()
	secret := newSecret(args)

	c.Assert(secret.LatestRevision(), gc.Equals, 2)
	c.Assert(secret.LatestExpireTime(), gc.IsNil)
}

func (s *SecretsSerializationSuite) TestComputedExpireTimeNotNil(c *gc.C) {
	args := testSecretArgs()
	expireTime := time.Now()
	args.Revisions[1].ExpireTime = &expireTime
	secret := newSecret(args)

	c.Assert(secret.LatestRevision(), gc.Equals, 2)
	c.Assert(secret.LatestExpireTime(), gc.NotNil)
	c.Assert(secret.LatestExpireTime(), jc.DeepEquals, &expireTime)
}

func (s *SecretsSerializationSuite) TestSecretMatches(c *gc.C) {
	args := testSecretArgs()

	secret := newSecret(args)
	out, err := yaml.Marshal(secret)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.YAMLEquals, secret)
}

func (s *SecretsSerializationSuite) exportImport(c *gc.C, secret_ *secret, version int) *secret {
	initial := secrets{
		Version:  version,
		Secrets_: []*secret{secret_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	secrets, err := importSecrets(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(secrets, gc.HasLen, 1)
	return secrets[0]
}

func (s *SecretsSerializationSuite) TestParsingSerializedData(c *gc.C) {
	args := testSecretArgs()
	original := newSecret(args)
	secret := s.exportImport(c, original, 2)
	c.Assert(secret, jc.DeepEquals, original)
}

type oldSecret struct {
	ID_          string            `yaml:"id"`
	Version_     int               `yaml:"secret-version"`
	Description_ string            `yaml:"description"`
	Label_       string            `yaml:"label"`
	Owner_       string            `yaml:"owner"`
	Created_     time.Time         `yaml:"create-time"`
	Updated_     time.Time         `yaml:"update-time"`
	Revisions_   []*secretRevision `yaml:"revisions"`
}

type oldSecrets struct {
	Version  int          `yaml:"version"`
	Secrets_ []*oldSecret `yaml:"secrets"`
}

func (s *SecretsSerializationSuite) TestParsingSerializedDataNoAutoPrune(c *gc.C) {
	args := testSecretArgs()
	original := &oldSecret{
		ID_:      args.ID,
		Version_: args.Version,
		Created_: args.Created.UTC(),
		Updated_: args.Updated.UTC(),
	}
	initial := oldSecrets{
		Version:  1,
		Secrets_: []*oldSecret{original},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	secrets, err := importSecrets(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(secrets, gc.HasLen, 1)

	c.Assert(secrets[0], jc.DeepEquals, &secret{
		ID_:      original.ID_,
		Version_: original.Version_,
		Owner_:   original.Owner_,
		Created_: original.Created_,
		Updated_: original.Updated_,
	})
}
