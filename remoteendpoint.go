// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package description

// RemoteEndpoint represents a connection point that can be related to
// another application.
type RemoteEndpoint interface {
	Name() string
	Role() string
	Interface() string
	Limit() int
	Scope() string
}

type remoteEndpoints struct {
	Version   int               `yaml:"version"`
	Endpoints []*remoteEndpoint `yaml:"endpoints"`
}

type remoteEndpoint struct {
	Name_      string `yaml:"name"`
	Role_      string `yaml:"role"`
	Interface_ string `yaml:"interface"`
	Limit_     int    `yaml:"limit"`
	Scope_     string `yaml:"scope"`
}

// RemoteEndpointArgs is an argument struct used to add a remote
// endpoint to a remote application.
type RemoteEndpointArgs struct {
	Name      string
	Role      string
	Interface string
	Limit     int
	Scope     string
}

func newRemoteEndpoint(args RemoteEndpointArgs) *remoteEndpoint {
	return &remoteEndpoint{
		Name_:      args.Name,
		Role_:      args.Role,
		Interface_: args.Interface,
		Limit_:     args.Limit,
		Scope_:     args.Scope,
	}
}

// Name implements RemoteEndpoint.
func (e *remoteEndpoint) Name() string {
	return e.Name_
}

// Role implements RemoteEndpoint.
func (e *remoteEndpoint) Role() string {
	return e.Role_
}

// Interface implements RemoteEndpoint.
func (e *remoteEndpoint) Interface() string {
	return e.Interface_
}

// Limit implements RemoteEndpoint.
func (e *remoteEndpoint) Limit() int {
	return e.Limit_
}

// Scope implements RemoteEndpoint.
func (e *remoteEndpoint) Scope() string {
	return e.Scope_
}
