// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"strings"

	"github.com/juju/errors"
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
		ID:        "db/0",
		Kind:      "magic",
		UnitOwner: "postgresql/0",
		Name:      "db",
		Attachments: []string{
			"postgresql/0",
			"postgresql/1",
		},
		Constraints: &StorageInstanceConstraints{
			Pool: "radiance",
			Size: 1234,
		},
	}
}

func (s *StorageSerializationSuite) TestNewStorage(c *gc.C) {
	storage := testStorage()

	c.Check(storage.ID(), gc.Equals, "db/0")
	c.Check(storage.Kind(), gc.Equals, "magic")
	owner, ok := storage.UnitOwner()
	c.Check(ok, jc.IsTrue)
	c.Check(owner, gc.Equals, "postgresql/0")
	c.Check(storage.Name(), gc.Equals, "db")
	c.Check(storage.Attachments(), jc.DeepEquals, []string{
		"postgresql/0",
		"postgresql/1",
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
	c.Check(err, jc.ErrorIs, errors.NotValid)
}

func (s *StorageSerializationSuite) TestStorageMatches(c *gc.C) {
	out, err := yaml.Marshal(testStorage())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(string(out), jc.YAMLEquals, testStorage())
}

func (s *StorageSerializationSuite) TestStorageMatchesV2(c *gc.C) {
	testStorage := testStorage()
	testStorage.UnitOwner_ = ""
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

func (s *StorageSerializationSuite) TestParsingSerializedDataV2(c *gc.C) {
	original := testStorage()
	original.UnitOwner_ = ""
	original.Attachments_ = nil
	original.Constraints_ = nil
	storage := s.exportImport(c, original, 2)
	c.Assert(storage, jc.DeepEquals, original)
}

func (s *StorageSerializationSuite) TestParsingSerializedDataV3(c *gc.C) {
	original := testStorage()
	initial := storages{
		Version:   3,
		Storages_: []*storage{original},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	legacy := strings.ReplaceAll(string(bytes), "unit-owner: postgresql/0", "owner: unit-postgresql-0")

	var source map[string]interface{}
	err = yaml.Unmarshal([]byte(legacy), &source)
	c.Assert(err, jc.ErrorIsNil)

	storages, err := importStorages(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(storages, gc.HasLen, 1)
	c.Assert(storages[0], jc.DeepEquals, original)
}

func (s *StorageSerializationSuite) TestParsingSerializedDataV3UnitNameWithDash(c *gc.C) {
	// Arrange a storage using the version 4 args but a unit tag for the
	// owner. Then convert to the version 3 by replacing 'owner' with
	// 'unit-owner'.
	original := newStorage(StorageArgs{
		ID:        "db/0",
		Kind:      "magic",
		UnitOwner: "unit-post-gresql-0",
		Name:      "db",
		Attachments: []string{
			"post-gresql/0",
			"postgresql/1",
		},
		Constraints: &StorageInstanceConstraints{
			Pool: "radiance",
			Size: 1234,
		},
	})
	initial := storages{
		Version:   3,
		Storages_: []*storage{original},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	legacy := strings.ReplaceAll(string(bytes), "unit-owner", "owner")

	var source map[string]interface{}
	err = yaml.Unmarshal([]byte(legacy), &source)
	c.Assert(err, jc.ErrorIsNil)

	// Act
	storages, err := importStorages(source)

	// Assert: UnitOwner_ is now the unit name rather than the unit tag.
	// Ensure that a dash in the application name is not replaced.
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(storages, gc.HasLen, 1)
	c.Assert(storages[0], jc.DeepEquals, &storage{
		ID_:        "db/0",
		Kind_:      "magic",
		UnitOwner_: "post-gresql/0",
		Name_:      "db",
		Attachments_: []string{
			"post-gresql/0",
			"postgresql/1",
		},
		Constraints_: &StorageInstanceConstraints{
			Pool: "radiance",
			Size: 1234,
		},
	})
}

func (s *StorageSerializationSuite) TestParsingSerializedDataV4(c *gc.C) {
	original := testStorage()
	original.Attachments_ = nil
	storage := s.exportImport(c, original, 4)
	c.Assert(storage, jc.DeepEquals, original)
}
