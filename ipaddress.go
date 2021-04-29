// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type ipaddresses struct {
	Version      int          `yaml:"version"`
	IPAddresses_ []*ipaddress `yaml:"ip-addresses"`
}

type ipaddress struct {
	ProviderID_        string   `yaml:"provider-id,omitempty"`
	DeviceName_        string   `yaml:"device-name"`
	MachineID_         string   `yaml:"machine-id"`
	SubnetCIDR_        string   `yaml:"subnet-cidr"`
	ConfigMethod_      string   `yaml:"config-method"`
	Value_             string   `yaml:"value"`
	DNSServers_        []string `yaml:"dns-servers"`
	DNSSearchDomains_  []string `yaml:"dns-search-domains"`
	GatewayAddress_    string   `yaml:"gateway-address"`
	IsDefaultGateway_  bool     `yaml:"is-default-gateway"`
	ProviderNetworkID_ string   `yaml:"provider-network-id,omitempty"`
	ProviderSubnetID_  string   `yaml:"provider-subnet-id,omitempty"`
	Origin_            string   `yaml:"origin"`
}

// ProviderID implements IPAddress.
func (i *ipaddress) ProviderID() string {
	return i.ProviderID_
}

// DeviceName implements IPAddress.
func (i *ipaddress) DeviceName() string {
	return i.DeviceName_
}

// MachineID implements IPAddress.
func (i *ipaddress) MachineID() string {
	return i.MachineID_
}

// SubnetCIDR implements IPAddress.
func (i *ipaddress) SubnetCIDR() string {
	return i.SubnetCIDR_
}

// ConfigMethod implements IPAddress.
func (i *ipaddress) ConfigMethod() string {
	return i.ConfigMethod_
}

// Value implements IPAddress.
func (i *ipaddress) Value() string {
	return i.Value_
}

// DNSServers implements IPAddress.
func (i *ipaddress) DNSServers() []string {
	return i.DNSServers_
}

// DNSSearchDomains implements IPAddress.
func (i *ipaddress) DNSSearchDomains() []string {
	return i.DNSSearchDomains_
}

// GatewayAddress implements IPAddress.
func (i *ipaddress) GatewayAddress() string {
	return i.GatewayAddress_
}

// IsDefaultGateway implements IPAddress.
func (i *ipaddress) IsDefaultGateway() bool {
	return i.IsDefaultGateway_
}

// ProviderNetworkID implements IPAddress.
func (i *ipaddress) ProviderNetworkID() string {
	return i.ProviderNetworkID_
}

// ProviderSubnetID implements IPAddress.
func (i *ipaddress) ProviderSubnetID() string {
	return i.ProviderSubnetID_
}

// Origin implements IPAddress.
func (i *ipaddress) Origin() string {
	return i.Origin_
}

// IPAddressArgs is an argument struct used to create a
// new internal ipaddress type that supports the IPAddress interface.
type IPAddressArgs struct {
	ProviderID        string
	DeviceName        string
	MachineID         string
	SubnetCIDR        string
	ConfigMethod      string
	Value             string
	DNSServers        []string
	DNSSearchDomains  []string
	GatewayAddress    string
	IsDefaultGateway  bool
	ProviderNetworkID string
	ProviderSubnetID  string
	Origin            string
}

func newIPAddress(args IPAddressArgs) *ipaddress {
	return &ipaddress{
		ProviderID_:        args.ProviderID,
		DeviceName_:        args.DeviceName,
		MachineID_:         args.MachineID,
		SubnetCIDR_:        args.SubnetCIDR,
		ConfigMethod_:      args.ConfigMethod,
		Value_:             args.Value,
		DNSServers_:        args.DNSServers,
		DNSSearchDomains_:  args.DNSSearchDomains,
		GatewayAddress_:    args.GatewayAddress,
		IsDefaultGateway_:  args.IsDefaultGateway,
		ProviderNetworkID_: args.ProviderNetworkID,
		ProviderSubnetID_:  args.ProviderSubnetID,
		Origin_:            args.Origin,
	}
}

func importIPAddresses(source map[string]interface{}) ([]*ipaddress, error) {
	checker := versionedChecker("ip-addresses")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "ip-addresses version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	importFunc, ok := ipAddressDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList := valid["ip-addresses"].([]interface{})
	return importIPAddressList(sourceList, importFunc)
}

type ipAddressDeserializationFunc func(map[string]interface{}) (*ipaddress, error)

func importIPAddressList(sourceList []interface{}, importFunc ipAddressDeserializationFunc) ([]*ipaddress, error) {
	result := make([]*ipaddress, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for ip-address %d, %T", i, value)
		}
		ipaddress, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "ip-address %d", i)
		}
		result = append(result, ipaddress)
	}
	return result, nil
}

var ipAddressDeserializationFuncs = map[int]ipAddressDeserializationFunc{
	1: importIPAddressV1,
	2: importIPAddressV2,
	3: importIPAddressV3,
}

func parseDnsFields(valid map[string]interface{}) ([]string, []string) {
	dnsServersInterface := valid["dns-servers"].([]interface{})
	dnsServers := make([]string, len(dnsServersInterface))
	for i, d := range dnsServersInterface {
		dnsServers[i] = d.(string)
	}
	dnsSearchInterface := valid["dns-search-domains"].([]interface{})
	dnsSearch := make([]string, len(dnsSearchInterface))
	for i, d := range dnsSearchInterface {
		dnsSearch[i] = d.(string)
	}
	return dnsServers, dnsSearch
}

func importIPAddressV1(source map[string]interface{}) (*ipaddress, error) {
	fields, defaults := ipAddressV1Schema()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "ip address v1 schema check failed")
	}

	return ipAddressV1(coerced.(map[string]interface{})), nil
}

func importIPAddressV2(source map[string]interface{}) (*ipaddress, error) {
	fields, defaults := ipAddressV2Schema()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "ip address v2 schema check failed")
	}

	return ipAddressV2(coerced.(map[string]interface{})), nil
}

func importIPAddressV3(source map[string]interface{}) (*ipaddress, error) {
	fields, defaults := ipAddressV3Schema()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "ip address v2 schema check failed")
	}

	return ipAddressV3(coerced.(map[string]interface{})), nil
}

func ipAddressV3Schema() (schema.Fields, schema.Defaults) {
	fields, defaults := ipAddressV2Schema()

	fields["provider-network-id"] = schema.String()
	fields["provider-subnet-id"] = schema.String()
	fields["origin"] = schema.String()

	defaults["provider-network-id"] = ""
	defaults["provider-subnet-id"] = ""
	defaults["origin"] = ""

	return fields, defaults
}

func ipAddressV2Schema() (schema.Fields, schema.Defaults) {
	fields, defaults := ipAddressV1Schema()

	fields["is-default-gateway"] = schema.Bool()
	defaults["is-default-gateway"] = false

	return fields, defaults
}

func ipAddressV1Schema() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"provider-id":        schema.String(),
		"device-name":        schema.String(),
		"machine-id":         schema.String(),
		"subnet-cidr":        schema.String(),
		"config-method":      schema.String(),
		"value":              schema.String(),
		"dns-servers":        schema.List(schema.String()),
		"dns-search-domains": schema.List(schema.String()),
		"gateway-address":    schema.String(),
	}

	defaults := schema.Defaults{
		"provider-id": "",
	}

	return fields, defaults
}

func ipAddressV3(valid map[string]interface{}) *ipaddress {
	addr := ipAddressV2(valid)
	addr.ProviderNetworkID_ = valid["provider-network-id"].(string)
	addr.ProviderSubnetID_ = valid["provider-subnet-id"].(string)
	addr.Origin_ = valid["origin"].(string)
	return addr
}

func ipAddressV2(valid map[string]interface{}) *ipaddress {
	addr := ipAddressV1(valid)
	addr.IsDefaultGateway_ = valid["is-default-gateway"].(bool)
	return addr
}

func ipAddressV1(valid map[string]interface{}) *ipaddress {
	dnsServers, dnsSearch := parseDnsFields(valid)
	return &ipaddress{
		ProviderID_:       valid["provider-id"].(string),
		DeviceName_:       valid["device-name"].(string),
		MachineID_:        valid["machine-id"].(string),
		SubnetCIDR_:       valid["subnet-cidr"].(string),
		ConfigMethod_:     valid["config-method"].(string),
		Value_:            valid["value"].(string),
		DNSServers_:       dnsServers,
		DNSSearchDomains_: dnsSearch,
		GatewayAddress_:   valid["gateway-address"].(string),
	}
}
