// Copyright 2025 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type VirtualHostKeysSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&VirtualHostKeysSerializationSuite{})

func (s *VirtualHostKeysSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "virtual host keys"
	s.sliceName = "virtual-host-keys"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importVirtualHostKeys(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["virtual-host-keys"] = []interface{}{}
	}
}

func testVirtualHostKeyArgs() VirtualHostKeyArgs {
	return VirtualHostKeyArgs{
		ID:      "test-id",
		HostKey: []byte("foo"),
	}
}

func (s *VirtualHostKeysSerializationSuite) TestNewVirtualHostKey(c *gc.C) {
	args := testVirtualHostKeyArgs()
	virtualHostKey := newVirtualHostKey(args)

	c.Check(virtualHostKey.ID(), gc.Equals, args.ID)
	c.Check(virtualHostKey.HostKey(), gc.DeepEquals, []byte("foo"))
}

func (s *VirtualHostKeysSerializationSuite) TestVirtualHostKeyValid(c *gc.C) {
	args := testVirtualHostKeyArgs()
	virtualHostKey := newVirtualHostKey(args)
	c.Assert(virtualHostKey.Validate(), jc.ErrorIsNil)
}

func (s *VirtualHostKeysSerializationSuite) TestValidation(c *gc.C) {
	v := newVirtualHostKey(VirtualHostKeyArgs{ID: "", HostKey: []byte("foo")})
	err := v.Validate()
	c.Assert(err, gc.ErrorMatches, `empty id not valid`)

	v = newVirtualHostKey(VirtualHostKeyArgs{ID: "test-id", HostKey: []byte{}})
	err = v.Validate()
	c.Assert(err, gc.ErrorMatches, `zero length key not valid`)
}

func (s *VirtualHostKeysSerializationSuite) TestVirtualHostKeyMatches(c *gc.C) {
	args := testVirtualHostKeyArgs()

	virtualHostKey := newVirtualHostKey(args)
	out, err := yaml.Marshal(virtualHostKey)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.YAMLEquals, virtualHostKey)
}

func (s *VirtualHostKeysSerializationSuite) exportImport(c *gc.C, virtualHostKey_ *virtualHostKey, version int, expectError string) *virtualHostKey {
	initial := virtualHostKeys{
		Version:         version,
		VirtualHostKeys: []*virtualHostKey{virtualHostKey_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	virtualHostKeys, err := importVirtualHostKeys(source)
	if expectError != "" {
		c.Assert(err, gc.ErrorMatches, expectError)
		return nil
	}
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(virtualHostKeys, gc.HasLen, 1)
	return virtualHostKeys[0]
}

func (s *VirtualHostKeysSerializationSuite) TestParsingSerializedData(c *gc.C) {
	args := testVirtualHostKeyArgs()
	original := newVirtualHostKey(args)
	virtualHostKey := s.exportImport(c, original, 1, "")
	c.Assert(virtualHostKey, jc.DeepEquals, original)
}

func (s *VirtualHostKeysSerializationSuite) TestParsingInvalidHostKey(c *gc.C) {
	args := testVirtualHostKeyArgs()
	original := newVirtualHostKey(args)
	original.HostKey_ = "invalid"
	_ = s.exportImport(c, original, 1, ".*virtual host key not valid.*")
}
