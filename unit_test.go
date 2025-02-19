// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/names/v6"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type UnitSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&UnitSerializationSuite{})

func (s *UnitSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "units"
	s.sliceName = "units"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importUnits(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["units"] = []interface{}{}
	}
}

func minimalUnitMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"name":                     "ubuntu/0",
		"machine":                  "0",
		"agent-status":             minimalStatusMap(),
		"agent-status-history":     emptyStatusHistoryMap(),
		"workload-status":          minimalStatusMap(),
		"workload-status-history":  emptyStatusHistoryMap(),
		"workload-version-history": emptyStatusHistoryMap(),
		"password-hash":            "secure-hash",
		"tools":                    minimalAgentToolsMap(),
		"resources": map[interface{}]interface{}{
			"version":   2,
			"resources": []interface{}{},
		},
		"payloads": map[interface{}]interface{}{
			"version":  1,
			"payloads": []interface{}{},
		},
		"charm-state": map[interface{}]interface{}{
			"some-charm-key": "0xbadc0ffee",
		},
		"relation-state": map[interface{}]interface{}{
			1: "yaml-encoded state for relation 1",
			2: "yaml-encoded state for relation 2",
		},
		"uniter-state":       "yaml-encoded state for uniter",
		"storage-state":      "yaml-encoded state for storage",
		"meter-status-state": "yaml-encoded state for meter status worker",
	}
}

func minimalUnitMapCAAS() map[interface{}]interface{} {
	result := minimalUnitMap()
	delete(result, "tools")
	result["cloud-container"] = map[interface{}]interface{}{
		"version":     1,
		"provider-id": "some-provider",
		"address":     map[interface{}]interface{}{"version": 2, "value": "10.0.0.1", "type": "special"},
		"ports":       []interface{}{"80", "443"},
	}
	return result
}

func minimalCloudContainerArgs() CloudContainerArgs {
	return CloudContainerArgs{
		ProviderId: "some-provider",
		Address:    AddressArgs{Value: "10.0.0.1", Type: "special"},
		Ports:      []string{"80", "443"},
	}
}

func minimalUnit(args ...UnitArgs) *unit {
	if len(args) == 0 {
		args = []UnitArgs{minimalUnitArgs(IAAS)}
	}
	u := newUnit(args[0])
	u.SetAgentStatus(minimalStatusArgs())
	u.SetWorkloadStatus(minimalStatusArgs())
	if u.Type_ != CAAS {
		u.SetTools(minimalAgentToolsArgs())
	}
	return u
}

func minimalUnitArgs(modelType string) UnitArgs {
	result := UnitArgs{
		Tag:          names.NewUnitTag("ubuntu/0"),
		Type:         modelType,
		Machine:      names.NewMachineTag("0"),
		PasswordHash: "secure-hash",
		CharmState: map[string]string{
			"some-charm-key": "0xbadc0ffee",
		},
		RelationState: map[int]string{
			1: "yaml-encoded state for relation 1",
			2: "yaml-encoded state for relation 2",
		},
		UniterState:      "yaml-encoded state for uniter",
		StorageState:     "yaml-encoded state for storage",
		MeterStatusState: "yaml-encoded state for meter status worker",
	}
	if modelType == CAAS {
		result.CloudContainer = &CloudContainerArgs{
			ProviderId: "some-provider",
			Address:    AddressArgs{Value: "10.0.0.1", Type: "special"},
			Ports:      []string{"80", "443"},
		}
	}
	return result
}

func (s *UnitSerializationSuite) completeUnit() *unit {
	// This unit is about completeness, not reasonableness. That is why the
	// unit has a principal (normally only for subordinates), and also a list
	// of subordinates.
	args := UnitArgs{
		Tag:          names.NewUnitTag("ubuntu/0"),
		Machine:      names.NewMachineTag("0"),
		PasswordHash: "secure-hash",
		Principal:    names.NewUnitTag("principal/0"),
		Subordinates: []names.UnitTag{
			names.NewUnitTag("sub1/0"),
			names.NewUnitTag("sub2/0"),
		},
		WorkloadVersion: "malachite",
		MeterStatusCode: "meter code",
		MeterStatusInfo: "meter info",
	}
	unit := newUnit(args)
	unit.SetAgentStatus(minimalStatusArgs())
	unit.SetWorkloadStatus(minimalStatusArgs())
	unit.SetTools(minimalAgentToolsArgs())
	unit.SetCloudContainer(minimalCloudContainerArgs())
	unit.SetCharmState(map[string]string{
		"some-charm-key": "0xbadc0ffee",
	})
	unit.SetRelationState(map[int]string{
		1: "yaml-encoded state for relation 1",
		2: "yaml-encoded state for relation 2",
	})
	unit.SetUniterState("yaml-encoded state for uniter")
	unit.SetStorageState("yaml-encoded state for storage")
	unit.SetMeterStatusState("yaml-encoded state for meter status worker")
	return unit
}

func (s *UnitSerializationSuite) TestNewUnit(c *gc.C) {
	unit := s.completeUnit()

	c.Assert(unit.Tag(), gc.Equals, names.NewUnitTag("ubuntu/0"))
	c.Assert(unit.Name(), gc.Equals, "ubuntu/0")
	c.Assert(unit.Machine(), gc.Equals, names.NewMachineTag("0"))
	c.Assert(unit.PasswordHash(), gc.Equals, "secure-hash")
	c.Assert(unit.Principal(), gc.Equals, names.NewUnitTag("principal/0"))
	c.Assert(unit.Subordinates(), jc.DeepEquals, []names.UnitTag{
		names.NewUnitTag("sub1/0"),
		names.NewUnitTag("sub2/0"),
	})
	c.Assert(unit.WorkloadVersion(), gc.Equals, "malachite")
	c.Assert(unit.MeterStatusCode(), gc.Equals, "meter code")
	c.Assert(unit.MeterStatusInfo(), gc.Equals, "meter info")
	c.Assert(unit.Tools(), gc.NotNil)
	c.Assert(unit.WorkloadStatus(), gc.NotNil)
	c.Assert(unit.AgentStatus(), gc.NotNil)
	c.Assert(unit.CloudContainer(), gc.NotNil)
	c.Assert(unit.CharmState(), gc.DeepEquals, map[string]string{
		"some-charm-key": "0xbadc0ffee",
	})
	c.Assert(unit.RelationState(), gc.DeepEquals, map[int]string{
		1: "yaml-encoded state for relation 1",
		2: "yaml-encoded state for relation 2",
	})
	c.Assert(unit.UniterState(), gc.Equals, "yaml-encoded state for uniter")
	c.Assert(unit.StorageState(), gc.Equals, "yaml-encoded state for storage")
	c.Assert(unit.MeterStatusState(), gc.Equals, "yaml-encoded state for meter status worker")
}

func (s *UnitSerializationSuite) TestMinimalUnitValid(c *gc.C) {
	unit := minimalUnit()
	c.Assert(unit.Validate(), jc.ErrorIsNil)
}

func (s *UnitSerializationSuite) TestMinimalCAASUnitValid(c *gc.C) {
	unit := minimalUnit(minimalUnitArgs(CAAS))
	c.Assert(unit.Validate(), jc.ErrorIsNil)
}

func (s *UnitSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalUnit())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalUnitMap())
}

func (s *UnitSerializationSuite) exportImportVersion(c *gc.C, unit_ *unit, version int) *unit {
	initial := units{
		Version: version,
		Units_:  []*unit{unit_},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	units, err := importUnits(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(units, gc.HasLen, 1)
	return units[0]
}

func (s *UnitSerializationSuite) exportImportV2(c *gc.C, unit *unit) *unit {
	return s.exportImportVersion(c, unit, 2)
}

func (s *UnitSerializationSuite) exportImportLatest(c *gc.C, unit *unit) *unit {
	return s.exportImportVersion(c, unit, 3)
}

func (s *UnitSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := s.completeUnit()
	unit := s.exportImportLatest(c, initial)
	c.Assert(unit, jc.DeepEquals, initial)
}

func (s *UnitSerializationSuite) TestV1ParsingReturnsLatest(c *gc.C) {
	unitV1 := minimalUnit()

	// Make a unit with fields not in v1 removed.
	unitLatest := minimalUnit()
	unitLatest.CloudContainer_ = nil
	unitLatest.Type_ = ""
	unitLatest.CharmState_ = nil
	unitLatest.RelationState_ = nil
	unitLatest.UniterState_ = ""
	unitLatest.StorageState_ = ""
	unitLatest.MeterStatusState_ = ""

	unitResult := s.exportImportVersion(c, unitV1, 1)
	c.Assert(unitResult, jc.DeepEquals, unitLatest)
}

func (s *UnitSerializationSuite) TestV2ParsingReturnsLatest(c *gc.C) {
	unitV2 := s.completeUnit()

	// Make a unit with fields not in v2 removed.
	unitLatest := s.completeUnit()
	unitLatest.CharmState_ = nil
	unitLatest.RelationState_ = nil
	unitLatest.UniterState_ = ""
	unitLatest.StorageState_ = ""
	unitLatest.MeterStatusState_ = ""

	unitResult := s.exportImportVersion(c, unitV2, 2)
	c.Assert(unitResult, jc.DeepEquals, unitLatest)
}

func (s *UnitSerializationSuite) TestAnnotations(c *gc.C) {
	initial := minimalUnit()
	annotations := map[string]string{
		"string":  "value",
		"another": "one",
	}
	initial.SetAnnotations(annotations)

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.Annotations(), jc.DeepEquals, annotations)
}

func (s *UnitSerializationSuite) TestConstraints(c *gc.C) {
	initial := minimalUnit()
	args := ConstraintsArgs{
		Architecture: "amd64",
		Memory:       8 * gig,
		RootDisk:     40 * gig,
	}
	initial.SetConstraints(args)

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.Constraints(), jc.DeepEquals, newConstraints(args))
}

func (s *UnitSerializationSuite) TestCloudContainer(c *gc.C) {
	initial := minimalUnit(minimalUnitArgs(CAAS))
	args := CloudContainerArgs{
		ProviderId: "some-provider",
		Address:    AddressArgs{Value: "10.0.0.1", Type: "special"},
		Ports:      []string{"80", "443"},
	}
	initial.SetCloudContainer(args)

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.CloudContainer(), jc.DeepEquals, newCloudContainer(&args))
}

func (s *UnitSerializationSuite) TestCharmState(c *gc.C) {
	initial := minimalUnit()
	initial.SetCharmState(map[string]string{
		"foo": "bar",
	})

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.CharmState(), jc.DeepEquals, initial.CharmState())
}

func (s *UnitSerializationSuite) TestRelationState(c *gc.C) {
	initial := minimalUnit()
	initial.SetRelationState(map[int]string{
		42: "random",
	})

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.RelationState(), jc.DeepEquals, initial.RelationState())
}

func (s *UnitSerializationSuite) TestUniterState(c *gc.C) {
	initial := minimalUnit()
	initial.SetUniterState("my new uniter state")

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.UniterState(), jc.DeepEquals, initial.UniterState())
}

func (s *UnitSerializationSuite) TestStorageState(c *gc.C) {
	initial := minimalUnit()
	initial.SetStorageState("my new storage state")

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.StorageState(), jc.DeepEquals, initial.StorageState())
}

func (s *UnitSerializationSuite) TestMeterStatusState(c *gc.C) {
	initial := minimalUnit()
	initial.SetMeterStatusState("my new meter status state")

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.MeterStatusState(), jc.DeepEquals, initial.MeterStatusState())
}

func (s *UnitSerializationSuite) TestCAASUnitNoTools(c *gc.C) {
	initial := minimalUnit(minimalUnitArgs(CAAS))
	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.Tools_, gc.IsNil)
}

func (s *UnitSerializationSuite) TestAgentStatusHistory(c *gc.C) {
	initial := minimalUnit()
	args := testStatusHistoryArgs()
	initial.SetAgentStatusHistory(args)

	unit := s.exportImportLatest(c, initial)
	for i, point := range unit.AgentStatusHistory() {
		c.Check(point.Value(), gc.Equals, args[i].Value)
		c.Check(point.Message(), gc.Equals, args[i].Message)
		c.Check(point.Data(), jc.DeepEquals, args[i].Data)
		c.Check(point.Updated(), gc.Equals, args[i].Updated)
	}
}

func (s *UnitSerializationSuite) TestWorkloadStatusHistory(c *gc.C) {
	initial := minimalUnit()
	args := testStatusHistoryArgs()
	initial.SetWorkloadStatusHistory(args)

	unit := s.exportImportLatest(c, initial)
	for i, point := range unit.WorkloadStatusHistory() {
		c.Check(point.Value(), gc.Equals, args[i].Value)
		c.Check(point.Message(), gc.Equals, args[i].Message)
		c.Check(point.Data(), jc.DeepEquals, args[i].Data)
		c.Check(point.Updated(), gc.Equals, args[i].Updated)
	}
}

func (s *UnitSerializationSuite) TestResources(c *gc.C) {
	initial := minimalUnit()
	rFoo := initial.AddResource(UnitResourceArgs{
		Name: "foo",
		RevisionArgs: ResourceRevisionArgs{
			Revision: 3,
		},
	})
	rBar := initial.AddResource(UnitResourceArgs{
		Name: "bar",
		RevisionArgs: ResourceRevisionArgs{
			Revision: 1,
		},
	})

	unit := s.exportImportLatest(c, initial)
	c.Assert(unit.Resources(), jc.DeepEquals, []UnitResource{rFoo, rBar})
}

func (s *UnitSerializationSuite) TestPayloads(c *gc.C) {
	initial := minimalUnit()
	expected := initial.AddPayload(allPayloadArgs())
	c.Check(expected.Name(), gc.Equals, "bob")
	c.Check(expected.Type(), gc.Equals, "docker")
	c.Check(expected.RawID(), gc.Equals, "d06f00d")
	c.Check(expected.State(), gc.Equals, "running")
	c.Check(expected.Labels(), jc.DeepEquals, []string{"auto", "foo"})

	unit := s.exportImportLatest(c, initial)

	payloads := unit.Payloads()
	c.Assert(payloads, gc.HasLen, 1)
	c.Assert(payloads[0], jc.DeepEquals, expected)
}

func (s *UnitSerializationSuite) TestIAASMissingToolsValidated(c *gc.C) {
	u := minimalUnit()
	u.Tools_ = nil
	err := u.Validate()
	c.Assert(err, gc.ErrorMatches, `unit "ubuntu/0" missing tools not valid`)
}
