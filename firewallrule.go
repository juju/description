// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

// FirewallRule represents a firewall ruleset for a known service type, with
// whitelist CIDRs.
type FirewallRule interface {
	ID() string
	WellKnownService() string
	WhitelistCIDRs() []string
}

type firewallRules struct {
	Version       int             `yaml:"version"`
	FirewallRules []*firewallRule `yaml:"firewall-rules"`
}

type firewallRule struct {
	ID_               string   `yaml:"id"`
	WellKnownService_ string   `yaml:"well-known-service"`
	WhitelistCIDRs_   []string `yaml:"whitelist-cidrs"`
}

// FirewallRuleArgs is an argument struct used to add a firewall rule.
type FirewallRuleArgs struct {
	ID               string
	WellKnownService string
	WhitelistCIDRs   []string
}

func newFirewallRule(args FirewallRuleArgs) *firewallRule {
	f := &firewallRule{
		ID_:               args.ID,
		WellKnownService_: args.WellKnownService,
		WhitelistCIDRs_:   args.WhitelistCIDRs,
	}
	return f
}

// ID implements FirewallRule
func (f *firewallRule) ID() string {
	return f.ID_
}

// WellKnownService implements FirewallRule
func (f *firewallRule) WellKnownService() string {
	return f.WellKnownService_
}

// WhitelistCIDRs implements FirewallRule
func (f *firewallRule) WhitelistCIDRs() []string {
	return f.WhitelistCIDRs_
}

func importFirewallRules(source interface{}) ([]*firewallRule, error) {
	checker := versionedChecker("firewall-rules")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "firewall rules version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	getFields, ok := firewallRuleFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["firewall-rules"].([]interface{})
	return importFirewallRuleList(sourceList, schema.FieldMap(getFields()), version)
}

func importFirewallRuleList(sourceList []interface{}, checker schema.Checker, version int) ([]*firewallRule, error) {
	result := make([]*firewallRule, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for firewall rule %d, %T", i, value)
		}
		coerced, err := checker.Coerce(source, nil)
		if err != nil {
			return nil, errors.Annotatef(err, "firewall rule %d v%d schema check failed", i, version)
		}
		valid := coerced.(map[string]interface{})
		firewallRle, err := newFirewallRuleFromValid(valid, version)
		if err != nil {
			return nil, errors.Annotatef(err, "firewall rule %d", i)
		}
		result[i] = firewallRle
	}
	return result, nil
}

func newFirewallRuleFromValid(valid map[string]interface{}, version int) (*firewallRule, error) {
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.
	result := &firewallRule{
		ID_:               valid["id"].(string),
		WellKnownService_: valid["well-known-service"].(string),
		WhitelistCIDRs_:   convertToStringSlice(valid["whitelist-cidrs"]),
	}
	return result, nil
}

var firewallRuleFieldsFuncs = map[int]fieldsFunc{
	1: firewallRuleV1Fields,
}

func firewallRuleV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":                 schema.String(),
		"well-known-service": schema.String(),
		"whitelist-cidrs":    schema.List(schema.String()),
	}
	defaults := schema.Defaults{}
	return fields, defaults
}
