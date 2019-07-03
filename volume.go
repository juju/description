// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
	"gopkg.in/juju/names.v2"
)

type volumes struct {
	Version  int       `yaml:"version"`
	Volumes_ []*volume `yaml:"volumes"`
}

type volume struct {
	ID_          string `yaml:"id"`
	StorageID_   string `yaml:"storage-id,omitempty"`
	Provisioned_ bool   `yaml:"provisioned"`
	Size_        uint64 `yaml:"size"`
	Pool_        string `yaml:"pool,omitempty"`
	HardwareID_  string `yaml:"hardware-id,omitempty"`
	WWN_         string `yaml:"wwn,omitempty"`
	VolumeID_    string `yaml:"volume-id,omitempty"`
	Persistent_  bool   `yaml:"persistent"`

	Status_        *status `yaml:"status"`
	StatusHistory_ `yaml:"status-history"`

	Attachments_     volumeAttachments     `yaml:"attachments"`
	AttachmentPlans_ volumeAttachmentPlans `yaml:"attachmentplans"`
}

type volumeAttachments struct {
	Version      int                 `yaml:"version"`
	Attachments_ []*volumeAttachment `yaml:"attachments"`
}

type volumeAttachmentPlans struct {
	Version          int                     `yaml:"version"`
	AttachmentPlans_ []*volumeAttachmentPlan `yaml:"attachmentplans"`
}

type volumePlanInfo struct {
	DeviceType_       string            `yaml:"device-type,omitempty"`
	DeviceAttributes_ map[string]string `yaml:"device-attributes,omitempty"`
}

func (v volumePlanInfo) DeviceType() string {
	return v.DeviceType_
}

func (v volumePlanInfo) DeviceAttributes() map[string]string {
	return v.DeviceAttributes_
}

type volumeAttachment struct {
	HostID_         string         `yaml:"host-id"`
	Provisioned_    bool           `yaml:"provisioned"`
	ReadOnly_       bool           `yaml:"read-only"`
	DeviceName_     string         `yaml:"device-name"`
	DeviceLink_     string         `yaml:"device-link"`
	BusAddress_     string         `yaml:"bus-address"`
	VolumePlanInfo_ volumePlanInfo `yaml:"plan-info,omitempty"`
}

type volumeAttachmentPlan struct {
	MachineID_   string         `yaml:"machine-id"`
	BlockDevice_ *blockdevice   `yaml:"block-device,omitempty"`
	PlanInfo_    volumePlanInfo `yaml:"plan-info,omitempty"`
}

func (v volumeAttachmentPlan) Machine() names.MachineTag {
	return names.NewMachineTag(v.MachineID_)
}

func (v volumeAttachmentPlan) BlockDevice() BlockDevice {
	return v.BlockDevice_
}

func (v volumeAttachmentPlan) VolumePlanInfo() VolumePlanInfo {
	return v.PlanInfo_
}

// VolumeArgs is an argument struct used to add a volume to the Model.
type VolumeArgs struct {
	Tag         names.VolumeTag
	Storage     names.StorageTag
	Provisioned bool
	Size        uint64
	Pool        string
	HardwareID  string
	WWN         string
	VolumeID    string
	Persistent  bool
}

func newVolume(args VolumeArgs) *volume {
	v := &volume{
		ID_:            args.Tag.Id(),
		StorageID_:     args.Storage.Id(),
		Provisioned_:   args.Provisioned,
		Size_:          args.Size,
		Pool_:          args.Pool,
		HardwareID_:    args.HardwareID,
		WWN_:           args.WWN,
		VolumeID_:      args.VolumeID,
		Persistent_:    args.Persistent,
		StatusHistory_: newStatusHistory(),
	}
	v.setAttachments(nil)
	v.setAttachmentPlans(nil)
	return v
}

// Tag implements Volume.
func (v *volume) Tag() names.VolumeTag {
	return names.NewVolumeTag(v.ID_)
}

// Storage implements Volume.
func (v *volume) Storage() names.StorageTag {
	if v.StorageID_ == "" {
		return names.StorageTag{}
	}
	return names.NewStorageTag(v.StorageID_)
}

// Provisioned implements Volume.
func (v *volume) Provisioned() bool {
	return v.Provisioned_
}

// Size implements Volume.
func (v *volume) Size() uint64 {
	return v.Size_
}

// Pool implements Volume.
func (v *volume) Pool() string {
	return v.Pool_
}

// HardwareID implements Volume.
func (v *volume) HardwareID() string {
	return v.HardwareID_
}

// WWN implements Volume.
func (v *volume) WWN() string {
	return v.WWN_
}

// VolumeID implements Volume.
func (v *volume) VolumeID() string {
	return v.VolumeID_
}

// Persistent implements Volume.
func (v *volume) Persistent() bool {
	return v.Persistent_
}

// Status implements Volume.
func (v *volume) Status() Status {
	// To avoid typed nils check nil here.
	if v.Status_ == nil {
		return nil
	}
	return v.Status_
}

// SetStatus implements Volume.
func (v *volume) SetStatus(args StatusArgs) {
	v.Status_ = newStatus(args)
}

func (v *volume) setAttachments(attachments []*volumeAttachment) {
	v.Attachments_ = volumeAttachments{
		Version:      2,
		Attachments_: attachments,
	}
}

func (v *volume) setAttachmentPlans(attachmentPlans []*volumeAttachmentPlan) {
	v.AttachmentPlans_ = volumeAttachmentPlans{
		Version:          1,
		AttachmentPlans_: attachmentPlans,
	}
}

// Attachments implements Volume.
func (v *volume) Attachments() []VolumeAttachment {
	var result []VolumeAttachment
	for _, attachment := range v.Attachments_.Attachments_ {
		result = append(result, attachment)
	}
	return result
}

// AttachmentPlans implements Volume.
func (v *volume) AttachmentPlans() []VolumeAttachmentPlan {
	var result []VolumeAttachmentPlan
	for _, attachment := range v.AttachmentPlans_.AttachmentPlans_ {
		result = append(result, attachment)
	}
	return result
}

// AddAttachment implements Volume.
func (v *volume) AddAttachment(args VolumeAttachmentArgs) VolumeAttachment {
	a := newVolumeAttachment(args)
	v.Attachments_.Attachments_ = append(v.Attachments_.Attachments_, a)
	return a
}

// AddAttachmentPlan implements Volume.
func (v *volume) AddAttachmentPlan(args VolumeAttachmentPlanArgs) VolumeAttachmentPlan {
	a := newVolumeAttachmentPlan(args)
	v.AttachmentPlans_.AttachmentPlans_ = append(v.AttachmentPlans_.AttachmentPlans_, a)
	return a
}

// Validate implements Volume.
func (v *volume) Validate() error {
	if v.ID_ == "" {
		return errors.NotValidf("volume missing id")
	}
	if v.Size_ == 0 {
		return errors.NotValidf("volume %q missing size", v.ID_)
	}
	if v.Status_ == nil {
		return errors.NotValidf("volume %q missing status", v.ID_)
	}
	return nil
}

func importVolumes(source map[string]interface{}) ([]*volume, error) {
	checker := versionedChecker("volumes")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "volumes version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := volumeDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["volumes"].([]interface{})
	return importVolumeList(sourceList, importFunc)
}

func importVolumeList(sourceList []interface{}, importFunc volumeDeserializationFunc) ([]*volume, error) {
	result := make([]*volume, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for volume %d, %T", i, value)
		}
		volume, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "volume %d", i)
		}
		result = append(result, volume)
	}
	return result, nil
}

type volumeDeserializationFunc func(map[string]interface{}) (*volume, error)

var volumeDeserializationFuncs = map[int]volumeDeserializationFunc{
	1: importVolumeV1,
}

func importVolumeV1(source map[string]interface{}) (*volume, error) {
	fields := schema.Fields{
		"id":              schema.String(),
		"storage-id":      schema.String(),
		"provisioned":     schema.Bool(),
		"size":            schema.ForceUint(),
		"pool":            schema.String(),
		"hardware-id":     schema.String(),
		"wwn":             schema.String(),
		"volume-id":       schema.String(),
		"persistent":      schema.Bool(),
		"status":          schema.StringMap(schema.Any()),
		"attachments":     schema.StringMap(schema.Any()),
		"attachmentplans": schema.StringMap(schema.Any()),
	}

	defaults := schema.Defaults{
		"storage-id":      "",
		"pool":            "",
		"hardware-id":     "",
		"wwn":             "",
		"volume-id":       "",
		"attachments":     schema.Omit,
		"attachmentplans": schema.Omit,
	}
	addStatusHistorySchema(fields)
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "volume v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &volume{
		ID_:            valid["id"].(string),
		StorageID_:     valid["storage-id"].(string),
		Provisioned_:   valid["provisioned"].(bool),
		Size_:          valid["size"].(uint64),
		Pool_:          valid["pool"].(string),
		HardwareID_:    valid["hardware-id"].(string),
		WWN_:           valid["wwn"].(string),
		VolumeID_:      valid["volume-id"].(string),
		Persistent_:    valid["persistent"].(bool),
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

	attachments, err := importVolumeAttachments(valid["attachments"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}

	attachmentPlans, err := importVolumeAttachmentPlans(
		valid["attachmentplans"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.setAttachments(attachments)
	result.setAttachmentPlans(attachmentPlans)

	return result, nil
}

// VolumeAttachmentArgs is an argument struct used to add information about the
// cloud instance to a Volume.
type VolumeAttachmentArgs struct {
	Host        names.Tag
	Provisioned bool
	ReadOnly    bool
	DeviceName  string
	DeviceLink  string
	BusAddress  string

	DeviceType       string
	DeviceAttributes map[string]string
}

// VolumeAttachmentPlanArgs is an argument struct used to add information about
// a volume attached to an instance.
type VolumeAttachmentPlanArgs struct {
	Machine names.MachineTag

	DeviceName     string
	DeviceLinks    []string
	Label          string
	UUID           string
	HardwareId     string
	WWN            string
	BusAddress     string
	Size           uint64
	FilesystemType string
	InUse          bool
	MountPoint     string

	DeviceType       string
	DeviceAttributes map[string]string
}

func newVolumeAttachmentPlan(args VolumeAttachmentPlanArgs) *volumeAttachmentPlan {
	blockDevice := &blockdevice{
		Name_:           args.DeviceName,
		Links_:          args.DeviceLinks,
		Label_:          args.Label,
		UUID_:           args.UUID,
		HardwareID_:     args.HardwareId,
		WWN_:            args.WWN,
		BusAddress_:     args.BusAddress,
		Size_:           args.Size,
		FilesystemType_: args.FilesystemType,
		InUse_:          args.InUse,
		MountPoint_:     args.MountPoint,
	}
	planInfo := volumePlanInfo{
		DeviceType_:       args.DeviceType,
		DeviceAttributes_: args.DeviceAttributes,
	}
	return &volumeAttachmentPlan{
		MachineID_:   args.Machine.Id(),
		BlockDevice_: blockDevice,
		PlanInfo_:    planInfo,
	}
}

func newVolumeAttachment(args VolumeAttachmentArgs) *volumeAttachment {
	planInfo := volumePlanInfo{}
	if args.DeviceType != "" && args.DeviceAttributes != nil {
		planInfo.DeviceType_ = args.DeviceType
		planInfo.DeviceAttributes_ = args.DeviceAttributes
	}
	return &volumeAttachment{
		HostID_:         args.Host.Id(),
		Provisioned_:    args.Provisioned,
		ReadOnly_:       args.ReadOnly,
		DeviceName_:     args.DeviceName,
		DeviceLink_:     args.DeviceLink,
		BusAddress_:     args.BusAddress,
		VolumePlanInfo_: planInfo,
	}
}

func storageAttachmentHost(id string) names.Tag {
	if names.IsValidUnit(id) {
		return names.NewUnitTag(id)
	}
	return names.NewMachineTag(id)
}

// Host implements VolumeAttachment
func (a *volumeAttachment) Host() names.Tag {
	return storageAttachmentHost(a.HostID_)
}

// Provisioned implements VolumeAttachment
func (a *volumeAttachment) Provisioned() bool {
	return a.Provisioned_
}

// ReadOnly implements VolumeAttachment
func (a *volumeAttachment) ReadOnly() bool {
	return a.ReadOnly_
}

// DeviceName implements VolumeAttachment
func (a *volumeAttachment) DeviceName() string {
	return a.DeviceName_
}

// DeviceLink implements VolumeAttachment
func (a *volumeAttachment) DeviceLink() string {
	return a.DeviceLink_
}

// BusAddress implements VolumeAttachment
func (a *volumeAttachment) BusAddress() string {
	return a.BusAddress_
}

// VolumePlanInfo implements VolumeAttachment.
func (v *volumeAttachment) VolumePlanInfo() VolumePlanInfo {
	return v.VolumePlanInfo_
}

func importVolumeAttachments(source map[string]interface{}) ([]*volumeAttachment, error) {
	checker := versionedChecker("attachments")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "volume attachments version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := volumeAttachmentDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["attachments"].([]interface{})
	return importVolumeAttachmentList(sourceList, importFunc)
}

func importVolumeAttachmentList(sourceList []interface{}, importFunc volumeAttachmentDeserializationFunc) ([]*volumeAttachment, error) {
	result := make([]*volumeAttachment, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for volumeAttachment %d, %T", i, value)
		}
		volumeAttachment, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "volumeAttachment %d", i)
		}
		result = append(result, volumeAttachment)
	}
	return result, nil
}

type volumeAttachmentDeserializationFunc func(map[string]interface{}) (*volumeAttachment, error)

var volumeAttachmentDeserializationFuncs = map[int]volumeAttachmentDeserializationFunc{
	1: importVolumeAttachmentV1,
	2: importVolumeAttachmentV2,
}

func importVolumeAttachmentV1(source map[string]interface{}) (*volumeAttachment, error) {
	fields, defaults := volumeAttachmentV1Fields()
	return importVolumeAttachment(fields, defaults, 1, source)
}

func importVolumeAttachmentV2(source map[string]interface{}) (*volumeAttachment, error) {
	fields, defaults := volumeAttachmentV2Fields()
	return importVolumeAttachment(fields, defaults, 2, source)
}

func volumeAttachmentV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"machine-id":  schema.String(),
		"provisioned": schema.Bool(),
		"read-only":   schema.Bool(),
		"device-name": schema.String(),
		"device-link": schema.String(),
		"bus-address": schema.String(),
		"plan-info":   schema.StringMap(schema.Any()),
	}

	defaults := schema.Defaults{
		"plan-info": schema.Omit,
	}
	return fields, defaults
}

func volumeAttachmentV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := volumeAttachmentV1Fields()
	fields["host-id"] = schema.String()
	delete(fields, "machine-id")
	return fields, defaults
}

func importVolumeAttachment(fields schema.Fields, defaults schema.Defaults, importVersion int, source map[string]interface{}) (*volumeAttachment, error) {
	checker := schema.FieldMap(fields, defaults) // no defaults

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "volumeAttachment v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	var planInfo volumePlanInfo

	if valid["plan-info"] != nil {
		planInfo, err = importVolumePlanInfo(valid["plan-info"].(map[string]interface{}))
		if err != nil {
			return nil, errors.Annotatef(err, "volumeAttachmentPlanInfo schema check failed")
		}
	}

	result := &volumeAttachment{
		Provisioned_:    valid["provisioned"].(bool),
		ReadOnly_:       valid["read-only"].(bool),
		DeviceName_:     valid["device-name"].(string),
		DeviceLink_:     valid["device-link"].(string),
		BusAddress_:     valid["bus-address"].(string),
		VolumePlanInfo_: planInfo,
	}

	if importVersion >= 2 {
		result.HostID_ = valid["host-id"].(string)
	} else {
		result.HostID_ = valid["machine-id"].(string)
	}

	return result, nil
}

func importVolumeAttachmentPlans(source map[string]interface{}) ([]*volumeAttachmentPlan, error) {
	checker := versionedChecker("attachmentplans")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "volume attachment plans version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := volumeAttachmentPlanDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["attachmentplans"].([]interface{})
	return importVolumeAttachmentPlanList(sourceList, importFunc)
}

func importVolumeAttachmentPlanList(sourceList []interface{}, importFunc volumeAttachmentPlanDeserializationFunc) ([]*volumeAttachmentPlan, error) {
	result := make([]*volumeAttachmentPlan, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for volumeAttachmentPlan %d, %T", i, value)
		}
		volumeAttachmentPlan, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "volumeAttachmentPlan %d", i)
		}
		result = append(result, volumeAttachmentPlan)
	}
	return result, nil
}

type volumeAttachmentPlanDeserializationFunc func(map[string]interface{}) (*volumeAttachmentPlan, error)

var volumeAttachmentPlanDeserializationFuncs = map[int]volumeAttachmentPlanDeserializationFunc{
	1: importVolumeAttachmentPlanV1,
}

func coerceMapInterfacerToMapString(value map[string]interface{}) map[string]string {
	newMap := map[string]string{}

	for k, val := range value {
		newMap[k] = val.(string)
	}

	return newMap
}

func importVolumePlanInfo(source map[string]interface{}) (volumePlanInfo, error) {
	fields := schema.Fields{
		"device-type":       schema.String(),
		"device-attributes": schema.StringMap(schema.String()),
	}

	defaults := schema.Defaults{
		"device-attributes": schema.Omit,
	}

	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return volumePlanInfo{}, errors.Annotatef(err, "volumePlanInfo schema check failed")
	}
	valid := coerced.(map[string]interface{})

	planInfo := volumePlanInfo{
		DeviceType_: valid["device-type"].(string),
	}
	if valid["device-attributes"] != nil {
		planInfo.DeviceAttributes_ = coerceMapInterfacerToMapString(valid["device-attributes"].(map[string]interface{}))
	}

	return planInfo, nil
}

func importVolumeAttachmentPlanV1(source map[string]interface{}) (*volumeAttachmentPlan, error) {
	fields := schema.Fields{
		"machine-id":   schema.String(),
		"block-device": schema.StringMap(schema.Any()),
		"plan-info":    schema.StringMap(schema.Any()),
	}

	defaults := schema.Defaults{
		"block-device": schema.Omit,
		"plan-info":    schema.Omit,
	}

	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "volumeAttachment v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	planInfo, err := importVolumePlanInfo(valid["plan-info"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Annotatef(err, "volumeAttachmentPlanInfo schema check failed")
	}

	blockDeviceInfo, err := importBlockDeviceV1(valid["block-device"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Annotatef(err, "block devices version schema check failed")
	}

	result := &volumeAttachmentPlan{
		MachineID_:   valid["machine-id"].(string),
		BlockDevice_: blockDeviceInfo,
		PlanInfo_:    planInfo,
	}
	return result, nil
}
