// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// PortRanges represents a collection of port ranges that are open.
type PortRanges interface {
	ByUnit() map[string]UnitPortRanges
}

// UnitPortRanges represents the of port ranges opened by a particular Unit for
// one or more endpoints.
type UnitPortRanges interface {
	ByEndpoint() map[string][]UnitPortRange
}

// UnitPortRange represents a contiguous port range opened by a unit.
type UnitPortRange interface {
	FromPort() int
	ToPort() int
	Protocol() string
}

// OpenedPortRangeArgs is an argument struct used to add a new port range to
// a machine.
type OpenedPortRangeArgs struct {
	UnitName     string
	EndpointName string

	FromPort int
	ToPort   int
	Protocol string
}

type machinePortRanges struct {
	Version int `yaml:"version"`

	// The set of opened port ranges by unit.
	ByUnit_ map[string]*unitPortRanges `yaml:"machine-port-ranges"`
}

func newMachinePortRanges() *machinePortRanges {
	return &machinePortRanges{
		Version: 1,
		ByUnit_: make(map[string]*unitPortRanges),
	}
}

// ByUnit implements MachinePortRanges.
func (p *machinePortRanges) ByUnit() map[string]UnitPortRanges {
	res := make(map[string]UnitPortRanges, len(p.ByUnit_))
	for unitName, upr := range p.ByUnit_ {
		res[unitName] = upr
	}
	return res
}

type unitPortRanges struct {
	// The set of opened port ranges for each endpoint.
	ByEndpoint_ map[string][]*unitPortRange `yaml:"unit-port-ranges"`
}

func newUnitPortRanges() *unitPortRanges {
	return &unitPortRanges{
		ByEndpoint_: make(map[string][]*unitPortRange),
	}
}

// ByEndpoint implements UnitPortRanges.
func (p *unitPortRanges) ByEndpoint() map[string][]UnitPortRange {
	res := make(map[string][]UnitPortRange, len(p.ByEndpoint_))
	for endpointName, portRanges := range p.ByEndpoint_ {
		rangeList := make([]UnitPortRange, len(portRanges))
		for i, portRange := range portRanges {
			rangeList[i] = portRange
		}
		res[endpointName] = rangeList
	}
	return res
}

type unitPortRange struct {
	FromPort_ int    `yaml:"from-port"`
	ToPort_   int    `yaml:"to-port"`
	Protocol_ string `yaml:"protocol"`
}

func newUnitPortRange(fromPort, toPort int, protocol string) *unitPortRange {
	return &unitPortRange{
		FromPort_: fromPort,
		ToPort_:   toPort,
		Protocol_: protocol,
	}
}

// FromPort implements UnitPortRange.
func (p unitPortRange) FromPort() int {
	return p.FromPort_
}

// ToPort implements UnitPortRange.
func (p unitPortRange) ToPort() int {
	return p.ToPort_
}

// Protocol implements UnitPortRange.
func (p unitPortRange) Protocol() string {
	return p.Protocol_
}

func importMachinePortRanges(source map[string]interface{}) (*machinePortRanges, error) {
	checker := versionedMapChecker("machine-port-ranges")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "machine-port-ranges version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := machinePortRangeDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceMap := valid["machine-port-ranges"].(map[string]interface{})
	return importFunc(sourceMap)
}

type machinePortRangeDeserializationFunc func(map[string]interface{}) (*machinePortRanges, error)

var machinePortRangeDeserializationFuncs = map[int]machinePortRangeDeserializationFunc{
	1: importMachinePortRangeV1,
}

func importMachinePortRangeV1(source map[string]interface{}) (*machinePortRanges, error) {
	unitChecker := schema.FieldMap(schema.Fields{
		"unit-port-ranges": schema.StringMap(
			schema.List(schema.StringMap(schema.Any())),
		),
	}, nil) // no defaults

	portRangeChecker := schema.FieldMap(schema.Fields{
		"from-port": schema.Int(),
		"to-port":   schema.Int(),
		"protocol":  schema.String(),
	}, nil) // no defaults

	mpr := &machinePortRanges{
		Version: 1,
		ByUnit_: make(map[string]*unitPortRanges),
	}
	for unitName, unitSource := range source {
		coercedUnit, err := unitChecker.Coerce(unitSource, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "machine-port-range v1 schema check failed")
		}

		validFieldMap := coercedUnit.(map[string]interface{})
		validEndpointMap := validFieldMap["unit-port-ranges"].(map[string]interface{})
		upr := &unitPortRanges{
			ByEndpoint_: make(map[string][]*unitPortRange),
		}
		for endpointName, portRangeListSource := range validEndpointMap {
			validPortRangeListSource := portRangeListSource.([]interface{})
			portRangeList := make([]*unitPortRange, len(validPortRangeListSource))
			for i, portRangeSource := range validPortRangeListSource {
				validPortRangeSource := portRangeSource.(map[string]interface{})
				pr, err := importUnitPortRangeV1(validPortRangeSource, portRangeChecker)
				if err != nil {
					return nil, errors.Annotatef(err, "unable to parse V1 unit port range")
				}
				portRangeList[i] = pr
			}
			upr.ByEndpoint_[endpointName] = portRangeList
		}

		mpr.ByUnit_[unitName] = upr
	}

	return mpr, nil
}

func importUnitPortRangeV1(source map[string]interface{}, checker schema.Checker) (*unitPortRange, error) {
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "unit-port-range v1 schema check failed")
	}

	valid := coerced.(map[string]interface{})
	return &unitPortRange{
		FromPort_: int(valid["from-port"].(int64)),
		ToPort_:   int(valid["to-port"].(int64)),
		Protocol_: valid["protocol"].(string),
	}, nil
}
