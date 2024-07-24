// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/juju/os/v2/series"
	"github.com/juju/schema"
)

// CharmOriginArgs is an argument struct used to add information about the
// charm origin.
type CharmOriginArgs struct {
	Source   string
	ID       string
	Hash     string
	Revision int
	Channel  string
	Platform string
}

func newCharmOrigin(args CharmOriginArgs) *charmOrigin {
	return &charmOrigin{
		Version_:  2,
		Source_:   args.Source,
		ID_:       args.ID,
		Hash_:     args.Hash,
		Revision_: args.Revision,
		Channel_:  args.Channel,
		Platform_: args.Platform,
	}
}

// charmOrigin represents the origin of a charm.
type charmOrigin struct {
	Version_  int    `yaml:"version"`
	Source_   string `yaml:"source"`
	ID_       string `yaml:"id"`
	Hash_     string `yaml:"hash"`
	Revision_ int    `yaml:"revision"`
	Channel_  string `yaml:"channel"`
	Platform_ string `yaml:"platform"`
}

func platformFromSeries(s string) (string, error) {
	if s == "" {
		return "", errors.New("cannot convert empty series to a platform")
	}

	os, err := series.GetOSFromSeries(s)
	if err != nil {
		return "", fmt.Errorf("extracting os from series %q: %w", s, err)
	}
	version, err := series.SeriesVersion(s)
	if err != nil {
		return "", fmt.Errorf("extracting os version from series %q: %w", s, err)
	}

	return fmt.Sprintf("unknown/%s/%s", strings.ToLower(os.String()), version), nil
}

// Source implements CharmOrigin.
func (a *charmOrigin) Source() string {
	return a.Source_
}

// ID implements CharmOrigin.
func (a *charmOrigin) ID() string {
	return a.ID_
}

// Hash implements CharmOrigin.
func (a *charmOrigin) Hash() string {
	return a.Hash_
}

// Revision implements CharmOrigin.
func (a *charmOrigin) Revision() int {
	return a.Revision_
}

// Channel implements CharmOrigin.
func (a *charmOrigin) Channel() string {
	return a.Channel_
}

// Platform implements CharmOrigin.
func (a *charmOrigin) Platform() string {
	return a.Platform_
}

func importCharmOrigin(source map[string]interface{}) (*charmOrigin, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "charmOrigin version schema check failed")
	}

	importFunc, ok := charmOriginDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	return importFunc(source)
}

type charmOriginDeserializationFunc func(map[string]interface{}) (*charmOrigin, error)

var charmOriginDeserializationFuncs = map[int]charmOriginDeserializationFunc{
	1: importCharmOriginV1,
	2: importCharmOriginV2,
}

func importCharmOriginV1(source map[string]interface{}) (*charmOrigin, error) {
	return importCharmOriginVersion(source, 1)
}

func importCharmOriginV2(source map[string]interface{}) (*charmOrigin, error) {
	return importCharmOriginVersion(source, 2)
}

func importCharmOriginVersion(source map[string]interface{}, importVersion int) (*charmOrigin, error) {
	fields := schema.Fields{
		"source":   schema.String(),
		"id":       schema.String(),
		"hash":     schema.String(),
		"revision": schema.Int(),
		"channel":  schema.String(),
		"platform": schema.String(),
	}
	defaults := schema.Defaults{
		"source":   "unknown",
		"id":       schema.Omit,
		"hash":     schema.Omit,
		"revision": schema.Omit,
		"channel":  schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charmOrigin v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	var revision int
	switch t := valid["revision"].(type) {
	case int:
		revision = t
	case int64:
		revision = int(t)
	default:
		return nil, errors.Errorf("unexpected revision type %T", valid["revision"])
	}

	platform := valid["platform"].(string)
	if importVersion < 2 {
		parts := strings.Split(platform, "/")
		if len(parts) < 3 {
			return nil, errors.NotValidf("platform %q", platform)
		}
		pSeries := parts[2]
		vers, err := series.SeriesVersion(pSeries)
		if err != nil {
			return nil, errors.NotValidf("platform series %q", pSeries)
		}
		parts[2] = vers
		parts = append(parts, "stable")
		platform = strings.Join(parts, "/")
	}

	return &charmOrigin{
		Version_:  2,
		Source_:   valid["source"].(string),
		ID_:       valid["id"].(string),
		Hash_:     valid["hash"].(string),
		Revision_: revision,
		Channel_:  valid["channel"].(string),
		Platform_: platform,
	}, nil
}
