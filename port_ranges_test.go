// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type MachinePortRangeSerializationSuite struct {
}

var _ = gc.Suite(&MachinePortRangeSerializationSuite{})

func (*MachinePortRangeSerializationSuite) TestParsingSerializedData(c *gc.C) {
	initial := &machinePortRanges{
		Version: 1,
		ByUnit_: map[string]*unitPortRanges{
			"lorem/0": &unitPortRanges{
				ByEndpoint_: map[string][]*unitPortRange{
					"dmz": []*unitPortRange{
						newUnitPortRange(1234, 2345, "tcp"),
						newUnitPortRange(1337, 1337, "udp"),
					},
				},
			},
			"ipsum/0": &unitPortRanges{
				ByEndpoint_: map[string][]*unitPortRange{
					"": []*unitPortRange{
						newUnitPortRange(8080, 8080, "tcp"),
					},
				},
			},
		},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	imported, err := importMachinePortRanges(source)
	c.Assert(err, jc.ErrorIsNil)

	byUnit := imported.ByUnit()
	c.Assert(byUnit, gc.HasLen, 2)

	// Check lorem/0 ports
	loremPortsByEndpoint := byUnit["lorem/0"].ByEndpoint()
	c.Assert(loremPortsByEndpoint, gc.HasLen, 1)
	loremDMZPorts := loremPortsByEndpoint["dmz"]
	c.Assert(loremDMZPorts, gc.HasLen, 2)
	c.Assert(loremDMZPorts[0], gc.DeepEquals, newUnitPortRange(1234, 2345, "tcp"))
	c.Assert(loremDMZPorts[1], gc.DeepEquals, newUnitPortRange(1337, 1337, "udp"))

	// Check ipsum/0 ports
	ipsumPortsByEndpoint := byUnit["ipsum/0"].ByEndpoint()
	c.Assert(ipsumPortsByEndpoint, gc.HasLen, 1)
	ipsumAllEndpointPorts := ipsumPortsByEndpoint[""]
	c.Assert(ipsumAllEndpointPorts, gc.HasLen, 1)
	c.Assert(ipsumAllEndpointPorts[0], gc.DeepEquals, newUnitPortRange(8080, 8080, "tcp"))
}
