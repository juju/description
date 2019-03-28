// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/version"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/names.v2"
	"gopkg.in/yaml.v2"
)

type MachineSerializationSuite struct {
	SliceSerializationSuite
	PortRangeCheck
	StatusHistoryMixinSuite
}

var _ = gc.Suite(&MachineSerializationSuite{})

func (s *MachineSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "machines"
	s.sliceName = "machines"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importMachines(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["machines"] = []interface{}{}
	}
	s.StatusHistoryMixinSuite.creator = func() HasStatusHistory {
		return minimalMachine("1")
	}
	s.StatusHistoryMixinSuite.serializer = func(c *gc.C, initial interface{}) HasStatusHistory {
		return s.exportImport(c, initial.(*machine))
	}
}

func minimalMachineMap(id string, containers ...interface{}) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":             id,
		"nonce":          "a-nonce",
		"password-hash":  "some-hash",
		"instance":       minimalCloudInstanceMap(),
		"series":         "zesty",
		"tools":          minimalAgentToolsMap(),
		"jobs":           []interface{}{"host-units"},
		"containers":     containers,
		"status":         minimalStatusMap(),
		"status-history": emptyStatusHistoryMap(),
		"block-devices":  emptyBlockDeviceMap(),
	}
}

func minimalMachineMapWithPriorInstanceMap(id string, containers ...interface{}) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":             id,
		"nonce":          "a-nonce",
		"password-hash":  "some-hash",
		"instance":       minimalCloudInstanceMapV3(),
		"series":         "zesty",
		"tools":          minimalAgentToolsMap(),
		"jobs":           []interface{}{"host-units"},
		"containers":     containers,
		"status":         minimalStatusMap(),
		"status-history": emptyStatusHistoryMap(),
		"block-devices":  emptyBlockDeviceMap(),
	}
}

func minimalMachine(id string, containers ...*machine) *machine {
	m := newMachine(MachineArgs{
		Id:           names.NewMachineTag(id),
		Nonce:        "a-nonce",
		PasswordHash: "some-hash",
		Series:       "zesty",
		Jobs:         []string{"host-units"},
	})
	m.Containers_ = containers
	m.SetInstance(minimalCloudInstanceArgs())
	m.Instance().SetStatus(minimalStatusArgs())
	m.Instance().SetModificationStatus(minimalStatusArgs())
	m.SetTools(minimalAgentToolsArgs())
	m.SetStatus(minimalStatusArgs())
	return m
}

func minimalMachineWithPriorInstanceMap(id string, containers ...*machine) *machine {
	m := newMachine(MachineArgs{
		Id:           names.NewMachineTag(id),
		Nonce:        "a-nonce",
		PasswordHash: "some-hash",
		Series:       "zesty",
		Jobs:         []string{"host-units"},
	})
	m.Containers_ = containers
	m.SetInstance(minimalCloudInstanceArgs())
	// The new instance constructed by SetInstance will be the current
	// version, change it to 3 to match the version returned by
	// minimalMachineMapWithPriorInstanceMap.
	m.Instance_.Version = 3
	m.Instance().SetStatus(minimalStatusArgs())
	m.SetTools(minimalAgentToolsArgs())
	m.SetStatus(minimalStatusArgs())
	return m
}

func addMinimalMachine(model Model, id string) {
	m := model.AddMachine(MachineArgs{
		Id:           names.NewMachineTag(id),
		Nonce:        "a-nonce",
		PasswordHash: "some-hash",
		Series:       "zesty",
		Jobs:         []string{"host-units"},
	})
	m.SetInstance(minimalCloudInstanceArgs())
	m.SetTools(minimalAgentToolsArgs())
	m.SetStatus(minimalStatusArgs())
	m.Instance().SetStatus(minimalStatusArgs())
	m.Instance().SetModificationStatus(minimalStatusArgs())
}

func addMinimalMachineWithMissingModificationStatus(model Model, id string) {
	m := model.AddMachine(MachineArgs{
		Id:           names.NewMachineTag(id),
		Nonce:        "a-nonce",
		PasswordHash: "some-hash",
		Series:       "zesty",
		Jobs:         []string{"host-units"},
	})
	m.SetInstance(minimalCloudInstanceArgs())
	m.SetTools(minimalAgentToolsArgs())
	m.SetStatus(minimalStatusArgs())
	m.Instance().SetStatus(minimalStatusArgs())
}

func (s *MachineSerializationSuite) machineArgs(id string) MachineArgs {
	return MachineArgs{
		Id:            names.NewMachineTag(id),
		Nonce:         "a nonce",
		PasswordHash:  "some-hash",
		Placement:     "placement",
		Series:        "zesty",
		ContainerType: "magic",
		Jobs:          []string{"this", "that"},
	}
}

func (s *MachineSerializationSuite) TestNewMachine(c *gc.C) {
	m := newMachine(s.machineArgs("42"))
	c.Assert(m.Id(), gc.Equals, "42")
	c.Assert(m.Tag(), gc.Equals, names.NewMachineTag("42"))
	c.Assert(m.Nonce(), gc.Equals, "a nonce")
	c.Assert(m.PasswordHash(), gc.Equals, "some-hash")
	c.Assert(m.Placement(), gc.Equals, "placement")
	c.Assert(m.Series(), gc.Equals, "zesty")
	c.Assert(m.ContainerType(), gc.Equals, "magic")
	c.Assert(m.Jobs(), jc.DeepEquals, []string{"this", "that"})
	supportedContainers, ok := m.SupportedContainers()
	c.Assert(ok, jc.IsFalse)
	c.Assert(supportedContainers, gc.IsNil)
}

func (s *MachineSerializationSuite) TestMinimalMachineValid(c *gc.C) {
	m := minimalMachine("1")
	c.Assert(m.Validate(), jc.ErrorIsNil)
}

func (s *MachineSerializationSuite) TestValidateMissingID(c *gc.C) {
	m := newMachine(MachineArgs{})
	err := m.Validate()
	c.Check(err, jc.Satisfies, errors.IsNotValid)
	c.Check(err, gc.ErrorMatches, "machine missing id not valid")
}

func (s *MachineSerializationSuite) TestValidateMissingStatus(c *gc.C) {
	m := newMachine(s.machineArgs("42"))
	err := m.Validate()
	c.Check(err, jc.Satisfies, errors.IsNotValid)
	c.Check(err, gc.ErrorMatches, `machine "42" missing status not valid`)
}

func (s *MachineSerializationSuite) TestValidateMissingTools(c *gc.C) {
	m := newMachine(s.machineArgs("42"))
	m.SetStatus(minimalStatusArgs())
	err := m.Validate()
	c.Check(err, jc.Satisfies, errors.IsNotValid)
	c.Check(err, gc.ErrorMatches, `machine "42" missing tools not valid`)
}

func (s *MachineSerializationSuite) TestValidateMissingInstance(c *gc.C) {
	m := newMachine(s.machineArgs("42"))
	m.SetStatus(minimalStatusArgs())
	m.SetTools(minimalAgentToolsArgs())
	err := m.Validate()
	c.Check(err, jc.Satisfies, errors.IsNotValid)
	c.Check(err, gc.ErrorMatches, `machine "42" missing instance not valid`)
}

func (s *MachineSerializationSuite) TestValidateChecksInstance(c *gc.C) {
	m := newMachine(s.machineArgs("42"))
	m.SetStatus(minimalStatusArgs())
	m.SetTools(minimalAgentToolsArgs())
	m.SetInstance(minimalCloudInstanceArgs())
	err := m.Validate()
	c.Check(err, jc.Satisfies, errors.IsNotValid)
	c.Check(err, gc.ErrorMatches, `machine "42" instance: instance "instance id" missing status not valid`)
}

func (s *MachineSerializationSuite) TestNewMachineWithSupportedContainers(c *gc.C) {
	supported := []string{"lxd", "kvm"}
	args := s.machineArgs("id")
	args.SupportedContainers = &supported
	m := newMachine(args)
	supportedContainers, ok := m.SupportedContainers()
	c.Assert(ok, jc.IsTrue)
	c.Assert(supportedContainers, jc.DeepEquals, supported)
}

func (s *MachineSerializationSuite) TestNewMachineWithNoSupportedContainers(c *gc.C) {
	supported := []string{}
	args := s.machineArgs("id")
	args.SupportedContainers = &supported
	m := newMachine(args)
	supportedContainers, ok := m.SupportedContainers()
	c.Assert(ok, jc.IsTrue)
	c.Assert(supportedContainers, gc.HasLen, 0)
}

func (s *MachineSerializationSuite) TestNewMachineWithNoSupportedContainersNil(c *gc.C) {
	var supported []string
	args := s.machineArgs("id")
	args.SupportedContainers = &supported
	m := newMachine(args)
	supportedContainers, ok := m.SupportedContainers()
	c.Assert(ok, jc.IsTrue)
	c.Assert(supportedContainers, gc.HasLen, 0)
}

func (s *MachineSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalMachine("0"))
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalMachineMap("0"))
}

func (*MachineSerializationSuite) TestNestedParsing(c *gc.C) {
	machines, err := importMachines(map[string]interface{}{
		"version": 1,
		"machines": []interface{}{
			minimalMachineMap("0"),
			minimalMachineMap("1",
				minimalMachineMap("1/lxd/0"),
				minimalMachineMap("1/lxd/1"),
			),
			minimalMachineMap("2",
				minimalMachineMap("2/kvm/0",
					minimalMachineMap("2/kvm/0/lxd/0"),
					minimalMachineMap("2/kvm/0/lxd/1"),
				),
			),
		}})
	c.Assert(err, jc.ErrorIsNil)
	expected := []*machine{
		minimalMachine("0"),
		minimalMachine("1",
			minimalMachine("1/lxd/0"),
			minimalMachine("1/lxd/1"),
		),
		minimalMachine("2",
			minimalMachine("2/kvm/0",
				minimalMachine("2/kvm/0/lxd/0"),
				minimalMachine("2/kvm/0/lxd/1"),
			),
		),
	}
	c.Assert(machines, jc.DeepEquals, expected)
}

// TestNestedParsingWithPriorVersion tests the scenario of a using a version of
// a machine with a older version of a cloud instance. We want to ensure that
// we can mix and match type versions without inducing a panic that was found
// whilst developing.
func (*MachineSerializationSuite) TestNestedParsingWithPriorVersion(c *gc.C) {
	machines, err := importMachines(map[string]interface{}{
		"version": 1,
		"machines": []interface{}{
			minimalMachineMapWithPriorInstanceMap("0"),
			minimalMachineMapWithPriorInstanceMap("1",
				minimalMachineMapWithPriorInstanceMap("1/lxd/0"),
				minimalMachineMapWithPriorInstanceMap("1/lxd/1"),
			),
			minimalMachineMapWithPriorInstanceMap("2",
				minimalMachineMapWithPriorInstanceMap("2/kvm/0",
					minimalMachineMapWithPriorInstanceMap("2/kvm/0/lxd/0"),
					minimalMachineMapWithPriorInstanceMap("2/kvm/0/lxd/1"),
				),
			),
		}})
	c.Assert(err, jc.ErrorIsNil)
	expected := []*machine{
		minimalMachineWithPriorInstanceMap("0"),
		minimalMachineWithPriorInstanceMap("1",
			minimalMachineWithPriorInstanceMap("1/lxd/0"),
			minimalMachineWithPriorInstanceMap("1/lxd/1"),
		),
		minimalMachineWithPriorInstanceMap("2",
			minimalMachineWithPriorInstanceMap("2/kvm/0",
				minimalMachineWithPriorInstanceMap("2/kvm/0/lxd/0"),
				minimalMachineWithPriorInstanceMap("2/kvm/0/lxd/1"),
			),
		),
	}
	c.Assert(machines, jc.DeepEquals, expected)
}

func (s *MachineSerializationSuite) addOpenedPorts(m Machine) []OpenedPortsArgs {
	args := []OpenedPortsArgs{
		{
			SubnetID: "0.1.2.0/24",
			OpenedPorts: []PortRangeArgs{
				{
					UnitName: "magic/0",
					FromPort: 1234,
					ToPort:   2345,
					Protocol: "tcp",
				},
			},
		}, {
			SubnetID: "",
			OpenedPorts: []PortRangeArgs{
				{
					UnitName: "unicorn/0",
					FromPort: 80,
					ToPort:   80,
					Protocol: "tcp",
				},
			},
		},
	}
	m.AddOpenedPorts(args[0])
	m.AddOpenedPorts(args[1])
	return args
}

func (s *MachineSerializationSuite) TestOpenedPorts(c *gc.C) {
	m := newMachine(s.machineArgs("42"))
	args := s.addOpenedPorts(m)
	ports := m.OpenedPorts()
	c.Assert(ports, gc.HasLen, 2)
	withSubnet, withoutSubnet := ports[0], ports[1]
	c.Assert(withSubnet.SubnetID(), gc.Equals, "0.1.2.0/24")
	c.Assert(withoutSubnet.SubnetID(), gc.Equals, "")
	opened := withSubnet.OpenPorts()
	c.Assert(opened, gc.HasLen, 1)
	s.AssertPortRange(c, opened[0], args[0].OpenedPorts[0])
	opened = withoutSubnet.OpenPorts()
	c.Assert(opened, gc.HasLen, 1)
	s.AssertPortRange(c, opened[0], args[1].OpenedPorts[0])
}

func (s *MachineSerializationSuite) TestAnnotations(c *gc.C) {
	initial := minimalMachine("42")
	annotations := map[string]string{
		"string":  "value",
		"another": "one",
	}
	initial.SetAnnotations(annotations)

	machine := s.exportImport(c, initial)
	c.Assert(machine.Annotations(), jc.DeepEquals, annotations)
}

func (s *MachineSerializationSuite) TestConstraints(c *gc.C) {
	initial := minimalMachine("42")
	args := ConstraintsArgs{
		Architecture: "amd64",
		Memory:       8 * gig,
		RootDisk:     40 * gig,
	}
	initial.SetConstraints(args)

	machine := s.exportImport(c, initial)
	c.Assert(machine.Constraints(), jc.DeepEquals, newConstraints(args))
}

func (s *MachineSerializationSuite) exportImport(c *gc.C, machine_ *machine) *machine {
	initial := machines{
		Version:   1,
		Machines_: []*machine{machine_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	machines, err := importMachines(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(machines, gc.HasLen, 1)
	return machines[0]
}

func (s *MachineSerializationSuite) TestParsingSerializedData(c *gc.C) {
	// TODO: need to fully specify a machine.
	args := s.machineArgs("0")
	supported := []string{"kvm", "lxd"}
	args.SupportedContainers = &supported
	m := newMachine(args)
	m.SetTools(minimalAgentToolsArgs())
	m.SetStatus(minimalStatusArgs())
	m.SetInstance(minimalCloudInstanceArgs())
	m.Instance().SetStatus(minimalStatusArgs())
	m.Instance().SetModificationStatus(minimalStatusArgs())
	m.AddBlockDevice(allBlockDeviceArgs())
	s.addOpenedPorts(m)

	// Just use one set of address args for both machine and provider.
	addrArgs := []AddressArgs{
		{
			Value: "10.0.0.10",
			Type:  "special",
		}, {
			Value: "10.1.2.3",
			Type:  "other",
		},
	}
	m.SetAddresses(addrArgs, addrArgs)
	m.SetPreferredAddresses(addrArgs[0], addrArgs[1])

	// Make sure the machine is valid.
	c.Assert(m.Validate(), jc.ErrorIsNil)

	machine := s.exportImport(c, m)
	c.Assert(machine, jc.DeepEquals, m)
}

type AgentToolsSerializationSuite struct {
	SerializationSuite
}

var _ = gc.Suite(&AgentToolsSerializationSuite{})

func (s *AgentToolsSerializationSuite) SetUpTest(c *gc.C) {
	s.importName = "agentTools"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importAgentTools(m)
	}
}

func (s *AgentToolsSerializationSuite) TestNewAgentTools(c *gc.C) {
	args := AgentToolsArgs{
		Version: version.MustParseBinary("3.4.5-trusty-amd64"),
		URL:     "some-url",
		SHA256:  "long-hash",
		Size:    123456789,
	}
	instance := newAgentTools(args)

	c.Assert(instance.Version(), gc.Equals, args.Version)
	c.Assert(instance.URL(), gc.Equals, args.URL)
	c.Assert(instance.SHA256(), gc.Equals, args.SHA256)
	c.Assert(instance.Size(), gc.Equals, args.Size)
}

func minimalAgentToolsMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"version":       1,
		"tools-version": "3.4.5-trusty-amd64",
		"url":           "some-url",
		"sha256":        "long-hash",
		"size":          123456789,
	}
}

func minimalAgentToolsArgs() AgentToolsArgs {
	return AgentToolsArgs{
		Version: version.MustParseBinary("3.4.5-trusty-amd64"),
		URL:     "some-url",
		SHA256:  "long-hash",
		Size:    123456789,
	}
}

func minimalAgentTools() *agentTools {
	return newAgentTools(minimalAgentToolsArgs())
}

func (s *AgentToolsSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalAgentTools())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalAgentToolsMap())
}

func (s *AgentToolsSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := newAgentTools(AgentToolsArgs{
		Version: version.MustParseBinary("2.0.4-trusty-amd64"),
		URL:     "some-url",
		SHA256:  "long-hash",
		Size:    123456789,
	})
	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	instance, err := importAgentTools(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(instance, jc.DeepEquals, initial)
}
