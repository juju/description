// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/names/v6"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type StorageSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&StorageSerializationSuite{})

func (s *StorageSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "storages"
	s.sliceName = "storages"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importStorages(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["storages"] = []interface{}{}
	}
}

func testStorage() *storage {
	v := newStorage(testStorageArgs())
	return v
}

func testStorageArgs() StorageArgs {
	return StorageArgs{
		Tag:   names.NewStorageTag("db/0"),
		Kind:  "magic",
		Owner: names.NewApplicationTag("postgresql"),
		Name:  "db",
		Attachments: []names.UnitTag{
			names.NewUnitTag("postgresql/0"),
			names.NewUnitTag("postgresql/1"),
		},
		Constraints: &StorageInstanceConstraints{
			Pool: "radiance",
			Size: 1234,
		},
	}
}

func (s *StorageSerializationSuite) TestNewStorage(c *gc.C) {
	storage := testStorage()

	c.Check(storage.Tag(), gc.Equals, names.NewStorageTag("db/0"))
	c.Check(storage.Kind(), gc.Equals, "magic")
	owner, err := storage.Owner()
	c.Check(err, jc.ErrorIsNil)
	c.Check(owner, gc.Equals, names.NewApplicationTag("postgresql"))
	c.Check(storage.Name(), gc.Equals, "db")
	c.Check(storage.Attachments(), jc.DeepEquals, []names.UnitTag{
		names.NewUnitTag("postgresql/0"),
		names.NewUnitTag("postgresql/1"),
	})
}

func (s *StorageSerializationSuite) TestStorageValid(c *gc.C) {
	storage := testStorage()
	c.Assert(storage.Validate(), jc.ErrorIsNil)
}

func (s *StorageSerializationSuite) TestStorageValidMissingID(c *gc.C) {
	v := newStorage(StorageArgs{})
	err := v.Validate()
	c.Check(err, gc.ErrorMatches, `storage missing id not valid`)
	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *StorageSerializationSuite) TestStorageMatches(c *gc.C) {
	out, err := yaml.Marshal(testStorage())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.YAMLEquals, testStorage())
}

func (s *StorageSerializationSuite) TestStorageMatchesV2(c *gc.C) {
	testStorage := testStorage()
	testStorage.Owner_ = ""
	testStorage.Attachments_ = nil

	out, err := yaml.Marshal(testStorage)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.YAMLEquals, testStorage)
}

func (s *StorageSerializationSuite) exportImport(c *gc.C, storage_ *storage, version int) *storage {
	initial := storages{
		Version:   version,
		Storages_: []*storage{storage_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	storages, err := importStorages(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(storages, gc.HasLen, 1)
	return storages[0]
}

func (s *StorageSerializationSuite) TestParsingSerializedDataV1(c *gc.C) {
	original := testStorage()
	original.Constraints_ = nil
	storage := s.exportImport(c, original, 1)
	c.Assert(storage, jc.DeepEquals, original)
}

func (s *StorageSerializationSuite) TestParsingSerializedDataV2(c *gc.C) {
	original := testStorage()
	original.Owner_ = ""
	original.Attachments_ = nil
	original.Constraints_ = nil
	storage := s.exportImport(c, original, 2)
	c.Assert(storage, jc.DeepEquals, original)
}

func (s *StorageSerializationSuite) TestParsingSerializedDataV3(c *gc.C) {
	original := testStorage()
	original.Owner_ = ""
	original.Attachments_ = nil
	storage := s.exportImport(c, original, 3)
	c.Assert(storage, jc.DeepEquals, original)
}
