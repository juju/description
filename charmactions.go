// Copyright 2024 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"github.com/juju/errors"
	"github.com/juju/schema"
)

type CharmActionsArgs struct {
	Actions map[string]CharmAction
}

func newCharmActions(args CharmActionsArgs) *charmActions {
	actions := make(map[string]charmAction)
	if args.Actions != nil {
		for name, a := range args.Actions {
			actions[name] = charmAction{
				Description_:    a.Description(),
				Parallel_:       a.Parallel(),
				ExecutionGroup_: a.ExecutionGroup(),
				Parameters_:     a.Parameters(),
			}
		}
	}

	return &charmActions{
		Version_: 1,
		Actions_: actions,
	}
}

type charmActions struct {
	Version_ int                    `yaml:"version"`
	Actions_ map[string]charmAction `yaml:"actions"`
}

func (a *charmActions) Actions() map[string]CharmAction {
	actions := make(map[string]CharmAction)
	for i, b := range a.Actions_ {
		actions[i] = b
	}
	return actions
}

func importCharmActions(source map[string]interface{}) (*charmActions, error) {
	version, err := getVersion(source)
	if err != nil {
		return nil, errors.Annotate(err, "charmActions version schema check failed")
	}

	importFunc, ok := charmActionsDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	return importFunc(source)
}

var charmActionsDeserializationFuncs = map[int]func(map[string]interface{}) (*charmActions, error){
	1: importCharmActionsV1,
}

func importCharmActionsV1(source map[string]interface{}) (*charmActions, error) {
	return importCharmActionsVersion(source, 1)
}

func importCharmActionsVersion(source map[string]interface{}, version int) (*charmActions, error) {
	fields := schema.Fields{
		"actions": schema.StringMap(schema.Any()),
	}
	defaults := schema.Defaults{
		"actions": schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "charmActions v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})

	var actions map[string]charmAction
	if valid["actions"] != nil {
		actions = make(map[string]charmAction, len(valid["actions"].(map[string]interface{})))
		for name, v := range valid["actions"].(map[string]interface{}) {
			var err error
			actions[name], err = importCharmAction(v)
			if err != nil {
				return nil, errors.Annotate(err, "charmActions actions schema check failed")
			}
		}
	}

	return &charmActions{
		Version_: 1,
		Actions_: actions,
	}, nil
}

func importCharmAction(source interface{}) (charmAction, error) {
	fields := schema.Fields{
		"description":     schema.String(),
		"parallel":        schema.Bool(),
		"execution-group": schema.String(),
		"parameters":      schema.StringMap(schema.Any()),
	}
	defaults := schema.Defaults{
		"description":     schema.Omit,
		"parallel":        false,
		"execution-group": schema.Omit,
		"parameters":      schema.Omit,
	}
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return charmAction{}, errors.Annotatef(err, "charmAction schema check failed")
	}
	valid := coerced.(map[string]interface{})

	var parameters map[string]interface{}
	if valid["parameters"] != nil {
		parameters = valid["parameters"].(map[string]interface{})
	}

	return charmAction{
		Description_:    valid["description"].(string),
		Parallel_:       valid["parallel"].(bool),
		ExecutionGroup_: valid["execution-group"].(string),
		Parameters_:     parameters,
	}, nil
}

type charmAction struct {
	Description_    string                 `yaml:"description"`
	Parallel_       bool                   `yaml:"parallel"`
	ExecutionGroup_ string                 `yaml:"execution-group"`
	Parameters_     map[string]interface{} `yaml:"parameters"`
}

// Description returns the description of the action.
func (a charmAction) Description() string {
	return a.Description_
}

// Parallel returns whether the action can be run in parallel.
func (a charmAction) Parallel() bool {
	return a.Parallel_
}

// ExecutionGroup returns the execution group of the action.
func (a charmAction) ExecutionGroup() string {
	return a.ExecutionGroup_
}

// Parameters returns the parameters of the action.
func (a charmAction) Parameters() map[string]interface{} {
	return a.Parameters_
}
