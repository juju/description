// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type LinkLayerDeviceSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&LinkLayerDeviceSerializationSuite{})

func (s *LinkLayerDeviceSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "link-layer-devices"
	s.sliceName = "link-layer-devices"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importLinkLayerDevices(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["link-layer-devices"] = []interface{}{}
	}
}

func (s *LinkLayerDeviceSerializationSuite) TestNewLinkLayerDevice(c *gc.C) {
	args := LinkLayerDeviceArgs{
		ProviderID:      "magic",
		MachineID:       "bar",
		Name:            "foo",
		MTU:             54,
		Type:            "loopback",
		MACAddress:      "DEADBEEF",
		IsAutoStart:     true,
		IsUp:            true,
		ParentName:      "bam",
		VirtualPortType: "ovs",
	}
	device := newLinkLayerDevice(args)
	c.Assert(device.ProviderID(), gc.Equals, args.ProviderID)
	c.Assert(device.MachineID(), gc.Equals, args.MachineID)
	c.Assert(device.Name(), gc.Equals, args.Name)
	c.Assert(device.MTU(), gc.Equals, args.MTU)
	c.Assert(device.Type(), gc.Equals, args.Type)
	c.Assert(device.MACAddress(), gc.Equals, args.MACAddress)
	c.Assert(device.IsAutoStart(), gc.Equals, args.IsAutoStart)
	c.Assert(device.IsUp(), gc.Equals, args.IsUp)
	c.Assert(device.ParentName(), gc.Equals, args.ParentName)
	c.Assert(device.VirtualPortType(), gc.Equals, args.VirtualPortType)
}

func (s *LinkLayerDeviceSerializationSuite) TestParsingSerializedDataV1(c *gc.C) {
	initial := linklayerdevices{
		Version: 1,
		LinkLayerDevices_: []*linklayerdevice{
			newLinkLayerDevice(LinkLayerDeviceArgs{
				ProviderID:  "magic",
				MachineID:   "bar",
				Name:        "foo",
				MTU:         54,
				Type:        "loopback",
				MACAddress:  "DEADBEEF",
				IsAutoStart: true,
				IsUp:        true,
				ParentName:  "bam",
			}),
			newLinkLayerDevice(LinkLayerDeviceArgs{Name: "weeee"}),
		},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	devices, err := importLinkLayerDevices(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(devices, jc.DeepEquals, initial.LinkLayerDevices_)
}

func (s *LinkLayerDeviceSerializationSuite) TestParsingSerializedDataV2(c *gc.C) {
	initial := linklayerdevices{
		Version: 2,
		LinkLayerDevices_: []*linklayerdevice{
			newLinkLayerDevice(LinkLayerDeviceArgs{
				ProviderID:  "magic",
				MachineID:   "bar",
				Name:        "foo",
				MTU:         54,
				Type:        "loopback",
				MACAddress:  "DEADBEEF",
				IsAutoStart: true,
				IsUp:        true,
				ParentName:  "bam",
				// V2 adds the VirtualPortType field
				VirtualPortType: "ovs",
			}),
			newLinkLayerDevice(LinkLayerDeviceArgs{Name: "weeee"}),
		},
	}

	bytes, err := yaml.Marshal(initial)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	devices, err := importLinkLayerDevices(source)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(devices, jc.DeepEquals, initial.LinkLayerDevices_)
}
