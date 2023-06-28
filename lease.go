// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"time"

	"github.com/juju/errors"
	"github.com/juju/schema"
)

// Lease represents a Lease for a given application. This doesn't include the
// singular controller lease.
type Lease interface {
	Name() string
	Holder() string
	Start() time.Time
	Expiry() time.Time
	Pinned() bool
}

var _ Lease = (*lease)(nil)

type lease struct {
	Version_ int       `yaml:"version"`
	Name_    string    `yaml:"name"`
	Holder_  string    `yaml:"holder"`
	Start_   time.Time `yaml:"start"`
	Expiry_  time.Time `yaml:"expiry"`
	Pinned_  bool      `yaml:"pinned"`
}

// LeaseArgs is an argument struct used to add a lease to the model.
type LeaseArgs struct {
	Name   string
	Holder string
	Start  time.Time
	Expiry time.Time
	Pinned bool
}

func newLease(args *LeaseArgs) *lease {
	if args == nil {
		return nil
	}

	return &lease{
		Version_: 1,
		Name_:    args.Name,
		Holder_:  args.Holder,
		Start_:   args.Start,
		Expiry_:  args.Expiry,
		Pinned_:  args.Pinned,
	}
}

// Name implements Lease.
func (l *lease) Name() string {
	return l.Name_
}

// Holder implements Lease.
func (l *lease) Holder() string {
	return l.Holder_
}

// Start implements Lease.
func (l *lease) Start() time.Time {
	return l.Start_
}

// Expiry implements Lease.
func (l *lease) Expiry() time.Time {
	return l.Expiry_
}

// Pinned implements Lease.
func (l *lease) Pinned() bool {
	return l.Pinned_
}

func importLease(source map[string]any) (*lease, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "lease version schema check failed")
	}
	importFunc, ok := leaseDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	return importFunc(source)
}

type leaseDeserializationFunc func(map[string]any) (*lease, error)

var leaseDeserializationFuncs = map[int]leaseDeserializationFunc{
	1: importLeaseV1,
}

func importLeaseV1(source map[string]any) (*lease, error) {
	fields, defaults := leaseV1Schema()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "lease v1 schema check failed")
	}

	return leaseV1(coerced.(map[string]any)), nil
}

func leaseV1Schema() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"name":   schema.String(),
		"holder": schema.String(),
		"start":  schema.Time(),
		"expiry": schema.Time(),
		"pinned": schema.Bool(),
	}
	defaults := schema.Defaults{
		"pinned": false,
	}
	return fields, defaults
}

func leaseV1(valid map[string]any) *lease {
	return &lease{
		Version_: 1,
		Name_:    valid["name"].(string),
		Holder_:  valid["holder"].(string),
		Start_:   valid["start"].(time.Time).UTC(),
		Expiry_:  valid["expiry"].(time.Time).UTC(),
		Pinned_:  valid["pinned"].(bool),
	}
}
