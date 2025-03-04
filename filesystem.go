// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"regexp"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

type filesystems struct {
	Version      int           `yaml:"version"`
	Filesystems_ []*filesystem `yaml:"filesystems"`
}

type filesystem struct {
	ID_        string `yaml:"id"`
	StorageID_ string `yaml:"storage-id,omitempty"`
	VolumeID_  string `yaml:"volume-id,omitempty"`

	Provisioned_  bool   `yaml:"provisioned"`
	Size_         uint64 `yaml:"size"`
	Pool_         string `yaml:"pool,omitempty"`
	FilesystemID_ string `yaml:"filesystem-id,omitempty"`

	Status_        *status `yaml:"status"`
	StatusHistory_ `yaml:"status-history"`

	Attachments_ filesystemAttachments `yaml:"attachments"`
}

type filesystemAttachments struct {
	Version      int                     `yaml:"version"`
	Attachments_ []*filesystemAttachment `yaml:"attachments"`
}

type filesystemAttachment struct {
	MachineID_   string `yaml:"host-machine-id,omitempty"`
	UnitID_      string `yaml:"host-unit-id,omitempty"`
	Provisioned_ bool   `yaml:"provisioned"`
	MountPoint_  string `yaml:"mount-point,omitempty"`
	ReadOnly_    bool   `yaml:"read-only"`
}

// FilesystemArgs is an argument struct used to add a filesystem to the Model.
type FilesystemArgs struct {
	ID           string
	Storage      string
	Volume       string
	Provisioned  bool
	Size         uint64
	Pool         string
	FilesystemID string
}

func newFilesystem(args FilesystemArgs) *filesystem {
	f := &filesystem{
		ID_:            args.ID,
		StorageID_:     args.Storage,
		VolumeID_:      args.Volume,
		Provisioned_:   args.Provisioned,
		Size_:          args.Size,
		Pool_:          args.Pool,
		FilesystemID_:  args.FilesystemID,
		StatusHistory_: newStatusHistory(),
	}
	f.setAttachments(nil)
	return f
}

// ID implements Filesystem.
func (f *filesystem) ID() string {
	return f.ID_
}

// Volume implements Filesystem.
func (f *filesystem) Volume() string {
	return f.VolumeID_
}

// Storage implements Filesystem.
func (f *filesystem) Storage() string {
	return f.StorageID_
}

// Provisioned implements Filesystem.
func (f *filesystem) Provisioned() bool {
	return f.Provisioned_
}

// Size implements Filesystem.
func (f *filesystem) Size() uint64 {
	return f.Size_
}

// Pool implements Filesystem.
func (f *filesystem) Pool() string {
	return f.Pool_
}

// FilesystemID implements Filesystem.
func (f *filesystem) FilesystemID() string {
	return f.FilesystemID_
}

// Status implements Filesystem.
func (f *filesystem) Status() Status {
	// To avoid typed nils check nil here.
	if f.Status_ == nil {
		return nil
	}
	return f.Status_
}

// SetStatus implements Filesystem.
func (f *filesystem) SetStatus(args StatusArgs) {
	f.Status_ = newStatus(args)
}

func (f *filesystem) setAttachments(attachments []*filesystemAttachment) {
	f.Attachments_ = filesystemAttachments{
		Version:      3,
		Attachments_: attachments,
	}
}

// Attachments implements Filesystem.
func (f *filesystem) Attachments() []FilesystemAttachment {
	var result []FilesystemAttachment
	for _, attachment := range f.Attachments_.Attachments_ {
		result = append(result, attachment)
	}
	return result
}

// AddAttachment implements Filesystem.
func (f *filesystem) AddAttachment(args FilesystemAttachmentArgs) FilesystemAttachment {
	a := newFilesystemAttachment(args)
	f.Attachments_.Attachments_ = append(f.Attachments_.Attachments_, a)
	return a
}

// Validate implements Filesystem.
func (f *filesystem) Validate() error {
	if f.ID_ == "" {
		return errors.NotValidf("filesystem missing id")
	}
	if f.Size_ == 0 {
		return errors.NotValidf("filesystem %q missing size", f.ID_)
	}
	if f.Status_ == nil {
		return errors.NotValidf("filesystem %q missing status", f.ID_)
	}
	return nil
}

func importFilesystems(source map[string]interface{}) ([]*filesystem, error) {
	checker := versionedChecker("filesystems")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "filesystems version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := filesystemDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["filesystems"].([]interface{})
	return importFilesystemList(sourceList, importFunc)
}

func importFilesystemList(sourceList []interface{}, importFunc filesystemDeserializationFunc) ([]*filesystem, error) {
	result := make([]*filesystem, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for filesystem %d, %T", i, value)
		}
		filesystem, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "filesystem %d", i)
		}
		result = append(result, filesystem)
	}
	return result, nil
}

type filesystemDeserializationFunc func(map[string]interface{}) (*filesystem, error)

var filesystemDeserializationFuncs = map[int]filesystemDeserializationFunc{
	1: importFilesystemV1,
}

func importFilesystemV1(source map[string]interface{}) (*filesystem, error) {
	fields := schema.Fields{
		"id":            schema.String(),
		"storage-id":    schema.String(),
		"volume-id":     schema.String(),
		"provisioned":   schema.Bool(),
		"size":          schema.ForceUint(),
		"pool":          schema.String(),
		"filesystem-id": schema.String(),
		"status":        schema.StringMap(schema.Any()),
		"attachments":   schema.StringMap(schema.Any()),
	}

	defaults := schema.Defaults{
		"storage-id":    "",
		"volume-id":     "",
		"pool":          "",
		"filesystem-id": "",
		"attachments":   schema.Omit,
	}
	addStatusHistorySchema(fields, defaults)
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "filesystem v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &filesystem{
		ID_:            valid["id"].(string),
		StorageID_:     valid["storage-id"].(string),
		VolumeID_:      valid["volume-id"].(string),
		Provisioned_:   valid["provisioned"].(bool),
		Size_:          valid["size"].(uint64),
		Pool_:          valid["pool"].(string),
		FilesystemID_:  valid["filesystem-id"].(string),
		StatusHistory_: newStatusHistory(),
	}
	if err := result.importStatusHistory(valid); err != nil {
		return nil, errors.Trace(err)
	}

	status, err := importStatus(valid["status"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.Status_ = status

	attachments, err := importFilesystemAttachments(valid["attachments"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.setAttachments(attachments)

	return result, nil
}

// FilesystemAttachmentArgs is an argument struct used to add information about the
// cloud instance to a Filesystem.
type FilesystemAttachmentArgs struct {
	HostUnit    string
	HostMachine string
	Provisioned bool
	ReadOnly    bool
	MountPoint  string
}

func newFilesystemAttachment(args FilesystemAttachmentArgs) *filesystemAttachment {
	return &filesystemAttachment{
		UnitID_:      args.HostUnit,
		MachineID_:   args.HostMachine,
		Provisioned_: args.Provisioned,
		ReadOnly_:    args.ReadOnly,
		MountPoint_:  args.MountPoint,
	}
}

// HostUnit implements FilesystemAttachment
func (a *filesystemAttachment) HostUnit() (string, bool) {
	return a.UnitID_, a.UnitID_ != ""
}

// HostMachine implements FilesystemAttachment
func (a *filesystemAttachment) HostMachine() (string, bool) {
	return a.MachineID_, a.MachineID_ != ""
}

// Provisioned implements FilesystemAttachment
func (a *filesystemAttachment) Provisioned() bool {
	return a.Provisioned_
}

// ReadOnly implements FilesystemAttachment
func (a *filesystemAttachment) ReadOnly() bool {
	return a.ReadOnly_
}

// MountPoint implements FilesystemAttachment
func (a *filesystemAttachment) MountPoint() string {
	return a.MountPoint_
}

func importFilesystemAttachments(source map[string]interface{}) ([]*filesystemAttachment, error) {
	checker := versionedChecker("attachments")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "filesystem attachments version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := filesystemAttachmentDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["attachments"].([]interface{})
	return importFilesystemAttachmentList(sourceList, importFunc)
}

func importFilesystemAttachmentList(sourceList []interface{}, importFunc filesystemAttachmentDeserializationFunc) ([]*filesystemAttachment, error) {
	result := make([]*filesystemAttachment, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for filesystemAttachment %d, %T", i, value)
		}
		filesystemAttachment, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "filesystemAttachment %d", i)
		}
		result = append(result, filesystemAttachment)
	}
	return result, nil
}

type filesystemAttachmentDeserializationFunc func(map[string]interface{}) (*filesystemAttachment, error)

var filesystemAttachmentDeserializationFuncs = map[int]filesystemAttachmentDeserializationFunc{
	1: importFilesystemAttachmentV1,
	2: importFilesystemAttachmentV2,
	3: importFilesystemAttachmentV3,
}

func importFilesystemAttachmentV1(source map[string]interface{}) (*filesystemAttachment, error) {
	fields, defaults := filesystemAttachmentV1Fields()
	return importFilesystemAttachment(fields, defaults, 1, source)
}

func importFilesystemAttachmentV2(source map[string]interface{}) (*filesystemAttachment, error) {
	fields, defaults := filesystemAttachmentV2Fields()
	return importFilesystemAttachment(fields, defaults, 2, source)
}

func importFilesystemAttachmentV3(source map[string]interface{}) (*filesystemAttachment, error) {
	fields, defaults := filesystemAttachmentV3Fields()
	return importFilesystemAttachment(fields, defaults, 3, source)
}

func filesystemAttachmentV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"machine-id":  schema.String(),
		"provisioned": schema.Bool(),
		"read-only":   schema.Bool(),
		"mount-point": schema.String(),
	}
	defaults := schema.Defaults{
		"mount-point": "",
	}
	return fields, defaults
}

func filesystemAttachmentV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := filesystemAttachmentV1Fields()
	fields["host-id"] = schema.String()
	delete(fields, "machine-id")
	return fields, defaults
}

func filesystemAttachmentV3Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := filesystemAttachmentV2Fields()
	fields["host-unit-id"] = schema.String()
	defaults["host-unit-id"] = schema.Omit
	fields["host-machine-id"] = schema.String()
	defaults["host-machine-id"] = schema.Omit
	delete(fields, "host-id")
	delete(defaults, "host-id")
	return fields, defaults
}

const (
	numberSnippet      = "(?:0|[1-9][0-9]*)"
	applicationSnippet = "(?:[a-z][a-z0-9]*(?:-[a-z0-9]*[a-z][a-z0-9]*)*)"
	unitSnippet        = "(" + applicationSnippet + ")/(" + numberSnippet + ")"
)

var validUnit = regexp.MustCompile("^" + unitSnippet + "$")

func importFilesystemAttachment(fields schema.Fields, defaults schema.Defaults, importVersion int, source map[string]interface{}) (*filesystemAttachment, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "filesystemAttachment schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	result := &filesystemAttachment{
		Provisioned_: valid["provisioned"].(bool),
		ReadOnly_:    valid["read-only"].(bool),
		MountPoint_:  valid["mount-point"].(string),
	}

	switch importVersion {
	case 1:
		result.MachineID_ = valid["machine-id"].(string)
	case 2:
		host := valid["host-id"].(string)
		if validUnit.MatchString(host) {
			result.UnitID_ = host
		} else {
			result.MachineID_ = host
		}
	default:
		if m, ok := valid["host-machine-id"].(string); ok {
			result.MachineID_ = m
		}
		if u, ok := valid["host-unit-id"].(string); ok {
			result.UnitID_ = u
		}
	}

	return result, nil
}
