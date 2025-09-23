// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/names/v5"
	"github.com/juju/os/v2/series"
	"github.com/juju/schema"
	"github.com/juju/version/v2"
)

// Machine represents an existing live machine or container running in the
// model.
type Machine interface {
	HasAnnotations
	HasConstraints
	HasStatus
	HasStatusHistory

	Id() string
	Tag() names.MachineTag
	Nonce() string
	PasswordHash() string
	Placement() string
	Base() string
	ContainerType() string
	Jobs() []string
	SupportedContainers() ([]string, bool)

	Instance() CloudInstance
	SetInstance(CloudInstanceArgs)

	// Life() string -- only transmit alive things?
	ProviderAddresses() []Address
	MachineAddresses() []Address
	SetAddresses(machine []AddressArgs, provider []AddressArgs)

	PreferredPublicAddress() Address
	PreferredPrivateAddress() Address
	SetPreferredAddresses(public AddressArgs, private AddressArgs)

	Tools() AgentTools
	SetTools(AgentToolsArgs)

	Containers() []Machine
	AddContainer(MachineArgs) Machine

	BlockDevices() []BlockDevice
	AddBlockDevice(BlockDeviceArgs) BlockDevice

	OpenedPortRanges() PortRanges
	AddOpenedPortRange(OpenedPortRangeArgs)

	Validate() error

	Hostname() string
	SetHostname(hostname string)
}

type machines struct {
	Version   int        `yaml:"version"`
	Machines_ []*machine `yaml:"machines"`
}

type machine struct {
	Id_           string         `yaml:"id"`
	Nonce_        string         `yaml:"nonce"`
	PasswordHash_ string         `yaml:"password-hash"`
	Placement_    string         `yaml:"placement,omitempty"`
	Instance_     *cloudInstance `yaml:"instance,omitempty"`
	// Series obsolete from v3. Retained for tests.
	Series_        string `yaml:"series,omitempty"`
	Base_          string `yaml:"base"`
	ContainerType_ string `yaml:"container-type,omitempty"`

	Status_        *status `yaml:"status"`
	StatusHistory_ `yaml:"status-history"`

	ProviderAddresses_ []*address `yaml:"provider-addresses,omitempty"`
	MachineAddresses_  []*address `yaml:"machine-addresses,omitempty"`

	PreferredPublicAddress_  *address `yaml:"preferred-public-address,omitempty"`
	PreferredPrivateAddress_ *address `yaml:"preferred-private-address,omitempty"`

	Tools_ *agentTools `yaml:"tools"`
	Jobs_  []string    `yaml:"jobs"`

	SupportedContainers_ *[]string `yaml:"supported-containers,omitempty"`

	Containers_ []*machine `yaml:"containers"`

	OpenedPortRanges_ *deployedPortRanges `yaml:"opened-port-ranges,omitempty"`

	Annotations_ `yaml:"annotations,omitempty"`

	Constraints_ *constraints `yaml:"constraints,omitempty"`

	BlockDevices_ blockdevices `yaml:"block-devices,omitempty"`

	Hostname_ string `yaml:"hostname"`
}

// MachineArgs is an argument struct used to add a machine to the Model.
type MachineArgs struct {
	Id            names.MachineTag
	Nonce         string
	PasswordHash  string
	Placement     string
	Series        string
	Base          string
	ContainerType string
	Jobs          []string
	// A null value means that we don't yet know which containers
	// are supported. An empty slice means 'no containers are supported'.
	SupportedContainers *[]string
}

func newMachine(args MachineArgs) *machine {
	var jobs []string
	if count := len(args.Jobs); count > 0 {
		jobs = make([]string, count)
		copy(jobs, args.Jobs)
	}
	m := &machine{
		Id_:            args.Id.Id(),
		Nonce_:         args.Nonce,
		PasswordHash_:  args.PasswordHash,
		Placement_:     args.Placement,
		Series_:        args.Series,
		Base_:          args.Base,
		ContainerType_: args.ContainerType,
		Jobs_:          jobs,
		StatusHistory_: newStatusHistory(),
	}
	if args.SupportedContainers != nil {
		supported := make([]string, len(*args.SupportedContainers))
		copy(supported, *args.SupportedContainers)
		m.SupportedContainers_ = &supported
	}
	m.setBlockDevices(nil)
	return m
}

// Id implements Machine.
func (m *machine) Id() string {
	return m.Id_
}

// Tag implements Machine.
func (m *machine) Tag() names.MachineTag {
	return names.NewMachineTag(m.Id_)
}

// Nonce implements Machine.
func (m *machine) Nonce() string {
	return m.Nonce_
}

// PasswordHash implements Machine.
func (m *machine) PasswordHash() string {
	return m.PasswordHash_
}

// Placement implements Machine.
func (m *machine) Placement() string {
	return m.Placement_
}

// Instance implements Machine.
func (m *machine) Instance() CloudInstance {
	// To avoid typed nils check nil here.
	if m.Instance_ == nil {
		return nil
	}
	return m.Instance_
}

// SetInstance implements Machine.
func (m *machine) SetInstance(args CloudInstanceArgs) {
	m.Instance_ = newCloudInstance(args)
}

// Base implements Machine.
func (m *machine) Base() string {
	return m.Base_
}

// ContainerType implements Machine.
func (m *machine) ContainerType() string {
	return m.ContainerType_
}

// Status implements Machine.
func (m *machine) Status() Status {
	// To avoid typed nils check nil here.
	if m.Status_ == nil {
		return nil
	}
	return m.Status_
}

// SetStatus implements Machine.
func (m *machine) SetStatus(args StatusArgs) {
	m.Status_ = newStatus(args)
}

// ProviderAddresses implements Machine.
func (m *machine) ProviderAddresses() []Address {
	var result []Address
	for _, addr := range m.ProviderAddresses_ {
		result = append(result, addr)
	}
	return result
}

// MachineAddresses implements Machine.
func (m *machine) MachineAddresses() []Address {
	var result []Address
	for _, addr := range m.MachineAddresses_ {
		result = append(result, addr)
	}
	return result
}

// SetAddresses implements Machine.
func (m *machine) SetAddresses(margs []AddressArgs, pargs []AddressArgs) {
	m.MachineAddresses_ = nil
	m.ProviderAddresses_ = nil
	for _, args := range margs {
		if args.Value != "" {
			m.MachineAddresses_ = append(m.MachineAddresses_, newAddress(args))
		}
	}
	for _, args := range pargs {
		if args.Value != "" {
			m.ProviderAddresses_ = append(m.ProviderAddresses_, newAddress(args))
		}
	}
}

// PreferredPublicAddress implements Machine.
func (m *machine) PreferredPublicAddress() Address {
	// To avoid typed nils check nil here.
	if m.PreferredPublicAddress_ == nil {
		return nil
	}
	return m.PreferredPublicAddress_
}

// PreferredPrivateAddress implements Machine.
func (m *machine) PreferredPrivateAddress() Address {
	// To avoid typed nils check nil here.
	if m.PreferredPrivateAddress_ == nil {
		return nil
	}
	return m.PreferredPrivateAddress_
}

// SetPreferredAddresses implements Machine.
func (m *machine) SetPreferredAddresses(public AddressArgs, private AddressArgs) {
	if public.Value != "" {
		m.PreferredPublicAddress_ = newAddress(public)
	}
	if private.Value != "" {
		m.PreferredPrivateAddress_ = newAddress(private)
	}
}

// Tools implements Machine.
func (m *machine) Tools() AgentTools {
	// To avoid a typed nil, check before returning.
	if m.Tools_ == nil {
		return nil
	}
	return m.Tools_
}

// SetTools implements Machine.
func (m *machine) SetTools(args AgentToolsArgs) {
	m.Tools_ = newAgentTools(args)
}

// Jobs implements Machine.
func (m *machine) Jobs() []string {
	return m.Jobs_
}

// SupportedContainers implements Machine.
func (m *machine) SupportedContainers() ([]string, bool) {
	if m.SupportedContainers_ == nil {
		return nil, false
	}
	return *m.SupportedContainers_, true
}

// Containers implements Machine.
func (m *machine) Containers() []Machine {
	var result []Machine
	for _, container := range m.Containers_ {
		result = append(result, container)
	}
	return result
}

// BlockDevices implements Machine.
func (m *machine) BlockDevices() []BlockDevice {
	var result []BlockDevice
	for _, device := range m.BlockDevices_.BlockDevices_ {
		result = append(result, device)
	}
	return result
}

// AddBlockDevice implements Machine.
func (m *machine) AddBlockDevice(args BlockDeviceArgs) BlockDevice {
	return m.BlockDevices_.add(args)
}

func (m *machine) setBlockDevices(devices []*blockdevice) {
	m.BlockDevices_ = blockdevices{
		Version:       2,
		BlockDevices_: devices,
	}
}

// AddContainer implements Machine.
func (m *machine) AddContainer(args MachineArgs) Machine {
	container := newMachine(args)
	m.Containers_ = append(m.Containers_, container)
	return container
}

// OpenedPortRanges implements Machine.
func (m *machine) OpenedPortRanges() PortRanges {
	if m.OpenedPortRanges_ == nil {
		m.OpenedPortRanges_ = newDeployedPortRanges()
	}
	return m.OpenedPortRanges_
}

// AddOpenedPortRange implements Machine.
func (m *machine) AddOpenedPortRange(args OpenedPortRangeArgs) {
	if m.OpenedPortRanges_ == nil {
		m.OpenedPortRanges_ = newDeployedPortRanges()
	}

	if m.OpenedPortRanges_.ByUnit_[args.UnitName] == nil {
		m.OpenedPortRanges_.ByUnit_[args.UnitName] = newUnitPortRanges()
	}

	m.OpenedPortRanges_.ByUnit_[args.UnitName].ByEndpoint_[args.EndpointName] = append(
		m.OpenedPortRanges_.ByUnit_[args.UnitName].ByEndpoint_[args.EndpointName],
		newUnitPortRange(args.FromPort, args.ToPort, args.Protocol),
	)
}

// Constraints implements HasConstraints.
func (m *machine) Constraints() Constraints {
	if m.Constraints_ == nil {
		return nil
	}
	return m.Constraints_
}

// SetConstraints implements HasConstraints.
func (m *machine) SetConstraints(args ConstraintsArgs) {
	m.Constraints_ = newConstraints(args)
}

// Validate implements Machine.
func (m *machine) Validate() error {
	if m.Id_ == "" {
		return errors.NotValidf("machine missing id")
	}
	if m.Base_ != "" {
		parts := strings.Split(m.Base_, "@")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return errors.NotValidf("machine %q base %q", m.Id_, m.Base_)
		}
	}
	if m.Status_ == nil {
		return errors.NotValidf("machine %q missing status", m.Id_)
	}
	// Since all exports should be done when machines are stable,
	// there should always be tools and cloud instance.
	if m.Tools_ == nil {
		return errors.NotValidf("machine %q missing tools", m.Id_)
	}
	if m.Instance_ == nil {
		return errors.NotValidf("machine %q missing instance", m.Id_)
	}
	if err := m.Instance_.Validate(); err != nil {
		return errors.Annotatef(err, "machine %q instance", m.Id_)
	}
	for _, container := range m.Containers_ {
		if err := container.Validate(); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// SetHostname implements Machine.
func (m *machine) SetHostname(hostname string) {
	m.Hostname_ = hostname
}

// Hostname implements Machine.
func (m *machine) Hostname() string {
	return m.Hostname_
}

func importMachines(source map[string]interface{}) ([]*machine, error) {
	checker := versionedChecker("machines")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "machines version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := machineDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["machines"].([]interface{})
	return importMachineList(sourceList, importFunc)
}

func importMachineList(sourceList []interface{}, importFunc machineDeserializationFunc) ([]*machine, error) {
	result := make([]*machine, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for machine %d, %T", i, value)
		}
		machine, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "machine %d", i)
		}
		result = append(result, machine)
	}
	return result, nil
}

type machineDeserializationFunc func(map[string]interface{}) (*machine, error)

var machineDeserializationFuncs = map[int]machineDeserializationFunc{
	1: importMachineV1,
	2: importMachineV2,
	3: importMachineV3,
	4: importMachineV4,
}

func importMachineV1(source map[string]interface{}) (*machine, error) {
	fields, defaults := machineSchemaV1()
	return importMachine(fields, defaults, 1, source, importMachineV1)
}

func importMachineV2(source map[string]interface{}) (*machine, error) {
	fields, defaults := machineSchemaV2()
	return importMachine(fields, defaults, 2, source, importMachineV2)
}

func importMachineV3(source map[string]interface{}) (*machine, error) {
	fields, defaults := machineSchemaV3()
	return importMachine(fields, defaults, 3, source, importMachineV3)
}

func importMachineV4(source map[string]interface{}) (*machine, error) {
	fields, defaults := machineSchemaV4()
	return importMachine(fields, defaults, 4, source, importMachineV4)
}

func importMachine(
	fields schema.Fields, defaults schema.Defaults, importVersion int, source map[string]interface{},
	importFunc machineDeserializationFunc,
) (*machine, error) {
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "machine v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &machine{
		Id_:            valid["id"].(string),
		Nonce_:         valid["nonce"].(string),
		PasswordHash_:  valid["password-hash"].(string),
		Placement_:     valid["placement"].(string),
		ContainerType_: valid["container-type"].(string),
		StatusHistory_: newStatusHistory(),
		Jobs_:          convertToStringSlice(valid["jobs"]),
	}
	if importVersion < 3 {
		mSeries := valid["series"].(string)
		os, err := series.GetOSFromSeries(mSeries)
		if err != nil {
			return nil, errors.NotValidf("base series %q", mSeries)
		}
		vers, err := series.SeriesVersion(mSeries)
		if err != nil {
			return nil, errors.NotValidf("base series %q", mSeries)
		}
		result.Base_ = fmt.Sprintf("%s@%s", strings.ToLower(os.String()), vers)
	} else {
		result.Base_ = valid["base"].(string)
	}

	result.importAnnotations(valid)
	if err := result.importStatusHistory(valid); err != nil {
		return nil, errors.Trace(err)
	}

	if constraintsMap, ok := valid["constraints"]; ok {
		constraints, err := importConstraints(constraintsMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.Constraints_ = constraints
	}

	if supported, ok := valid["supported-containers"]; ok {
		supportedList := supported.([]interface{})
		s := make([]string, len(supportedList))
		for i, containerType := range supportedList {
			s[i] = containerType.(string)
		}
		result.SupportedContainers_ = &s
	}

	if instanceMap, ok := valid["instance"]; ok {
		instance, err := importCloudInstance(instanceMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.Instance_ = instance
	}

	if blockDeviceMap, ok := valid["block-devices"]; ok {
		devices, err := importBlockDevices(blockDeviceMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.setBlockDevices(devices)
	} else {
		result.setBlockDevices(nil)
	}

	// Tools and status are required, so we expect them to be there.
	tools, err := importAgentTools(valid["tools"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.Tools_ = tools

	status, err := importStatus(valid["status"].(map[string]interface{}))
	if err != nil {
		return nil, errors.Trace(err)
	}
	result.Status_ = status

	if addresses, ok := valid["provider-addresses"]; ok {
		providerAddresses, err := importAddresses(addresses.([]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.ProviderAddresses_ = providerAddresses
	}

	if addresses, ok := valid["machine-addresses"]; ok {
		machineAddresses, err := importAddresses(addresses.([]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.MachineAddresses_ = machineAddresses
	}

	if address, ok := valid["preferred-public-address"]; ok {
		publicAddress, err := importAddress(address.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.PreferredPublicAddress_ = publicAddress
	}

	if address, ok := valid["preferred-private-address"]; ok {
		privateAddress, err := importAddress(address.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}
		result.PreferredPrivateAddress_ = privateAddress
	}

	machineList := valid["containers"].([]interface{})
	machines, err := importMachineList(machineList, importFunc)
	if err != nil {
		return nil, errors.Annotatef(err, "containers")
	}
	result.Containers_ = machines

	if portsMap, ok := valid["opened-ports"]; ok {
		portsList, err := importOpenedPorts(portsMap.(map[string]interface{}))
		if err != nil {
			return nil, errors.Trace(err)
		}

		// Convert to the new open port ranges format. As subnets
		// were never actively used by older controllers (ranges were
		// opened to all subnets), we can emulate the original semantics
		// by opening each range for all endpoints.
		for _, portList := range portsList {
			for _, pr := range portList.OpenPorts() {
				result.AddOpenedPortRange(OpenedPortRangeArgs{
					UnitName:     pr.UnitName(),
					EndpointName: "",
					FromPort:     pr.FromPort(),
					ToPort:       pr.ToPort(),
					Protocol:     pr.Protocol(),
				})
			}
		}
	}

	if importVersion >= 2 {
		machPortRangesSource, ok := valid["opened-port-ranges"].(map[string]interface{})
		if ok {
			// V1 legacy ports have been converted to open port ranges by the V1
			// machine parser. V2 machines *only* include port ranges so if present
			// we can simply parse them and inject them into the parsed machine
			// instance.
			machPortRanges, err := importMachinePortRanges(machPortRangesSource)
			if err != nil {
				return nil, errors.Trace(err)
			}
			result.OpenedPortRanges_ = machPortRanges
		}
	}

	if importVersion >= 4 {
		hostname, ok := valid["hostname"].(string)
		if ok {
			result.Hostname_ = hostname
		}
	}

	return result, nil
}

func machineSchemaV1() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":                   schema.String(),
		"nonce":                schema.String(),
		"password-hash":        schema.String(),
		"placement":            schema.String(),
		"instance":             schema.StringMap(schema.Any()),
		"series":               schema.String(),
		"container-type":       schema.String(),
		"jobs":                 schema.List(schema.String()),
		"status":               schema.StringMap(schema.Any()),
		"supported-containers": schema.List(schema.String()),
		"tools":                schema.StringMap(schema.Any()),
		"containers":           schema.List(schema.StringMap(schema.Any())),
		"opened-ports":         schema.StringMap(schema.Any()),

		"provider-addresses":        schema.List(schema.StringMap(schema.Any())),
		"machine-addresses":         schema.List(schema.StringMap(schema.Any())),
		"preferred-public-address":  schema.StringMap(schema.Any()),
		"preferred-private-address": schema.StringMap(schema.Any()),

		"block-devices": schema.StringMap(schema.Any()),
	}

	defaults := schema.Defaults{
		"placement":      "",
		"container-type": "",
		// Even though we are expecting instance data for every machine,
		// it isn't strictly necessary, so we allow it to not exist here.
		"instance":                  schema.Omit,
		"supported-containers":      schema.Omit,
		"opened-ports":              schema.Omit,
		"block-devices":             schema.Omit,
		"provider-addresses":        schema.Omit,
		"machine-addresses":         schema.Omit,
		"preferred-public-address":  schema.Omit,
		"preferred-private-address": schema.Omit,
	}

	addAnnotationSchema(fields, defaults)
	addConstraintsSchema(fields, defaults)
	addStatusHistorySchema(fields)

	return fields, defaults
}

func machineSchemaV2() (schema.Fields, schema.Defaults) {
	fields, defaults := machineSchemaV1()

	fields["opened-port-ranges"] = schema.StringMap(schema.Any())
	defaults["opened-port-ranges"] = schema.Omit

	return fields, defaults
}

func machineSchemaV3() (schema.Fields, schema.Defaults) {
	fields, defaults := machineSchemaV2()

	fields["base"] = schema.String()
	delete(fields, "series")
	delete(defaults, "schema")

	return fields, defaults
}

func machineSchemaV4() (schema.Fields, schema.Defaults) {
	fields, defaults := machineSchemaV3()
	fields["hostname"] = schema.String()
	return fields, defaults
}

// AgentToolsArgs is an argument struct used to add information about the
// tools the agent is using to a Machine.
type AgentToolsArgs struct {
	Version version.Binary
	URL     string
	SHA256  string
	Size    int64
}

func newAgentTools(args AgentToolsArgs) *agentTools {
	return &agentTools{
		Version_:      2,
		ToolsVersion_: args.Version,
		URL_:          args.URL,
		SHA256_:       args.SHA256,
		Size_:         args.Size,
	}
}

// Keeping the agentTools with the machine code, because we hope
// that one day we will succeed in merging the unit agents with the
// machine agents.
type agentTools struct {
	Version_      int            `yaml:"version"`
	ToolsVersion_ version.Binary `yaml:"tools-version"`
	URL_          string         `yaml:"url"`
	SHA256_       string         `yaml:"sha256"`
	Size_         int64          `yaml:"size"`
}

// Version implements AgentTools.
func (a *agentTools) Version() version.Binary {
	return a.ToolsVersion_
}

// URL implements AgentTools.
func (a *agentTools) URL() string {
	return a.URL_
}

// SHA256 implements AgentTools.
func (a *agentTools) SHA256() string {
	return a.SHA256_
}

// Size implements AgentTools.
func (a *agentTools) Size() int64 {
	return a.Size_
}

func importAgentTools(source map[string]interface{}) (*agentTools, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "agentTools version schema check failed")
	}

	importFunc, ok := agentToolsDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type agentToolsDeserializationFunc func(map[string]interface{}) (*agentTools, error)

var agentToolsDeserializationFuncs = map[int]agentToolsDeserializationFunc{
	1: importAgentToolsV1,
	2: importAgentToolsV2,
}

func importAgentToolsV2(source map[string]interface{}) (*agentTools, error) {
	// v2 is the same format as v1 except that agent binary version string
	// contains the os name not the series name.
	return importAgentToolsVersion(source, 2)
}

func importAgentToolsV1(source map[string]interface{}) (*agentTools, error) {
	return importAgentToolsVersion(source, 1)
}

func importAgentToolsVersion(source map[string]interface{}, formatVersion int) (*agentTools, error) {
	fields := schema.Fields{
		"tools-version": schema.String(),
		"url":           schema.String(),
		"sha256":        schema.String(),
		"size":          schema.Int(),
	}
	checker := schema.FieldMap(fields, nil) // no defaults

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "agentTools v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	verString := valid["tools-version"].(string)
	toolsVersion, err := version.ParseBinary(verString)
	if err != nil {
		return nil, errors.Annotatef(err, "agentTools tools-version")
	}

	return &agentTools{
		Version_:      formatVersion,
		ToolsVersion_: toolsVersion,
		URL_:          valid["url"].(string),
		SHA256_:       valid["sha256"].(string),
		Size_:         valid["size"].(int64),
	}, nil
}
