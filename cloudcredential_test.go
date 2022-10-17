// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type CloudCredentialSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&CloudCredentialSerializationSuite{})

func (s *CloudCredentialSerializationSuite) SetUpTest(c *gc.C) {
	s.SerializationSuite.SetUpTest(c)
	s.importName = "cloudCredential"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importCloudCredential(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["owner"] = ""
		m["cloud"] = ""
		m["name"] = ""
		m["auth-type"] = ""
	}
}

func (s *CloudCredentialSerializationSuite) TestMissingOwner(c *gc.C) {
	testMap := s.makeMap(1)
	delete(testMap, "owner")
	_, err := importCloudCredential(testMap)
	c.Check(err.Error(), gc.Equals, "cloudCredential v1 schema check failed: owner: expected string, got nothing")
}

func (s *CloudCredentialSerializationSuite) TestMissingCloud(c *gc.C) {
	testMap := s.makeMap(1)
	delete(testMap, "cloud")
	_, err := importCloudCredential(testMap)
	c.Check(err.Error(), gc.Equals, "cloudCredential v1 schema check failed: cloud: expected string, got nothing")
}

func (s *CloudCredentialSerializationSuite) TestMissingName(c *gc.C) {
	testMap := s.makeMap(1)
	delete(testMap, "name")
	_, err := importCloudCredential(testMap)
	c.Check(err.Error(), gc.Equals, "cloudCredential v1 schema check failed: name: expected string, got nothing")
}

func (s *CloudCredentialSerializationSuite) TestMissingAuthType(c *gc.C) {
	testMap := s.makeMap(1)
	delete(testMap, "auth-type")
	_, err := importCloudCredential(testMap)
	c.Check(err.Error(), gc.Equals, "cloudCredential v1 schema check failed: auth-type: expected string, got nothing")
}

func (*CloudCredentialSerializationSuite) allArgs() CloudCredentialArgs {
	return CloudCredentialArgs{
		Owner:    names.NewUserTag("me"),
		Cloud:    names.NewCloudTag("altostratus"),
		Name:     "creds",
		AuthType: "fuzzy",
		Attributes: map[string]string{
			"key": "value",
		},
	}
}

func (s *CloudCredentialSerializationSuite) TestAllArgs(c *gc.C) {
	args := s.allArgs()
	creds := newCloudCredential(args)

	c.Check(creds.Owner(), gc.Equals, args.Owner.Id())
	c.Check(creds.Cloud(), gc.Equals, args.Cloud.Id())
	c.Check(creds.Name(), gc.Equals, args.Name)
	c.Check(creds.AuthType(), gc.Equals, args.AuthType)
	c.Check(creds.Attributes(), jc.DeepEquals, args.Attributes)
}

func (s *CloudCredentialSerializationSuite) TestParsingSerializedData(c *gc.C) {
	args := s.allArgs()
	initial := newCloudCredential(args)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	imported, err := importCloudCredential(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(imported, jc.DeepEquals, initial)
}

func (s *CloudCredentialSerializationSuite) TestV2MigrationSteps(c *gc.C) {
	tests := []struct {
		InitialSource  map[string]interface{}
		PostAuthType   string
		PostAttributes map[string]string
	}{
		{
			InitialSource: map[string]interface{}{
				"version":   1,
				"owner":     "wallyworld",
				"cloud":     "k8s",
				"name":      "ipv6rockz",
				"auth-type": "oauth2withcert",
				"attributes": map[string]string{
					"ClientCertificateData": "aa=",
					"ClientKeyData":         "aa=",
				},
			},
			PostAuthType: "clientcertificate",
			PostAttributes: map[string]string{
				"ClientCertificateData": "aa=",
				"ClientKeyData":         "aa=",
			},
		},
		{
			InitialSource: map[string]interface{}{
				"version":   1,
				"owner":     "wallyworld",
				"cloud":     "k8s",
				"name":      "ipv6rockz",
				"auth-type": "oauth2withcert",
				"attributes": map[string]string{
					"Token": "aa=",
				},
			},
			PostAuthType: "oauth2",
			PostAttributes: map[string]string{
				"Token": "aa=",
			},
		},
		{
			InitialSource: map[string]interface{}{
				"version":   1,
				"owner":     "wallyworld",
				"cloud":     "k8s",
				"name":      "ipv6rockz",
				"auth-type": "certificate",
				"attributes": map[string]string{
					"Token": "aa=",
				},
			},
			PostAuthType: "oauth2",
			PostAttributes: map[string]string{
				"Token": "aa=",
			},
		},
		{
			InitialSource: map[string]interface{}{
				"version":   1,
				"owner":     "wallyworld",
				"cloud":     "k8s",
				"name":      "ipv6rockz",
				"auth-type": "certificate",
				"attributes": map[string]string{
					"ClientCertificateData": "aa=",
					"ClientKeyData":         "aa=",
				},
			},
			PostAuthType: "certificate",
			PostAttributes: map[string]string{
				"ClientCertificateData": "aa=",
				"ClientKeyData":         "aa=",
			},
		},
		{
			InitialSource: map[string]interface{}{
				"version":   1,
				"owner":     "wallyworld",
				"cloud":     "k8s",
				"name":      "ipv6rockz",
				"auth-type": "certificate",
			},
			PostAuthType:   "certificate",
			PostAttributes: nil,
		},
	}

	for _, test := range tests {
		cred, err := importCloudCredential(test.InitialSource)
		c.Assert(err, jc.ErrorIsNil)
		c.Assert(cred.AuthType(), gc.Equals, test.PostAuthType)
		c.Assert(cred.Attributes(), jc.DeepEquals, test.PostAttributes)
	}
}

func (s *CloudCredentialSerializationSuite) TestV1MigrationToLatest(c *gc.C) {
	InitialSource := map[string]interface{}{
		"version":   1,
		"owner":     "wallyworld",
		"cloud":     "k8s",
		"name":      "ipv6rockz",
		"auth-type": "certificate",
		"attributes": map[string]string{
			"ClientCertificateData": "aa=",
			"ClientKeyData":         "aa=",
		},
	}

	cred, err := importCloudCredential(InitialSource)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(cred.Version, gc.Equals, 2)
	c.Assert(cred, gc.DeepEquals, &cloudCredential{
		Version:   2,
		Owner_:    "wallyworld",
		Cloud_:    "k8s",
		Name_:     "ipv6rockz",
		AuthType_: "certificate",
		Attributes_: map[string]string{
			"ClientCertificateData": "aa=",
			"ClientKeyData":         "aa=",
		},
	})
}
