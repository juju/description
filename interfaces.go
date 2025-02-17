// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/names/v5"
	"github.com/juju/version/v2"
)

// AgentTools represent the version and related binary file
// that the machine and unit agents are using.
type AgentTools interface {
	Version() version.Binary
	URL() string
	SHA256() string
	Size() int64
}

// Space represents a network space, which is a named collection of subnets.
type Space interface {
	Id() string
	Name() string
	Public() bool
	ProviderID() string
}

// LinkLayerDevice represents a link layer device.
type LinkLayerDevice interface {
	Name() string
	MTU() uint
	ProviderID() string
	MachineID() string
	Type() string
	MACAddress() string
	IsAutoStart() bool
	IsUp() bool
	ParentName() string
	VirtualPortType() string
}

// IPAddress represents an IP address.
type IPAddress interface {
	ProviderID() string
	DeviceName() string
	MachineID() string
	SubnetCIDR() string
	ConfigMethod() string
	Value() string
	DNSServers() []string
	DNSSearchDomains() []string
	GatewayAddress() string
	IsDefaultGateway() bool
	ProviderNetworkID() string
	ProviderSubnetID() string
	Origin() string
	IsShadow() bool
	IsSecondary() bool
}

// SSHHostKey represents an ssh host key.
type SSHHostKey interface {
	MachineID() string
	Keys() []string
}

// CloudImageMetadata represents an IP cloudimagemetadata.
type CloudImageMetadata interface {
	Stream() string
	Region() string
	Version() string
	Arch() string
	VirtType() string
	RootStorageType() string
	RootStorageSize() (uint64, bool)
	DateCreated() int64
	Source() string
	Priority() int
	ImageId() string
	ExpireAt() *time.Time
}

// Volume represents a volume (disk, logical volume, etc.) in the model.
type Volume interface {
	HasStatus
	HasStatusHistory

	Tag() names.VolumeTag
	Storage() names.StorageTag

	Provisioned() bool

	Size() uint64
	Pool() string

	HardwareID() string
	WWN() string
	VolumeID() string
	Persistent() bool

	Attachments() []VolumeAttachment
	AttachmentPlans() []VolumeAttachmentPlan
	AddAttachment(VolumeAttachmentArgs) VolumeAttachment
	AddAttachmentPlan(VolumeAttachmentPlanArgs) VolumeAttachmentPlan
}

// VolumeAttachment represents a volume attached to a machine.
type VolumeAttachment interface {
	Host() names.Tag
	Provisioned() bool
	ReadOnly() bool
	DeviceName() string
	DeviceLink() string
	BusAddress() string
	VolumePlanInfo() VolumePlanInfo
}

type VolumeAttachmentPlan interface {
	Machine() names.MachineTag
	BlockDevice() BlockDevice
	VolumePlanInfo() VolumePlanInfo
}

type VolumePlanInfo interface {
	DeviceType() string
	DeviceAttributes() map[string]string
}

// Filesystem represents a filesystem in the model.
type Filesystem interface {
	HasStatus
	HasStatusHistory

	Tag() names.FilesystemTag
	Volume() names.VolumeTag
	Storage() names.StorageTag

	Provisioned() bool

	Size() uint64
	Pool() string

	FilesystemID() string

	Attachments() []FilesystemAttachment
	AddAttachment(FilesystemAttachmentArgs) FilesystemAttachment
}

// FilesystemAttachment represents a filesystem attached to a machine.
type FilesystemAttachment interface {
	Host() names.Tag
	Provisioned() bool
	MountPoint() string
	ReadOnly() bool
}

// Storage represents the state of a unit or application-wide storage instance
// in the model.
type Storage interface {
	Tag() names.StorageTag
	Kind() string
	// Owner returns the tag of the application or unit that owns this storage
	// instance.
	Owner() (names.Tag, error)
	Name() string

	Attachments() []names.UnitTag

	// Constraints returns the storage instance constraints, and a boolean
	// reporting whether there are any.
	Constraints() (StorageInstanceConstraints, bool)

	Validate() error
}

// StoragePool represents a named storage pool and its settings.
type StoragePool interface {
	Name() string
	Provider() string
	Attributes() map[string]interface{}
}

// StorageDirective represents the user-specified storage directive for
// provisioning storage instances for an application unit.
type StorageDirective interface {
	// Pool is the name of the storage pool from which to provision the
	// storage instances.
	Pool() string
	// Size is the required size of the storage instances, in MiB.
	Size() uint64
	// Count is the required number of storage instances.
	Count() uint64
}

// StorageInstanceConstraints represents the user-specified constraints
// for provisioning a single storage instance for an application unit.
type StorageInstanceConstraints struct {
	Pool string
	Size uint64
}

// Subnet represents a network subnet.
type Subnet interface {
	ID() string
	ProviderId() string
	ProviderNetworkId() string
	ProviderSpaceId() string
	CIDR() string
	VLANTag() int
	AvailabilityZones() []string
	IsPublic() bool
	SpaceID() string
	SpaceName() string
	FanLocalUnderlay() string
	FanOverlay() string
}

// FirewallRule represents a firewall ruleset for a known service type, with
// whitelist CIDRs.
type FirewallRule interface {
	ID() string
	WellKnownService() string
	WhitelistCIDRs() []string
}

// CharmOrigin represents a charm source, where the charm originates from to
// help support multiple store locations.
type CharmOrigin interface {
	Source() string
	ID() string
	Hash() string
	Revision() int
	Channel() string
	Platform() string
}

// CharmMetadata represents the metadata of a charm.
type CharmMetadata interface {
	Name() string
	Summary() string
	Description() string
	Subordinate() bool
	MinJujuVersion() string
	RunAs() string
	Assumes() string
	Provides() map[string]CharmMetadataRelation
	Requires() map[string]CharmMetadataRelation
	Peers() map[string]CharmMetadataRelation
	ExtraBindings() map[string]string
	Categories() []string
	Tags() []string
	Storage() map[string]CharmMetadataStorage
	Devices() map[string]CharmMetadataDevice
	Payloads() map[string]CharmMetadataPayload
	Resources() map[string]CharmMetadataResource
	Terms() []string
	Containers() map[string]CharmMetadataContainer
}

// CharmMetadataRelation represents a relation in the metadata of a charm.
type CharmMetadataRelation interface {
	Name() string
	Role() string
	Interface() string
	Optional() bool
	Limit() int
	Scope() string
}

// CharmMetadataStorage represents a storage in the metadata of a charm.
type CharmMetadataStorage interface {
	Name() string
	Description() string
	Type() string
	Shared() bool
	Readonly() bool
	CountMin() int
	CountMax() int
	MinimumSize() int
	Location() string
	Properties() []string
}

// CharmMetadataDevice represents a device in the metadata of a charm.
type CharmMetadataDevice interface {
	Name() string
	Description() string
	Type() string
	CountMin() int
	CountMax() int
}

// CharmMetadataPayload represents a payload in the metadata of a charm.
type CharmMetadataPayload interface {
	Name() string
	Type() string
}

// CharmMetadataResource represents a resource in the metadata of a charm.
type CharmMetadataResource interface {
	Name() string
	Type() string
	Path() string
	Description() string
}

// CharmMetadataContainer represents a container in the metadata of a charm.
type CharmMetadataContainer interface {
	Resource() string
	Mounts() []CharmMetadataContainerMount
	Uid() *int
	Gid() *int
}

// CharmMetadataContainerMount represents a mount in the metadata of a charm
// container.
type CharmMetadataContainerMount interface {
	Storage() string
	Location() string
}

// CharmManifest represents the manifest of a charm.
type CharmManifest interface {
	Bases() []CharmManifestBase
}

// CharmManifestBase represents the metadata of a charm base.
type CharmManifestBase interface {
	Name() string
	Channel() string
	Architectures() []string
}

// CharmActions represents the actions of a charm.
type CharmActions interface {
	Actions() map[string]CharmAction
}

// CharmAction represents an action in the metadata of a charm.
type CharmAction interface {
	Description() string
	Parallel() bool
	ExecutionGroup() string
	Parameters() map[string]interface{}
}

// CharmConfigs represents the configuration of a charm.
type CharmConfigs interface {
	Configs() map[string]CharmConfig
}

// CharmConfig represents the configuration of a charm.
type CharmConfig interface {
	Type() string
	Default() interface{}
	Description() string
}

// CharmConfig represents a virtual host key of a unit/machine.
type VirtualHostKey interface {
	ID() string
	HostKey() []byte
}
