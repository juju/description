// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type FilesystemSerializationSuite struct {
	SliceSerializationSuite
	StatusHistoryMixinSuite
}

var _ = gc.Suite(&FilesystemSerializationSuite{})

func (s *FilesystemSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "filesystems"
	s.sliceName = "filesystems"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importFilesystems(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["filesystems"] = []interface{}{}
	}
	s.StatusHistoryMixinSuite.creator = func() HasStatusHistory {
		return testFilesystem()
	}
	s.StatusHistoryMixinSuite.serializer = func(c *gc.C, initial interface{}) HasStatusHistory {
		return s.exportImport(c, initial.(*filesystem))
	}
}

func testFilesystemMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":             "1234",
		"storage-id":     "test/1",
		"volume-id":      "4321",
		"provisioned":    true,
		"size":           int(20 * gig),
		"pool":           "swimming",
		"filesystem-id":  "some filesystem id",
		"status":         minimalStatusMap(),
		"status-history": emptyStatusHistoryMap(),
		"attachments": map[interface{}]interface{}{
			"version":     3,
			"attachments": []interface{}{},
		},
	}
}

func testFilesystem() *filesystem {
	v := newFilesystem(testFilesystemArgs())
	v.SetStatus(minimalStatusArgs())
	return v
}

func testFilesystemArgs() FilesystemArgs {
	return FilesystemArgs{
		ID:           "1234",
		Storage:      "test/1",
		Volume:       "4321",
		Provisioned:  true,
		Size:         20 * gig,
		Pool:         "swimming",
		FilesystemID: "some filesystem id",
	}
}

func (s *FilesystemSerializationSuite) TestNewFilesystem(c *gc.C) {
	filesystem := testFilesystem()

	c.Check(filesystem.ID(), gc.Equals, "1234")
	c.Check(filesystem.Storage(), gc.Equals, "test/1")
	c.Check(filesystem.Volume(), gc.Equals, "4321")
	c.Check(filesystem.Provisioned(), jc.IsTrue)
	c.Check(filesystem.Size(), gc.Equals, 20*gig)
	c.Check(filesystem.Pool(), gc.Equals, "swimming")
	c.Check(filesystem.FilesystemID(), gc.Equals, "some filesystem id")

	c.Check(filesystem.Attachments(), gc.HasLen, 0)
}

func (s *FilesystemSerializationSuite) TestFilesystemValid(c *gc.C) {
	filesystem := testFilesystem()
	c.Assert(filesystem.Validate(), jc.ErrorIsNil)
}

func (s *FilesystemSerializationSuite) TestFilesystemValidMissingID(c *gc.C) {
	v := newFilesystem(FilesystemArgs{})
	err := v.Validate()
	c.Check(err, gc.ErrorMatches, `filesystem missing id not valid`)
	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *FilesystemSerializationSuite) TestFilesystemValidMissingSize(c *gc.C) {
	v := newFilesystem(FilesystemArgs{
		ID: "123",
	})
	err := v.Validate()
	c.Check(err, gc.ErrorMatches, `filesystem "123" missing size not valid`)
	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *FilesystemSerializationSuite) TestFilesystemValidMissingStatus(c *gc.C) {
	v := newFilesystem(FilesystemArgs{
		ID:   "123",
		Size: 5,
	})
	err := v.Validate()
	c.Check(err, gc.ErrorMatches, `filesystem "123" missing status not valid`)
	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *FilesystemSerializationSuite) TestFilesystemValidMinimal(c *gc.C) {
	v := newFilesystem(FilesystemArgs{
		ID:   "123",
		Size: 5,
	})
	v.SetStatus(minimalStatusArgs())
	err := v.Validate()
	c.Check(err, jc.ErrorIsNil)
}

func (s *FilesystemSerializationSuite) TestFilesystemMatches(c *gc.C) {
	bytes, err := yaml.Marshal(testFilesystem())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, testFilesystemMap())
}

func (s *FilesystemSerializationSuite) exportImport(c *gc.C, filesystem_ *filesystem) *filesystem {
	initial := filesystems{
		Version:      1,
		Filesystems_: []*filesystem{filesystem_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	filesystems, err := importFilesystems(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(filesystems, gc.HasLen, 1)
	return filesystems[0]
}

func (s *FilesystemSerializationSuite) TestAddingAttachments(c *gc.C) {
	// The core code does not care about duplicates, so we'll just add
	// the same attachment twice.
	original := testFilesystem()
	attachment1 := original.AddAttachment(testFilesystemAttachmentArgs("1"))
	attachment2 := original.AddAttachment(testFilesystemAttachmentArgs("2"))
	filesystem := s.exportImport(c, original)
	c.Assert(filesystem, jc.DeepEquals, original)
	attachments := filesystem.Attachments()
	c.Assert(attachments, gc.HasLen, 2)
	c.Check(attachments[0], jc.DeepEquals, attachment1)
	c.Check(attachments[1], jc.DeepEquals, attachment2)
}

func (s *FilesystemSerializationSuite) TestParsingSerializedData(c *gc.C) {
	original := testFilesystem()
	original.AddAttachment(testFilesystemAttachmentArgs())
	filesystem := s.exportImport(c, original)
	c.Assert(filesystem, jc.DeepEquals, original)
}

type FilesystemAttachmentSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&FilesystemAttachmentSerializationSuite{})

func (s *FilesystemAttachmentSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "filesystem attachments"
	s.sliceName = "attachments"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importFilesystemAttachments(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["attachments"] = []interface{}{}
	}
}

func testFilesystemAttachmentMap() map[string]interface{} {
	return map[string]interface{}{
		"host-machine-id": "42",
		"provisioned":     true,
		"read-only":       true,
		"mount-point":     "/some/dir",
	}
}

func testFilesystemAttachment() *filesystemAttachment {
	return newFilesystemAttachment(testFilesystemAttachmentArgs())
}

func testFilesystemAttachmentArgs(id ...string) FilesystemAttachmentArgs {
	machineID := "42"
	if len(id) > 0 {
		machineID = id[0]
	}
	return FilesystemAttachmentArgs{
		HostMachine: machineID,
		Provisioned: true,
		ReadOnly:    true,
		MountPoint:  "/some/dir",
	}
}

func (s *FilesystemAttachmentSerializationSuite) TestNewFilesystemAttachment(c *gc.C) {
	attachment := testFilesystemAttachment()

	m, ok := attachment.HostMachine()
	c.Check(ok, jc.IsTrue)
	c.Check(m, gc.Equals, "42")
	c.Check(attachment.Provisioned(), jc.IsTrue)
	c.Check(attachment.ReadOnly(), jc.IsTrue)
	c.Check(attachment.MountPoint(), gc.Equals, "/some/dir")
}

func (s *FilesystemAttachmentSerializationSuite) TestFilesystemAttachmentMatches(c *gc.C) {
	bytes, err := yaml.Marshal(testFilesystemAttachment())
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, testFilesystemAttachmentMap())
}

func (s *FilesystemAttachmentSerializationSuite) exportImportLatest(c *gc.C, attachment *filesystemAttachment) *filesystemAttachment {
	return s.exportImportVersion(c, attachment, 3)
}

func (s *FilesystemAttachmentSerializationSuite) exportImportVersion(c *gc.C, attachment *filesystemAttachment, version int) *filesystemAttachment {
	initial := filesystemAttachments{
		Version:      version,
		Attachments_: []*filesystemAttachment{attachment},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	attachments, err := importFilesystemAttachments(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attachments, gc.HasLen, 1)
	return attachments[0]
}

func (s *FilesystemAttachmentSerializationSuite) TestParsingSerializedData(c *gc.C) {
	original := testFilesystemAttachment()
	attachment := s.exportImportLatest(c, original)
	c.Assert(attachment, jc.DeepEquals, original)
}

func (s *FilesystemAttachmentSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	attachmentMapV1 := testFilesystemAttachmentMap()
	attachmentMapV1["machine-id"] = attachmentMapV1["host-machine-id"]
	delete(attachmentMapV1, "host-id")

	attachment, err := importFilesystemAttachmentV1(attachmentMapV1)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attachment, jc.DeepEquals, &filesystemAttachment{
		MachineID_:   "42",
		MountPoint_:  "/some/dir",
		ReadOnly_:    true,
		Provisioned_: true,
	})
}

func (s *FilesystemAttachmentSerializationSuite) TestV2ParsingReturnsLatest(c *gc.C) {
	attachmentMapV2 := testFilesystemAttachmentMap()
	attachmentMapV2["host-id"] = attachmentMapV2["host-machine-id"]
	delete(attachmentMapV2, "host-machine-id")

	attachment, err := importFilesystemAttachmentV2(attachmentMapV2)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attachment, jc.DeepEquals, &filesystemAttachment{
		MachineID_:   "42",
		MountPoint_:  "/some/dir",
		ReadOnly_:    true,
		Provisioned_: true,
	})
}

func (s *FilesystemAttachmentSerializationSuite) TestV2UnitParsingReturnsLatest(c *gc.C) {
	attachmentMapV2 := testFilesystemAttachmentMap()
	attachmentMapV2["host-id"] = "gitlab/0"
	delete(attachmentMapV2, "host-machine-id")

	attachment, err := importFilesystemAttachmentV2(attachmentMapV2)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attachment, jc.DeepEquals, &filesystemAttachment{
		UnitID_:      "gitlab/0",
		MountPoint_:  "/some/dir",
		ReadOnly_:    true,
		Provisioned_: true,
	})
}

func (s *FilesystemAttachmentSerializationSuite) TestUnitAttachmentParsing(c *gc.C) {
	attachmentMap := testFilesystemAttachmentMap()
	attachmentMap["host-unit-id"] = "gitlab/0"
	delete(attachmentMap, "host-machine-id")

	attachment, err := importFilesystemAttachmentV3(attachmentMap)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(attachment, jc.DeepEquals, &filesystemAttachment{
		UnitID_:      "gitlab/0",
		MountPoint_:  "/some/dir",
		ReadOnly_:    true,
		Provisioned_: true,
	})
}
