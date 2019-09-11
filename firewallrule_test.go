// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

type FirewallRuleSerializationSuite struct {
	SliceSerializationSuite
}

var _ = gc.Suite(&FirewallRuleSerializationSuite{})

func (s *FirewallRuleSerializationSuite) SetUpTest(c *gc.C) {
	s.SliceSerializationSuite.SetUpTest(c)
	s.importName = "firewall rules"
	s.sliceName = "firewall-rules"
	s.importFunc = func(m map[string]interface{}) (interface{}, error) {
		return importFirewallRules(m)
	}
	s.testFields = func(m map[string]interface{}) {
		m["firewall-rules"] = []interface{}{}
	}
}

func minimalFirewallRuleMap() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"id":                 "firewall-rlz-id",
		"well-known-service": "juju-application-offer",
		"whitelist-cidrs": []interface{}{
			"1.2.3.4/24",
			"0.0.0.1",
		},
	}
}

func minimalFirewallRule() *firewallRule {
	c := newFirewallRule(FirewallRuleArgs{
		ID:               "firewall-rlz-id",
		WellKnownService: "juju-application-offer",
		WhitelistCIDRs: []string{
			"1.2.3.4/24",
			"0.0.0.1",
		},
	})
	return c
}

func (*FirewallRuleSerializationSuite) TestNew(c *gc.C) {
	e := minimalFirewallRule()
	c.Check(e.ID(), gc.Equals, "firewall-rlz-id")
	c.Check(e.WellKnownService(), gc.Equals, "juju-application-offer")
	c.Check(e.WhitelistCIDRs(), gc.DeepEquals, []string{
		"1.2.3.4/24",
		"0.0.0.1",
	})
}

func (*FirewallRuleSerializationSuite) TestBadSchema1(c *gc.C) {
	container := map[string]interface{}{
		"version":        1,
		"firewall-rules": []interface{}{1234},
	}
	_, err := importFirewallRules(container)
	c.Assert(err, gc.ErrorMatches, `firewall rules version schema check failed: firewall-rules\[0\]: expected map, got int\(1234\)`)
}

func (*FirewallRuleSerializationSuite) TestBadSchema2(c *gc.C) {
	m := minimalFirewallRuleMap()
	m["id"] = true
	container := map[string]interface{}{
		"version":        1,
		"firewall-rules": []interface{}{m},
	}
	_, err := importFirewallRules(container)
	c.Assert(err, gc.ErrorMatches, `firewall rule 0 v1 schema check failed: id: expected string, got bool\(true\)`)
}

func (*FirewallRuleSerializationSuite) TestMinimalMatches(c *gc.C) {
	bytes, err := yaml.Marshal(minimalFirewallRuleMap())
	c.Assert(err, jc.ErrorIsNil)

	var source map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(source, jc.DeepEquals, minimalFirewallRuleMap())
}

func (s *FirewallRuleSerializationSuite) TestRoundTrip(c *gc.C) {
	rIn := minimalFirewallRule()
	rOut := s.exportImport(c, rIn)
	c.Assert(rOut, jc.DeepEquals, rIn)
}

func (s *FirewallRuleSerializationSuite) exportImport(c *gc.C, firewallRuleIn *firewallRule) *firewallRule {
	firewallRulesIn := &firewallRules{
		Version:       1,
		FirewallRules: []*firewallRule{firewallRuleIn},
	}
	bytes, err := yaml.Marshal(firewallRulesIn)
	c.Assert(err, jc.ErrorIsNil)

	var source map[string]interface{}
	err = yaml.Unmarshal(bytes, &source)
	c.Assert(err, jc.ErrorIsNil)

	firewallRulesOut, err := importFirewallRules(source)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(firewallRulesOut, gc.HasLen, 1)
	return firewallRulesOut[0]
}
