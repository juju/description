// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/juju/names/v6"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type ModelSerializationSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&ModelSerializationSuite{})

func (*ModelSerializationSuite) TestNil(c *gc.C) {
	_, err := importModel(nil)
	c.Check(err, gc.ErrorMatches, "version: expected int, got nothing")
}

func (*ModelSerializationSuite) TestMissingVersion(c *gc.C) {
	_, err := importModel(map[string]interface{}{})
	c.Check(err, gc.ErrorMatches, "version: expected int, got nothing")
}

func (*ModelSerializationSuite) TestNonIntVersion(c *gc.C) {
	_, err := importModel(map[string]interface{}{
		"version": "hello",
	})
	c.Check(err.Error(), gc.Equals, `version: expected int, got string("hello")`)
}

func (*ModelSerializationSuite) TestUnknownVersion(c *gc.C) {
	_, err := importModel(map[string]interface{}{
		"version": 42,
	})
	c.Check(err.Error(), gc.Equals, `version 42 not valid`)
}

func (*ModelSerializationSuite) TestUpdateConfig(c *gc.C) {
	model := NewModel(ModelArgs{
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	model.UpdateConfig(map[string]interface{}{
		"name": "something else",
		"key":  "value",
	})
	c.Assert(model.Config(), jc.DeepEquals, map[string]interface{}{
		"name": "something else",
		"uuid": "some-uuid",
		"key":  "value",
	})
}

func (*ModelSerializationSuite) TestType(c *gc.C) {
	model := NewModel(ModelArgs{Type: "faas"})
	c.Check(model.Type(), gc.Equals, "faas")
}

func (*ModelSerializationSuite) TestCloudCredentials(c *gc.C) {
	owner := "me"
	model := NewModel(ModelArgs{
		Owner: owner,
	})
	args := CloudCredentialArgs{
		Owner:    owner,
		Cloud:    "altostratus",
		Name:     "creds",
		AuthType: "fuzzy",
		Attributes: map[string]string{
			"key": "value",
		},
	}
	model.SetCloudCredential(args)
	creds := model.CloudCredential()

	c.Check(creds.Owner(), gc.Equals, args.Owner)
	c.Check(creds.Cloud(), gc.Equals, args.Cloud)
	c.Check(creds.Name(), gc.Equals, args.Name)
	c.Check(creds.AuthType(), gc.Equals, args.AuthType)
	c.Check(creds.Attributes(), jc.DeepEquals, args.Attributes)
}

func (s *ModelSerializationSuite) exportImport(c *gc.C, initial Model) Model {
	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)
	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	return model
}

func (s *ModelSerializationSuite) TestVersions(c *gc.C) {
	args := ModelArgs{
		Type:  IAAS,
		Owner: "magic",
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		LatestToolsVersion: "2.0.1",
		EnvironVersion:     123,
		Blocks: map[string]string{
			"all-changes": "locked down",
		},
		Cloud:       "vapour",
		CloudRegion: "east-west",
	}
	initial := NewModel(args).(*model)
	c.Assert(initial.Applications_.Version, gc.Equals, 14)
	c.Assert(initial.Actions_.Version, gc.Equals, 4)
	c.Assert(initial.Operations_.Version, gc.Equals, 2)
	c.Assert(initial.Secrets_.Version, gc.Equals, 2)
	c.Assert(initial.Filesystems_.Version, gc.Equals, len(filesystemDeserializationFuncs))
	c.Assert(initial.Relations_.Version, gc.Equals, len(relationFieldsFuncs))
	c.Assert(initial.RemoteEntities_.Version, gc.Equals, len(remoteEntityFieldsFuncs))
	c.Assert(initial.RemoteApplications_.Version, gc.Equals, len(remoteApplicationFieldsFuncs))
	c.Assert(initial.Spaces_.Version, gc.Equals, len(spaceFieldsFuncs))
	c.Assert(initial.Volumes_.Version, gc.Equals, len(volumeDeserializationFuncs))
	c.Assert(initial.FirewallRules_.Version, gc.Equals, len(firewallRuleFieldsFuncs))
	c.Assert(initial.OfferConnections_.Version, gc.Equals, len(offerConnectionDeserializationFuncs))
	c.Assert(initial.ExternalControllers_.Version, gc.Equals, len(externalControllerDeserializationFuncs))
}

func (s *ModelSerializationSuite) TestSetBlocks(c *gc.C) {
	args := ModelArgs{
		Type:  IAAS,
		Owner: "magic",
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		LatestToolsVersion: "2.0.1",
		EnvironVersion:     123,
		Blocks: map[string]string{
			"all-changes": "locked down",
		},
		Cloud:       "vapour",
		CloudRegion: "east-west",
	}
	initial := NewModel(args).(*model)

	initial.SetBlocks(map[string]string{
		"all-changes": "unlocked",
		"some-other":  "value",
	})

	c.Assert(initial.Blocks(), jc.DeepEquals, map[string]string{
		"all-changes": "unlocked",
		"some-other":  "value",
	})
}

func (s *ModelSerializationSuite) TestParsingYAML(c *gc.C) {
	s.testParsingYAMLWithMachine(c, func(initial Model) {
		addMinimalMachine(initial, "0")
	})
}

func (s *ModelSerializationSuite) TestParsingYAMLWithMissingModificationStatus(c *gc.C) {
	s.testParsingYAMLWithMachine(c, func(initial Model) {
		addMinimalMachineWithMissingModificationStatus(initial, "0")
	})
}

func (s *ModelSerializationSuite) testParsingYAMLWithMachine(c *gc.C, machineFn func(Model)) {
	args := ModelArgs{
		AgentVersion: "3.1.1",
		Type:         IAAS,
		Owner:        "magic",
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
		LatestToolsVersion: "2.0.1",
		EnvironVersion:     123,
		Blocks: map[string]string{
			"all-changes": "locked down",
		},
		Cloud:       "vapour",
		CloudRegion: "east-west",
	}
	initial := NewModel(args)
	initial.SetCloudCredential(CloudCredentialArgs{
		Name:  "creds",
		Cloud: "vapour",
		Owner: "admin",
	})
	adminUser := "admin"
	initial.AddUser(UserArgs{
		Name:        adminUser,
		CreatedBy:   adminUser,
		DateCreated: time.Date(2015, 10, 9, 12, 34, 56, 0, time.UTC),
	})
	machineFn(initial)
	addMinimalApplication(initial)
	model := s.exportImport(c, initial)

	c.Check(model.AgentVersion(), gc.Equals, "3.1.1")
	c.Assert(model.Type(), gc.Equals, IAAS)
	c.Assert(model.Owner(), gc.Equals, args.Owner)
	c.Assert(model.UUID(), gc.Equals, "some-uuid")
	c.Assert(model.Config(), jc.DeepEquals, args.Config)
	c.Assert(model.Cloud(), gc.Equals, "vapour")
	c.Assert(model.CloudRegion(), gc.Equals, "east-west")
	c.Assert(model.CloudCredential(), jc.DeepEquals, initial.CloudCredential())
	c.Assert(model.LatestToolsVersion(), gc.Equals, args.LatestToolsVersion)
	c.Assert(model.EnvironVersion(), gc.Equals, args.EnvironVersion)
	c.Assert(model.Blocks(), jc.DeepEquals, args.Blocks)
	users := model.Users()
	c.Assert(users, gc.HasLen, 1)
	c.Assert(users[0].Name(), gc.Equals, adminUser)
	machines := model.Machines()
	c.Assert(machines, gc.HasLen, 1)
	c.Assert(machines[0].Id(), gc.Equals, "0")
	applications := model.Applications()
	c.Assert(applications, gc.HasLen, 1)
	c.Assert(applications[0].Name(), gc.Equals, "ubuntu")
}

func (s *ModelSerializationSuite) TestParsingOptionals(c *gc.C) {
	args := ModelArgs{
		Owner: "magic",
		Config: map[string]interface{}{
			"name": "awesome",
			"uuid": "some-uuid",
		},
	}
	initial := NewModel(args)
	model := s.exportImport(c, initial)
	c.Assert(model.LatestToolsVersion(), gc.Equals, "")
}

func (s *ModelSerializationSuite) TestAnnotations(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	annotations := map[string]string{
		"string":  "value",
		"another": "one",
	}
	initial.SetAnnotations(annotations)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Annotations(), jc.DeepEquals, annotations)
}

func (s *ModelSerializationSuite) TestSequences(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	initial.SetSequence("machine", 4)
	initial.SetSequence("application-foo", 3)
	initial.SetSequence("application-bar", 1)
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(model.Sequences(), jc.DeepEquals, map[string]int{
		"machine":         4,
		"application-foo": 3,
		"application-bar": 1,
	})
}

func (s *ModelSerializationSuite) TestConstraints(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	args := ConstraintsArgs{
		Architecture: "amd64",
		Memory:       8 * gig,
	}
	initial.SetConstraints(args)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Constraints(), jc.DeepEquals, newConstraints(args))
}

func (*ModelSerializationSuite) TestModelValidation(c *gc.C) {
	model := NewModel(ModelArgs{})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing model owner not valid")
	c.Assert(err, jc.ErrorIs, errors.NotValid)
}

func (*ModelSerializationSuite) TestModelValidationAgentVersion(c *gc.C) {
	model := NewModel(ModelArgs{
		Owner: "owner", CloudRegion: "some-region",
		AgentVersion: "1.2.3.4.5",
	})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `agent version "1.2.3.4.5" not valid`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksMachines(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner", CloudRegion: "some-region"})
	model.AddMachine(MachineArgs{})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "machine missing id not valid")
	c.Assert(err, jc.ErrorIs, errors.NotValid)
}

func (s *ModelSerializationSuite) addMachineToModel(model Model, id string) Machine {
	machine := model.AddMachine(MachineArgs{Id: id})
	machine.SetInstance(CloudInstanceArgs{InstanceId: "magic"})
	machine.SetTools(minimalAgentToolsArgs())
	machine.SetStatus(minimalStatusArgs())
	machine.Instance().SetStatus(minimalStatusArgs())
	machine.Instance().SetModificationStatus(minimalStatusArgs())
	return machine
}

func (s *ModelSerializationSuite) TestAddBlockDevices(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner", CloudRegion: "some-region"})
	model.AddMachine(MachineArgs{Id: "666"})
	err := model.AddBlockDevice("666", BlockDeviceArgs{
		Name:           "foo",
		Links:          []string{"a-link"},
		Label:          "label",
		UUID:           "device-uuid",
		HardwareID:     "hardware-id",
		WWN:            "wwn",
		BusAddress:     "bus-address",
		SerialID:       "serial-id",
		Size:           100,
		FilesystemType: "ext4",
		InUse:          true,
		MountPoint:     "/path/to/here",
	})
	c.Assert(err, jc.ErrorIsNil)
	m := model.Machines()
	c.Assert(m, gc.HasLen, 1)
	blockDevices := m[0].BlockDevices()
	c.Assert(blockDevices, gc.HasLen, 1)
	bd := blockDevices[0]
	c.Assert(bd.Name(), gc.Equals, "foo")
	c.Assert(bd.Links(), jc.DeepEquals, []string{"a-link"})
	c.Assert(bd.Label(), gc.Equals, "label")
	c.Assert(bd.UUID(), gc.Equals, "device-uuid")
	c.Assert(bd.HardwareID(), gc.Equals, "hardware-id")
	c.Assert(bd.WWN(), gc.Equals, "wwn")
	c.Assert(bd.BusAddress(), gc.Equals, "bus-address")
	c.Assert(bd.SerialID(), gc.Equals, "serial-id")
	c.Assert(bd.Size(), gc.Equals, uint64(100))
	c.Assert(bd.FilesystemType(), gc.Equals, "ext4")
	c.Assert(bd.InUse(), jc.IsTrue)
	c.Assert(bd.MountPoint(), gc.Equals, "/path/to/here")
}

func (s *ModelSerializationSuite) TestAddBlockDevicesMachineNotFound(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner", CloudRegion: "some-region"})
	err := model.AddBlockDevice("666", BlockDeviceArgs{
		Name: "foo",
	})
	c.Assert(err, jc.ErrorIs, errors.NotFound)
}

func (s *ModelSerializationSuite) TestModelValidationChecksMachinesGood(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner", CloudRegion: "some-region"})
	s.addMachineToModel(model, "0")
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksOpenPortsUnits(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner", CloudRegion: "some-region"})
	machine := s.addMachineToModel(model, "0")
	machine.AddOpenedPortRange(OpenedPortRangeArgs{
		UnitName:     "missing/0",
		EndpointName: "",
		FromPort:     8080,
		ToPort:       8080,
		Protocol:     "tcp",
	})
	err := model.Validate()
	c.Assert(err.Error(), gc.Equals, "unknown unit names in open ports: [missing/0]")
}

func (s *ModelSerializationSuite) TestModelValidationChecksApplications(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner", CloudRegion: "some-region"})
	model.AddApplication(ApplicationArgs{})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "application missing name not valid")
	c.Assert(err, jc.ErrorIs, errors.NotValid)
}

func (s *ModelSerializationSuite) addApplicationToModel(model Model, name string, numUnits int) Application {
	application := model.AddApplication(ApplicationArgs{
		Name:               name,
		CharmConfig:        map[string]interface{}{},
		LeadershipSettings: map[string]interface{}{},
	})
	application.SetStatus(minimalStatusArgs())
	for i := 0; i < numUnits; i++ {
		// The index i is used as both the machine id and the unit id.
		// A happy coincidence.
		machine := s.addMachineToModel(model, fmt.Sprint(i))
		unit := application.AddUnit(UnitArgs{
			Name:    fmt.Sprintf("%s/%d", name, i),
			Machine: machine.Id(),
		})
		unit.SetTools(minimalAgentToolsArgs())
		unit.SetAgentStatus(minimalStatusArgs())
		unit.SetWorkloadStatus(minimalStatusArgs())
	}

	return application
}

func (s *ModelSerializationSuite) wordpressModel() (Model, Endpoint, Endpoint) {
	model := NewModel(ModelArgs{
		Owner: "owner",
		Config: map[string]interface{}{
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	s.addApplicationToModel(model, "wordpress", 2)
	s.addApplicationToModel(model, "mysql", 1)

	// Add a relation between wordpress and mysql.
	rel := model.AddRelation(RelationArgs{
		Id:  42,
		Key: "special key",
	})
	rel.SetStatus(minimalStatusArgs())

	wordpressEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: "wordpress",
		Name:            "db",
		// Ignoring other aspects of endpoints.
	})
	mysqlEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: "mysql",
		Name:            "mysql",
		// Ignoring other aspects of endpoints.
	})
	return model, wordpressEndpoint, mysqlEndpoint
}

func (s *ModelSerializationSuite) wordpressModelWithSettings() Model {
	model, wordpressEndpoint, mysqlEndpoint := s.wordpressModel()

	s.setEndpointSettings(wordpressEndpoint, "wordpress/0", "wordpress/1")
	s.setEndpointSettings(mysqlEndpoint, "mysql/0")

	return model
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelationsMissingSettings(c *gc.C) {
	model, _, _ := s.wordpressModel()
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing relation settings for units \\[wordpress/0 wordpress/1\\] in relation 42")
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelationsMissingSettings2(c *gc.C) {
	model, wordpressEndpoint, _ := s.wordpressModel()

	s.setEndpointSettings(wordpressEndpoint, "wordpress/0", "wordpress/1")

	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "missing relation settings for units \\[mysql/0\\] in relation 42")
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelations(c *gc.C) {
	model := s.wordpressModelWithSettings()
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) addSubordinateEndpoints(c *gc.C, rel Relation, app string) (Endpoint, Endpoint) {
	appEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: app,
		Name:            "logging",
		Scope:           "container",
		// Ignoring other aspects of endpoints.
	})
	loggingEndpoint := rel.AddEndpoint(EndpointArgs{
		ApplicationName: "logging",
		Name:            "logging",
		Scope:           "container",
		// Ignoring other aspects of endpoints.
	})
	return appEndpoint, loggingEndpoint
}

func (s *ModelSerializationSuite) setEndpointSettings(ep Endpoint, units ...string) {
	for _, unit := range units {
		ep.SetUnitSettings(unit, map[string]interface{}{
			"key": "value",
		})
	}
}

func (s *ModelSerializationSuite) TestModelValidationChecksRelationsWithSubordinates(c *gc.C) {
	model := s.wordpressModelWithSettings()

	s.addApplicationToModel(model, "logging", 3)

	// Add a subordinate relations between logging and both wordpress and mysql.
	rel := model.AddRelation(RelationArgs{
		Id:  43,
		Key: "some key",
	})
	wordpressEndpoint, loggingEndpoint := s.addSubordinateEndpoints(c, rel, "wordpress")
	s.setEndpointSettings(wordpressEndpoint, "wordpress/0", "wordpress/1")
	s.setEndpointSettings(loggingEndpoint, "logging/0", "logging/1")

	rel = model.AddRelation(RelationArgs{
		Id:  44,
		Key: "other key",
	})
	mysqlEndpoint, loggingEndpoint := s.addSubordinateEndpoints(c, rel, "mysql")
	s.setEndpointSettings(mysqlEndpoint, "mysql/0")
	s.setEndpointSettings(loggingEndpoint, "logging/2")

	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelSerializationWithRelations(c *gc.C) {
	initial := s.wordpressModelWithSettings()
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, initial)
}

func (s *ModelSerializationSuite) TestModelSerializationWithRemoteEntities(c *gc.C) {
	model := NewModel(ModelArgs{
		Owner: "owner",
		Config: map[string]interface{}{
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	model.AddRemoteEntity(RemoteEntityArgs{
		Token: "xxx-aaa-bbb",
	})
	model.AddRemoteEntity(RemoteEntityArgs{
		Token:    "zzz-ccc-yyy",
		Macaroon: "some-macaroon-that-should-be-discharged",
	})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
	bytes, err := yaml.Marshal(model)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, model)
}

func (s *ModelSerializationSuite) TestModelSerializationWithRelationNetworks(c *gc.C) {
	model := NewModel(ModelArgs{
		Owner: "owner",
		Config: map[string]interface{}{
			"uuid": "some-uuid",
		},
		CloudRegion: "some-region",
	})
	model.AddRelationNetwork(RelationNetworkArgs{
		ID:          "ctrl-uuid-3",
		RelationKey: "relation-key",
		CIDRS:       []string{"10.0.1.0/16"},
	})
	model.AddRelationNetwork(RelationNetworkArgs{
		ID:          "ctrl-uuid-4",
		RelationKey: "relation-key",
		CIDRS:       []string{"12.0.1.1/24"},
	})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
	bytes, err := yaml.Marshal(model)
	c.Assert(err, jc.ErrorIsNil)
	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(result, jc.DeepEquals, model)
}

func (s *ModelSerializationSuite) TestModelValidationChecksSubnets(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddSubnet(SubnetArgs{CIDR: "10.0.0.0/24", SpaceUUID: "deadbeef-1bad-500d-9000-4b1d0d06f00d"})
	model.AddSubnet(SubnetArgs{CIDR: "10.0.1.0/24"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `subnet "10.0.0.0/24" references non-existent space "deadbeef-1bad-500d-9000-4b1d0d06f00d"`)
	model.AddSpace(SpaceArgs{UUID: "deadbeef-1bad-500d-9000-4b1d0d06f00d"})
	err = model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressMachineID(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddIPAddress(IPAddressArgs{Value: "192.168.1.0", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.0" references non-existent machine "42"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressDeviceName(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{Value: "192.168.1.0", MachineID: "42", DeviceName: "foo"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.0" references non-existent device "foo"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressValueEmpty(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{MachineID: "42", DeviceName: "foo"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address has invalid value ""`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressValueInvalid(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{MachineID: "42", DeviceName: "foo", Value: "foobar"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address has invalid value "foobar"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressSubnetEmpty(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{MachineID: "42", DeviceName: "foo", Value: "192.168.1.1"}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.1" has empty subnet CIDR`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressSubnetInvalid(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{
		MachineID:  "42",
		DeviceName: "foo",
		Value:      "192.168.1.1",
		SubnetCIDR: "foo",
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.1" has invalid subnet CIDR "foo"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressSucceeds(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{
		MachineID:  "42",
		DeviceName: "foo",
		Value:      "192.168.1.1",
		SubnetCIDR: "192.168.1.0/24",
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressGatewayAddressInvalid(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{
		MachineID:      "42",
		DeviceName:     "foo",
		Value:          "192.168.1.1",
		SubnetCIDR:     "192.168.1.0/24",
		GatewayAddress: "foo",
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ip address "192.168.1.1" has invalid gateway address "foo"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksAddressGatewayAddressValid(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := IPAddressArgs{
		MachineID:        "42",
		DeviceName:       "foo",
		Value:            "192.168.1.2",
		SubnetCIDR:       "192.168.1.0/24",
		GatewayAddress:   "192.168.1.1",
		IsDefaultGateway: true,
	}
	model.AddIPAddress(args)
	s.addMachineToModel(model, "42")
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksLinkLayerDeviceMachineId(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo", MachineID: "42"})
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" references non-existent machine "42"`)
	s.addMachineToModel(model, "42")
	err = model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksLinkLayerName(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{MachineID: "42"})
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, "device has empty name.*")
}

func (s *ModelSerializationSuite) TestModelValidationChecksLinkLayerMACAddress(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", MACAddress: "DEADBEEF"}
	model.AddLinkLayerDevice(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" has invalid MACAddress "DEADBEEF"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentExists(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "bar", MACAddress: "01:23:45:67:89:ab"}
	model.AddLinkLayerDevice(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" has non-existent parent "bar"`)
	model.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "bar", MachineID: "42"})
	err = model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentIsNotItself(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "foo"}
	model.AddLinkLayerDevice(args)
	s.addMachineToModel(model, "42")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" is its own parent`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentIsABridge(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar"}
	model.AddLinkLayerDevice(args2)
	s.addMachineToModel(model, "42")
	s.addMachineToModel(model, "43")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `device "foo" on a container but not a bridge`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksChildDeviceContained(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "42", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar", Type: "bridge"}
	model.AddLinkLayerDevice(args2)
	s.addMachineToModel(model, "42")
	s.addMachineToModel(model, "43")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `ParentName "m#43#d#bar" for non-container machine "42"`)
}

func (s *ModelSerializationSuite) TestModelValidationChecksParentOnHost(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "41/lxd/0", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar", Type: "bridge"}
	model.AddLinkLayerDevice(args2)
	machine := s.addMachineToModel(model, "41")
	container := machine.AddContainer(MachineArgs{Id: "41/lxd/0"})
	container.SetInstance(CloudInstanceArgs{InstanceId: "magic"})
	container.Instance().SetStatus(minimalStatusArgs())
	container.SetTools(minimalAgentToolsArgs())
	container.SetStatus(minimalStatusArgs())
	s.addMachineToModel(model, "43")
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `parent machine of device "foo" not host machine "41"`)
}

func (s *ModelSerializationSuite) TestModelValidationLinkLayerDeviceContainer(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	args := LinkLayerDeviceArgs{MachineID: "43/lxd/0", Name: "foo", ParentName: "m#43#d#bar"}
	model.AddLinkLayerDevice(args)
	args2 := LinkLayerDeviceArgs{MachineID: "43", Name: "bar", Type: "bridge"}
	model.AddLinkLayerDevice(args2)
	machine := s.addMachineToModel(model, "43")
	container := machine.AddContainer(MachineArgs{Id: "43/lxd/0"})
	container.SetInstance(CloudInstanceArgs{InstanceId: "magic"})
	container.Instance().SetStatus(minimalStatusArgs())
	container.SetTools(minimalAgentToolsArgs())
	container.SetStatus(minimalStatusArgs())
	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func (s *ModelSerializationSuite) TestNewModelSetsRemoteApplications(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "mogwai"})
	c.Assert(model.RemoteApplications(), gc.IsNil)
}

func (s *ModelSerializationSuite) TestModelValidationHandlesRemoteApplications(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "ink-spots"})
	remoteApp := model.AddRemoteApplication(RemoteApplicationArgs{
		Name:            "mysql",
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModelUUID: "some-model",
		IsConsumerProxy: true,
	})
	remoteApp.AddEndpoint(RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})

	s.addApplicationToModel(model, "wordpress", 1)
	rel := model.AddRelation(RelationArgs{
		Id:  101,
		Key: "wordpress:db mysql:db",
	})
	rel.AddEndpoint(EndpointArgs{
		ApplicationName: "wordpress",
		Name:            "db",
	})
	rel.AddEndpoint(EndpointArgs{
		ApplicationName: "mysql",
		Name:            "db",
	})

	err := model.Validate()
	c.Assert(err, jc.ErrorIsNil)
}

func asStringMap(c *gc.C, model Model) map[string]interface{} {
	bytes, err := Serialize(model)
	c.Assert(err, jc.ErrorIsNil)

	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	c.Assert(err, jc.ErrorIsNil)
	return data
}

func (s *ModelSerializationSuite) TestSerializesRemoteApplications(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	rapp := model.AddRemoteApplication(RemoteApplicationArgs{
		Name:            "bloom",
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModelUUID: "some-model",
		IsConsumerProxy: true,
		ConsumeVersion:  666,
	})
	rapp.AddEndpoint(RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})
	rapp.SetStatus(StatusArgs{
		Value:   "running",
		Updated: time.Date(2017, 5, 9, 12, 1, 0, 0, time.UTC),
	})
	data := asStringMap(c, model)
	remoteSection, ok := data["remote-applications"]
	c.Assert(ok, jc.IsTrue)

	// Re-serialize just that bit so we can check it.
	bytes, err := yaml.Marshal(remoteSection)
	c.Assert(err, jc.ErrorIsNil)

	expected := `
remote-applications:
- consume-version: 666
  endpoints:
    endpoints:
    - interface: mysql
      name: db
      role: provider
    version: 1
  is-consumer-proxy: true
  name: bloom
  offer-uuid: offer-uuid
  source-model-uuid: some-model
  spaces:
    spaces: []
    version: 1
  status:
    status:
      neverset: false
      updated: "2017-05-09T12:01:00Z"
      value: running
    version: 2
  url: other.mysql
version: 3
`[1:]
	c.Assert(string(bytes), gc.Equals, expected)
}

func (s *ModelSerializationSuite) TestImportingWithRemoteApplications(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "veils"})
	rapp := initial.AddRemoteApplication(RemoteApplicationArgs{
		Name:            "bloom",
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModelUUID: "some-model",
		IsConsumerProxy: true,
	})
	rapp.AddEndpoint(RemoteEndpointArgs{
		Name:      "db",
		Role:      "provider",
		Interface: "mysql",
	})
	rapp.SetStatus(StatusArgs{
		Value:   "hey",
		Updated: time.Now(),
	})
	remoteApplications := initial.RemoteApplications()

	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)

	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	ra := result.RemoteApplications()
	c.Assert(ra, gc.HasLen, 1)
	c.Assert(ra[0], gc.DeepEquals, remoteApplications[0])
}

func (s *ModelSerializationSuite) TestRemoteApplicationsGetter(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	model.AddRemoteApplication(RemoteApplicationArgs{
		Name:            "bloom",
		OfferUUID:       "offer-uuid",
		URL:             "other.mysql",
		SourceModelUUID: "some-model",
		IsConsumerProxy: true,
	})
	result := model.RemoteApplications()
	c.Assert(result, gc.HasLen, 1)
}

func (s *ModelSerializationSuite) TestSerializesOfferConnections(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	model.AddOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "some-model-uuid",
		UserName:        "fred",
	})
	data := asStringMap(c, model)
	offerSection, ok := data["offer-connections"]
	c.Assert(ok, jc.IsTrue)

	// Re-serialize just that bit so we can check it.
	bytes, err := yaml.Marshal(offerSection)
	c.Assert(err, jc.ErrorIsNil)

	expected := `
offer-connections:
- offer-uuid: offer-uuid
  relation-id: 1
  relation-key: relation-key
  source-model-uuid: some-model-uuid
  user-name: fred
version: 1
`[1:]
	c.Assert(string(bytes), gc.Equals, expected)
}

func (s *ModelSerializationSuite) TestImportingWithOfferConnections(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "veils"})
	initial.AddOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "some-model-uuid",
		UserName:        "fred",
	})
	offerConnections := initial.OfferConnections()

	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)

	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	offers := result.OfferConnections()
	c.Assert(offers, gc.HasLen, 1)
	c.Assert(offers[0], gc.DeepEquals, offerConnections[0])
}

func (s *ModelSerializationSuite) TestOfferConnectionsGetter(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	model.AddOfferConnection(OfferConnectionArgs{
		OfferUUID:       "offer-uuid",
		RelationID:      1,
		RelationKey:     "relation-key",
		SourceModelUUID: "some-model-uuid",
		UserName:        "fred",
	})
	result := model.OfferConnections()
	c.Assert(result, gc.HasLen, 1)
}

func (s *ModelSerializationSuite) TestSerializesExternalControllers(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	model.AddExternalController(ExternalControllerArgs{
		ID:     "name",
		Alias:  "moon-ball",
		Addrs:  []string{"1.2.3.4", "10.12.11.243"},
		CACert: "magic-cert",
		Models: []string{"aaaa-bbbb"},
	})
	data := asStringMap(c, model)
	ctrlSection, ok := data["external-controllers"]
	c.Assert(ok, jc.IsTrue)

	// Re-serialize just that bit so we can check it.
	bytes, err := yaml.Marshal(ctrlSection)
	c.Assert(err, jc.ErrorIsNil)

	expected := `
external-controllers:
- addrs:
  - 1.2.3.4
  - 10.12.11.243
  alias: moon-ball
  ca-cert: magic-cert
  id: name
  models:
  - aaaa-bbbb
version: 1
`[1:]
	c.Assert(string(bytes), gc.Equals, expected)
}

func (s *ModelSerializationSuite) TestImportingWithExternalControllers(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "veils"})
	initial.AddExternalController(ExternalControllerArgs{
		ID:     "name",
		Alias:  "moon-ball",
		Addrs:  []string{"1.2.3.4", "10.12.11.243"},
		CACert: "magic-cert",
		Models: []string{"aaaa-bbbb"},
	})
	offerConnections := initial.ExternalControllers()

	bytes, err := Serialize(initial)
	c.Assert(err, jc.ErrorIsNil)

	result, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	offers := result.ExternalControllers()
	c.Assert(offers, gc.HasLen, 1)
	c.Assert(offers[0], gc.DeepEquals, offerConnections[0])
}

func (s *ModelSerializationSuite) TestExternalControllersGetter(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	model.AddExternalController(ExternalControllerArgs{
		ID:     "name",
		Alias:  "moon-ball",
		Addrs:  []string{"1.2.3.4", "10.12.11.243"},
		CACert: "magic-cert",
	})
	result := model.ExternalControllers()
	c.Assert(result, gc.HasLen, 1)
}

func (s *ModelSerializationSuite) TestSetAndGetSLA(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	sla := model.SetSLA("essential", "bob", "creds")
	c.Assert(sla.Level(), gc.Equals, "essential")
	c.Assert(sla.Owner(), gc.Equals, "bob")
	c.Assert(sla.Credentials(), gc.Equals, "creds")

	getSla := model.SLA()
	c.Assert(getSla.Level(), gc.Equals, sla.Level())
	c.Assert(getSla.Credentials(), jc.DeepEquals, sla.Credentials())
}

func (s *ModelSerializationSuite) TestSLA(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	sla := initial.SetSLA("essential", "bob", "creds")
	c.Assert(sla.Level(), gc.Equals, "essential")
	c.Assert(sla.Owner(), gc.Equals, "bob")
	c.Assert(sla.Credentials(), gc.Equals, "creds")

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.SLA().Level(), gc.Equals, "essential")
	c.Assert(model.SLA().Owner(), gc.Equals, "bob")
	c.Assert(model.SLA().Credentials(), gc.Equals, "creds")
}

func (s *ModelSerializationSuite) TestGetAndSetMeterStatus(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "veils"})
	ms := model.SetMeterStatus("RED", "info message")
	c.Assert(ms.Code(), gc.Equals, "RED")
	c.Assert(ms.Info(), gc.Equals, "info message")

	getms := model.MeterStatus()
	c.Assert(getms.Code(), gc.Equals, ms.Code())
	c.Assert(getms.Info(), gc.Equals, ms.Info())
}

func (s *ModelSerializationSuite) TestMeterStatus(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	ms := initial.SetMeterStatus("RED", "info message")
	c.Assert(ms.Code(), gc.Equals, "RED")
	c.Assert(ms.Info(), gc.Equals, "info message")

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.MeterStatus().Code(), gc.Equals, "RED")
	c.Assert(model.MeterStatus().Info(), gc.Equals, "info message")
}

func (s *ModelSerializationSuite) TestPasswordHash(c *gc.C) {
	initial := NewModel(ModelArgs{PasswordHash: "some-hash"})
	c.Assert(initial.PasswordHash(), gc.Equals, "some-hash")

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.PasswordHash(), gc.Equals, "some-hash")
}

func (s *ModelSerializationSuite) TestSerializesToLatestVersion(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "ben-harper"})
	data := asStringMap(c, initial)
	versionValue, ok := data["version"]
	c.Assert(ok, jc.IsTrue)
	version, ok := versionValue.(int)
	c.Assert(ok, jc.IsTrue)
	c.Assert(version, gc.Equals, 14)
}

func (s *ModelSerializationSuite) TestSpaces(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	space := initial.AddSpace(SpaceArgs{UUID: "deadbeef-1bad-500d-9000-4b1d0d06f00d", Name: "special"})
	c.Assert(space.Name(), gc.Equals, "special")
	c.Assert(space.UUID(), gc.Equals, "deadbeef-1bad-500d-9000-4b1d0d06f00d")

	spaces := initial.Spaces()
	c.Assert(spaces, gc.HasLen, 1)
	c.Assert(spaces[0], gc.Equals, space)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Spaces(), jc.DeepEquals, spaces)

}

func (s *ModelSerializationSuite) TestLinkLayerDevice(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	device := initial.AddLinkLayerDevice(LinkLayerDeviceArgs{Name: "foo"})
	c.Assert(device.Name(), gc.Equals, "foo")
	devices := initial.LinkLayerDevices()
	c.Assert(devices, gc.HasLen, 1)
	c.Assert(devices[0], jc.DeepEquals, device)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	c.Logf(string(bytes))

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.LinkLayerDevices(), jc.DeepEquals, devices)
}

func (s *ModelSerializationSuite) TestSubnets(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	initial.AddSubnet(SubnetArgs{CIDR: "10.0.20.0/24", SpaceID: "0"})
	subnet := initial.AddSubnet(SubnetArgs{CIDR: "10.0.0.0/24"})
	c.Assert(subnet.CIDR(), gc.Equals, "10.0.0.0/24")
	subnets := initial.Subnets()
	c.Assert(subnets, gc.HasLen, 2)
	c.Assert(subnets[1], jc.DeepEquals, subnet)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Subnets(), jc.DeepEquals, subnets)
}

func (s *ModelSerializationSuite) TestIPAddress(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	addr := initial.AddIPAddress(IPAddressArgs{Value: "10.0.0.4"})
	c.Assert(addr.Value(), gc.Equals, "10.0.0.4")
	addresses := initial.IPAddresses()
	c.Assert(addresses, gc.HasLen, 1)
	c.Assert(addresses[0], jc.DeepEquals, addr)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.IPAddresses(), jc.DeepEquals, addresses)
}

func (s *ModelSerializationSuite) TestSSHHostKey(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	key := initial.AddSSHHostKey(SSHHostKeyArgs{MachineID: "foo"})
	c.Assert(key.MachineID(), gc.Equals, "foo")
	keys := initial.SSHHostKeys()
	c.Assert(keys, gc.HasLen, 1)
	c.Assert(keys[0], jc.DeepEquals, key)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.SSHHostKeys(), jc.DeepEquals, keys)
}

func (s *ModelSerializationSuite) TestCloudImageMetadata(c *gc.C) {
	storageSize := uint64(3)
	initial := NewModel(ModelArgs{Owner: "owner"})
	image := initial.AddCloudImageMetadata(CloudImageMetadataArgs{
		Stream:          "stream",
		Region:          "region-test",
		Version:         "14.04",
		Arch:            "arch",
		VirtType:        "virtType-test",
		RootStorageType: "rootStorageType-test",
		RootStorageSize: &storageSize,
		Source:          "test",
		Priority:        2,
		ImageId:         "1",
		DateCreated:     2,
	})
	c.Assert(image.Stream(), gc.Equals, "stream")
	c.Assert(image.Region(), gc.Equals, "region-test")
	c.Assert(image.Version(), gc.Equals, "14.04")
	c.Assert(image.Arch(), gc.Equals, "arch")
	c.Assert(image.VirtType(), gc.Equals, "virtType-test")
	c.Assert(image.RootStorageType(), gc.Equals, "rootStorageType-test")
	value, ok := image.RootStorageSize()
	c.Assert(ok, jc.IsTrue)
	c.Assert(value, gc.Equals, uint64(3))
	c.Assert(image.Source(), gc.Equals, "test")
	c.Assert(image.Priority(), gc.Equals, 2)
	c.Assert(image.ImageId(), gc.Equals, "1")
	c.Assert(image.DateCreated(), gc.Equals, int64(2))

	metadata := initial.CloudImageMetadata()
	c.Assert(metadata, gc.HasLen, 1)
	c.Assert(metadata[0], jc.DeepEquals, image)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.CloudImageMetadata(), jc.DeepEquals, metadata)
}

func (s *ModelSerializationSuite) TestAction(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	enqueued := time.Now().UTC()
	action := initial.AddAction(ActionArgs{
		Name:       "foo",
		Enqueued:   enqueued,
		Parameters: map[string]interface{}{},
		Results:    map[string]interface{}{},
	})
	c.Assert(action.Name(), gc.Equals, "foo")
	c.Assert(action.Enqueued(), gc.Equals, enqueued)
	actions := initial.Actions()
	c.Assert(actions, gc.HasLen, 1)
	c.Assert(actions[0], jc.DeepEquals, action)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Actions(), jc.DeepEquals, actions)
}

func (s *ModelSerializationSuite) TestOperation(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	enqueued := time.Now().UTC()
	op := initial.AddOperation(OperationArgs{
		Summary:           "foo",
		Fail:              "fail",
		Enqueued:          enqueued,
		CompleteTaskCount: 666,
	})
	c.Assert(op.Summary(), gc.Equals, "foo")
	c.Assert(op.Fail(), gc.Equals, "fail")
	c.Assert(op.CompleteTaskCount(), gc.Equals, 666)
	c.Assert(op.Enqueued(), gc.Equals, enqueued)
	operations := initial.Operations()
	c.Assert(operations, gc.HasLen, 1)
	c.Assert(operations[0], jc.DeepEquals, op)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Operations(), jc.DeepEquals, operations)
}

func (s *ModelSerializationSuite) TestVolumeValidation(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddVolume(testVolumeArgs())
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `volume\[0\]: volume "1234" missing status not valid`)
}

func (s *ModelSerializationSuite) TestVolumes(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	volume := initial.AddVolume(testVolumeArgs())
	volume.SetStatus(minimalStatusArgs())
	volumes := initial.Volumes()
	c.Assert(volumes, gc.HasLen, 1)
	c.Assert(volumes[0], gc.Equals, volume)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Volumes(), jc.DeepEquals, volumes)
}

func (s *ModelSerializationSuite) TestFilesystemValidation(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddFilesystem(testFilesystemArgs())
	err := model.Validate()
	c.Assert(err, gc.ErrorMatches, `filesystem\[0\]: filesystem "1234" missing status not valid`)
}

func (s *ModelSerializationSuite) TestFilesystems(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	filesystem := initial.AddFilesystem(testFilesystemArgs())
	filesystem.SetStatus(minimalStatusArgs())
	filesystem.AddAttachment(testFilesystemAttachmentArgs())
	filesystems := initial.Filesystems()
	c.Assert(filesystems, gc.HasLen, 1)
	c.Assert(filesystems[0], gc.Equals, filesystem)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Filesystems(), jc.DeepEquals, filesystems)
}

func (s *ModelSerializationSuite) TestFirewallRule(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	firewallRule := initial.AddFirewallRule(MinimalFireWallArgs())
	firewallRules := initial.FirewallRules()
	c.Assert(firewallRules, gc.HasLen, 1)
	c.Assert(firewallRules[0], jc.DeepEquals, firewallRule)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.FirewallRules(), jc.DeepEquals, firewallRules)
}

func (s *ModelSerializationSuite) TestStorage(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	storage := initial.AddStorage(testStorageArgs())
	storages := initial.Storages()
	c.Assert(storages, gc.HasLen, 1)
	c.Assert(storages[0], jc.DeepEquals, storage)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Storages(), jc.DeepEquals, storages)
}

func (s *ModelSerializationSuite) TestStorageValidate(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	storageArgs := testStorageArgs()
	storageArgs.UnitOwner = ""
	storage := initial.AddStorage(storageArgs)
	storages := initial.Storages()
	c.Assert(storages, gc.HasLen, 1)
	c.Assert(storages[0], jc.DeepEquals, storage)

	err := initial.Validate()
	c.Assert(err, gc.ErrorMatches, `storage\[0\] attachment referencing unknown unit "postgresql/0" not valid`)
}

func (s *ModelSerializationSuite) TestStoragePools(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	poolOne := map[string]interface{}{
		"foo":   42,
		"value": true,
	}
	poolTwo := map[string]interface{}{
		"value": "spanner",
	}
	initial.AddStoragePool(StoragePoolArgs{
		Name: "one", Provider: "sparkles", Attributes: poolOne})
	initial.AddStoragePool(StoragePoolArgs{
		Name: "two", Provider: "spanner", Attributes: poolTwo})

	pools := initial.StoragePools()
	c.Assert(pools, gc.HasLen, 2)
	one, two := pools[0], pools[1]
	c.Check(one.Name(), gc.Equals, "one")
	c.Check(one.Provider(), gc.Equals, "sparkles")
	c.Check(one.Attributes(), jc.DeepEquals, poolOne)
	c.Check(two.Name(), gc.Equals, "two")
	c.Check(two.Provider(), gc.Equals, "spanner")
	c.Check(two.Attributes(), jc.DeepEquals, poolTwo)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)

	pools = model.StoragePools()
	c.Assert(pools, gc.HasLen, 2)
	one, two = pools[0], pools[1]
	c.Check(one.Name(), gc.Equals, "one")
	c.Check(one.Provider(), gc.Equals, "sparkles")
	c.Check(one.Attributes(), jc.DeepEquals, poolOne)
	c.Check(two.Name(), gc.Equals, "two")
	c.Check(two.Provider(), gc.Equals, "spanner")
	c.Check(two.Attributes(), jc.DeepEquals, poolTwo)
}

func (s *ModelSerializationSuite) TestSecretBackend(c *gc.C) {
	initial := NewModel(ModelArgs{
		Owner:           "owner",
		SecretBackendID: "backend-id",
	})
	c.Assert(initial.SecretBackendID(), gc.Equals, "backend-id")

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.SecretBackendID(), jc.DeepEquals, "backend-id")
}

func (s *ModelSerializationSuite) TestSecrets(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	secretArgs := testSecretArgs()
	secret := initial.AddSecret(secretArgs)
	c.Assert(secret.Id(), gc.Equals, secretArgs.ID)
	c.Assert(secret.Version(), gc.Equals, 1)
	c.Assert(secret.Description(), gc.Equals, secretArgs.Description)
	c.Assert(secret.Label(), gc.Equals, secretArgs.Label)
	c.Assert(secret.RotatePolicy(), gc.Equals, secretArgs.RotatePolicy)
	c.Assert(secret.AutoPrune(), gc.Equals, secretArgs.AutoPrune)
	c.Assert(secret.LatestRevisionChecksum(), gc.DeepEquals, secretArgs.LatestRevisionChecksum)
	owner, err := secret.Owner()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(owner.String(), gc.Equals, secretArgs.Owner.String())
	c.Assert(secret.Created(), gc.Equals, secretArgs.Created)
	c.Assert(secret.Updated(), gc.Equals, secretArgs.Updated)
	c.Assert(secret.NextRotateTime(), jc.DeepEquals, secretArgs.NextRotateTime)
	secrets := initial.Secrets()
	c.Assert(secrets, gc.HasLen, 1)
	c.Assert(secrets[0], jc.DeepEquals, secret)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.Secrets(), jc.DeepEquals, secrets)
}

func (s *ModelSerializationSuite) TestSecretValidate(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	secretArgs := testSecretArgs()
	secretArgs.Owner = names.NewUnitTag("foo/0")
	secret := initial.AddSecret(secretArgs)
	secrets := initial.Secrets()
	c.Assert(secrets, gc.HasLen, 1)
	c.Assert(secrets[0], jc.DeepEquals, secret)

	err := initial.Validate()
	c.Assert(err, gc.ErrorMatches, `secret\[0\] owner \(foo/0\) not valid`)
}

func (s *ModelSerializationSuite) TestSecretValidateRemoteConsumer(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	addMinimalApplication(initial)
	initial.AddRemoteApplication(RemoteApplicationArgs{
		Name: "foo",
	})
	secretArgs := testSecretArgs()
	secretArgs.Owner = names.NewApplicationTag("ubuntu")
	secretArgs.Consumers[0].Consumer = names.NewApplicationTag("ubuntu")
	secretArgs.Consumers[1].Consumer = names.NewUnitTag("ubuntu/0")
	secretArgs.ACL = nil
	secretArgs.RemoteConsumers[0].Consumer = names.NewUnitTag("bar/0")
	secret := initial.AddSecret(secretArgs)
	secrets := initial.Secrets()
	c.Assert(secrets, gc.HasLen, 1)
	c.Assert(secrets[0], jc.DeepEquals, secret)

	err := initial.Validate()
	c.Assert(err, gc.ErrorMatches, `secret\[0\] remote consumer \(bar/0\) not valid`)
}

func (s *ModelSerializationSuite) TestRemoteSecrets(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	remoteSecretArgs := testRemoteSecretArgs()
	remoteSecret := initial.AddRemoteSecret(remoteSecretArgs)
	c.Assert(remoteSecret.ID(), gc.Equals, remoteSecretArgs.ID)
	c.Assert(remoteSecret.Label(), gc.Equals, remoteSecretArgs.Label)
	c.Assert(remoteSecret.CurrentRevision(), gc.Equals, remoteSecretArgs.CurrentRevision)
	c.Assert(remoteSecret.SourceUUID(), gc.Equals, remoteSecretArgs.SourceUUID)
	consumer, err := remoteSecret.Consumer()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(consumer.String(), gc.Equals, remoteSecretArgs.Consumer.String())
	remoteSecrets := initial.RemoteSecrets()
	c.Assert(remoteSecrets, gc.HasLen, 1)
	c.Assert(remoteSecrets[0], jc.DeepEquals, remoteSecret)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.RemoteSecrets(), jc.DeepEquals, remoteSecrets)
}

func (s *ModelSerializationSuite) TestRemoteSecretsValidate(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	remoteSecretArgs := testRemoteSecretArgs()
	remoteSecretArgs.Consumer = names.NewApplicationTag("foo")
	remoteSecret := initial.AddRemoteSecret(remoteSecretArgs)
	remoteSecrets := initial.RemoteSecrets()
	c.Assert(remoteSecrets, gc.HasLen, 1)
	c.Assert(remoteSecrets[0], jc.DeepEquals, remoteSecret)

	err := initial.Validate()
	c.Assert(err, gc.ErrorMatches, `remote secret\[0\] consumer \(foo\) not valid`)
}

func (s *ModelSerializationSuite) TestAgentVersionPre11Import(c *gc.C) {
	initial := NewModel(ModelArgs{
		Config: map[string]any{
			"agent-version": "3.3.3",
		},
	})
	data := asStringMap(c, initial)
	data["version"] = 10
	// NOTE (tlm): This is a little bit of a hack for v13 onwards. From v13 we
	// are know longer including status in the resultant model. However because
	// we are forcing the version back to 10 for this test it still needs to be
	//available.
	data["status"] = map[string]any{}
	bytes, err := yaml.Marshal(data)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Check(err, jc.ErrorIsNil)

	c.Check(model.AgentVersion(), gc.Equals, "3.3.3")
}

func (s *ModelSerializationSuite) TestSetEnvironVersion(c *gc.C) {
	model := NewModel(ModelArgs{
		EnvironVersion: 3,
	})
	c.Assert(model.EnvironVersion(), gc.Equals, 3)
	model.SetEnvironVersion(5)
	c.Assert(model.EnvironVersion(), gc.Equals, 5)
}

func (s *ModelSerializationSuite) TestVirtualHostKeys(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	virtualHostKeyArgs := testVirtualHostKeyArgs()
	virtualHostKey := initial.AddVirtualHostKey(virtualHostKeyArgs)
	c.Assert(virtualHostKey.ID(), gc.Equals, virtualHostKeyArgs.ID)
	c.Assert(virtualHostKey.HostKey(), gc.DeepEquals, virtualHostKeyArgs.HostKey)

	virtualHostKeys := initial.VirtualHostKeys()
	c.Assert(virtualHostKeys, gc.HasLen, 1)
	c.Assert(virtualHostKeys[0], jc.DeepEquals, virtualHostKey)

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	model, err := Deserialize(bytes)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(model.VirtualHostKeys(), jc.DeepEquals, virtualHostKeys)
}

func (s *ModelSerializationSuite) TestVirtualHostKeysValidate(c *gc.C) {
	initial := NewModel(ModelArgs{Owner: "owner"})
	virtualHostKeyArgs := testVirtualHostKeyArgs()
	virtualHostKeyArgs.ID = ""
	virtualHostKey := initial.AddVirtualHostKey(virtualHostKeyArgs)
	virtualHostKeys := initial.VirtualHostKeys()
	c.Assert(virtualHostKeys, gc.HasLen, 1)
	c.Assert(virtualHostKeys[0], jc.DeepEquals, virtualHostKey)

	err := initial.Validate()
	c.Assert(err, gc.ErrorMatches, `virtual host key\[0\]: empty id not valid`)
}

func (s *ModelSerializationSuite) TestSetOwner(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.SetOwner("bob")
	c.Assert(model.Owner(), gc.Equals, "bob")
}

func (s *ModelSerializationSuite) TestSetUsers(c *gc.C) {
	model := NewModel(ModelArgs{Owner: "owner"})
	model.AddUser(UserArgs{Name: "alice", DisplayName: "Alice", Access: "foo"})
	c.Assert(model.Users(), gc.HasLen, 1)
	c.Assert(model.Users()[0].Name(), gc.Equals, "alice")
	model.SetUsers(nil)
	c.Assert(model.Users(), gc.HasLen, 0)
}
