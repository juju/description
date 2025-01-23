// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package description

import (
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/juju/names/v6"
	"github.com/juju/schema"
	"github.com/rs/xid"
)

// Secret represents a secret.
type Secret interface {
	Id() string
	Version() int
	Description() string
	Label() string
	RotatePolicy() string
	AutoPrune() bool
	Owner() (names.Tag, error)
	Created() time.Time
	Updated() time.Time

	NextRotateTime() *time.Time

	ACL() map[string]SecretAccess
	Consumers() []SecretConsumer
	RemoteConsumers() []SecretRemoteConsumer

	Revisions() []SecretRevision
	LatestRevision() int
	LatestRevisionChecksum() string
	LatestExpireTime() *time.Time

	Validate() error
}

type secrets struct {
	Version  int       `yaml:"version"`
	Secrets_ []*secret `yaml:"secrets"`
}

type secret struct {
	ID_           string            `yaml:"id"`
	Version_      int               `yaml:"secret-version"`
	Description_  string            `yaml:"description"`
	Label_        string            `yaml:"label"`
	RotatePolicy_ string            `yaml:"rotate-policy,omitempty"`
	Owner_        string            `yaml:"owner"`
	AutoPrune_    bool              `yaml:"auto-prune,omitempty"`
	Created_      time.Time         `yaml:"create-time"`
	Updated_      time.Time         `yaml:"update-time"`
	Revisions_    []*secretRevision `yaml:"revisions"`

	ACL_             map[string]*secretAccess `yaml:"acl,omitempty"`
	Consumers_       []*secretConsumer        `yaml:"consumers,omitempty"`
	RemoteConsumers_ []*secretRemoteConsumer  `yaml:"remote-consumers,omitempty"`

	NextRotateTime_ *time.Time `yaml:"next-rotate-time,omitempty"`

	LatestRevisionChecksum_ string `yaml:"latest-revision-checksum"`

	// These are updated when revisions are set
	// and are not exported.
	LatestRevision_   int        `yaml:"-"`
	LatestExpireTime_ *time.Time `yaml:"-"`
}

// Revisions implements secret.
func (i *secret) Revisions() []SecretRevision {
	var result []SecretRevision
	for _, rev := range i.Revisions_ {
		result = append(result, rev)
	}
	return result
}

func (i *secret) setRevisions(args []SecretRevisionArgs) {
	i.Revisions_ = nil
	for _, arg := range args {
		rev := newSecretRevision(arg)
		i.Revisions_ = append(i.Revisions_, rev)
	}
}

func (i *secret) updateComputedFields() {
	if len(i.Revisions_) > 0 {
		i.LatestExpireTime_ = i.Revisions_[len(i.Revisions_)-1].ExpireTime_
	}
	for _, rev := range i.Revisions_ {
		if i.LatestRevision_ < rev.Number_ {
			i.LatestRevision_ = rev.Number_
		}
	}
	for x, consumer := range i.Consumers_ {
		consumer.LatestRevision_ = i.LatestRevision_
		i.Consumers_[x] = consumer
	}
	for x, consumer := range i.RemoteConsumers_ {
		consumer.LatestRevision_ = i.LatestRevision_
		i.RemoteConsumers_[x] = consumer
	}
}

// LatestExpireTime implements Secret.
func (i *secret) LatestExpireTime() *time.Time {
	return i.LatestExpireTime_
}

// LatestRevision implements Secret.
func (i *secret) LatestRevision() int {
	return i.LatestRevision_
}

// LatestRevisionChecksum implements Secret.
func (i *secret) LatestRevisionChecksum() string {
	return i.LatestRevisionChecksum_
}

// Id implements Secret.
func (i *secret) Id() string {
	return i.ID_
}

// Version implements Secret.
func (i *secret) Version() int {
	return i.Version_
}

// Description implements Secret.
func (i *secret) Description() string {
	return i.Description_
}

// Label implements Secret.
func (i *secret) Label() string {
	return i.Label_
}

// RotatePolicy implements Secret.
func (i *secret) RotatePolicy() string {
	return i.RotatePolicy_
}

// AutoPrune implements Secret.
func (i *secret) AutoPrune() bool {
	return i.AutoPrune_
}

// Owner implements Secret.
func (i *secret) Owner() (names.Tag, error) {
	if i.Owner_ == "" {
		return nil, nil
	}
	tag, err := names.ParseTag(i.Owner_)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return tag, nil
}

// Created implements Secret.
func (i *secret) Created() time.Time {
	return i.Created_
}

// Updated implements Secret.
func (i *secret) Updated() time.Time {
	return i.Updated_
}

// ACL implements secret.
func (i *secret) ACL() map[string]SecretAccess {
	var result map[string]SecretAccess
	if len(i.ACL_) == 0 {
		return result
	}
	result = make(map[string]SecretAccess)
	for k, v := range i.ACL_ {
		result[k] = v
	}
	return result
}

// NextRotateTime implements Secret.
func (i *secret) NextRotateTime() *time.Time {
	return i.NextRotateTime_
}

// Consumers implements secret.
func (i *secret) Consumers() []SecretConsumer {
	var result []SecretConsumer
	for _, c := range i.Consumers_ {
		result = append(result, c)
	}
	return result
}

func (i *secret) setConsumers(args []SecretConsumerArgs) {
	i.Consumers_ = nil
	for _, arg := range args {
		c := newSecretConsumer(arg)
		i.Consumers_ = append(i.Consumers_, c)
	}
}

// RemoteConsumers implements secret.
func (i *secret) RemoteConsumers() []SecretRemoteConsumer {
	var result []SecretRemoteConsumer
	for _, c := range i.RemoteConsumers_ {
		result = append(result, c)
	}
	return result
}

func (i *secret) setRemoteConsumers(args []SecretRemoteConsumerArgs) {
	i.RemoteConsumers_ = nil
	for _, arg := range args {
		c := newSecretRemoteConsumer(arg)
		i.RemoteConsumers_ = append(i.RemoteConsumers_, c)
	}
}

// SecretArgs is an argument struct used to create a
// new internal secret type that supports the secret interface.
type SecretArgs struct {
	ID              string
	Version         int
	Description     string
	Label           string
	RotatePolicy    string
	Owner           names.Tag
	Created         time.Time
	Updated         time.Time
	Revisions       []SecretRevisionArgs
	ACL             map[string]SecretAccessArgs
	Consumers       []SecretConsumerArgs
	RemoteConsumers []SecretRemoteConsumerArgs

	NextRotateTime         *time.Time
	LatestRevisionChecksum string
	AutoPrune              bool
}

func newSecret(args SecretArgs) *secret {
	secret := &secret{
		ID_:                     args.ID,
		Version_:                args.Version,
		Description_:            args.Description,
		Label_:                  args.Label,
		RotatePolicy_:           args.RotatePolicy,
		AutoPrune_:              args.AutoPrune,
		LatestRevisionChecksum_: args.LatestRevisionChecksum,
		Created_:                args.Created.UTC(),
		Updated_:                args.Updated.UTC(),
		ACL_:                    newSecretAccess(args.ACL),
	}
	if args.NextRotateTime != nil {
		next := args.NextRotateTime.UTC()
		secret.NextRotateTime_ = &next
	}
	if args.Owner != nil {
		secret.Owner_ = args.Owner.String()
	}
	secret.setRevisions(args.Revisions)
	secret.setConsumers(args.Consumers)
	secret.setRemoteConsumers(args.RemoteConsumers)
	secret.updateComputedFields()
	return secret
}

// Validate implements Secret.
func (i *secret) Validate() error {
	if i.ID_ == "" {
		return errors.NotValidf("secret missing id")
	}
	if _, err := xid.FromString(i.ID_); err != nil {
		return errors.Wrap(err, errors.NotValidf("secret ID %q", i.ID_))
	}
	if _, err := i.Owner(); err != nil {
		return errors.Wrap(err, errors.NotValidf("secret %q invalid owner", i.ID_))
	}
	for tag := range i.ACL_ {
		if _, err := names.ParseTag(tag); err != nil {
			return errors.Wrap(err, errors.NotValidf("secret %q invalid access entity", i.ID_))
		}
	}
	for _, consumer := range i.Consumers_ {
		if _, err := names.ParseTag(consumer.Consumer_); err != nil {
			return errors.Wrap(err, errors.NotValidf("secret %q invalid consumer", i.ID_))
		}
	}
	for _, consumer := range i.RemoteConsumers_ {
		if _, err := names.ParseTag(consumer.Consumer_); err != nil {
			return errors.Wrap(err, errors.NotValidf("secret %q invalid remote consumer", i.ID_))
		}
	}
	return nil
}

func importSecrets(source map[string]interface{}) ([]*secret, error) {
	checker := versionedChecker("secrets")
	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "secrets version schema check failed")
	}
	valid := coerced.(map[string]interface{})

	version := int(valid["version"].(int64))
	sourceList := valid["secrets"].([]interface{})
	return importSecretList(sourceList, version)
}

func importSecretList(sourceList []interface{}, version int) ([]*secret, error) {
	getFields, ok := secretFieldsFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}

	result := make([]*secret, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for secret %d, %T", i, value)
		}
		secret, err := importSecret(source, version, getFields)
		if err != nil {
			return nil, errors.Annotatef(err, "secret %d", i)
		}
		result = append(result, secret)
	}
	return result, nil
}

var secretFieldsFuncs = map[int]fieldsFunc{
	1: secretV1Fields,
	2: secretV2Fields,
}

func secretV1Fields() (schema.Fields, schema.Defaults) {
	fields := schema.Fields{
		"id":               schema.String(),
		"secret-version":   schema.Int(),
		"description":      schema.String(),
		"label":            schema.String(),
		"rotate-policy":    schema.String(),
		"auto-prune":       schema.Bool(),
		"owner":            schema.String(),
		"create-time":      schema.Time(),
		"update-time":      schema.Time(),
		"next-rotate-time": schema.Time(),
		"revisions":        schema.List(schema.Any()),
		"acl":              schema.Map(schema.String(), schema.Any()),
		"consumers":        schema.List(schema.Any()),
		"remote-consumers": schema.List(schema.Any()),
	}
	// Some values don't have to be there.
	defaults := schema.Defaults{
		"rotate-policy":    schema.Omit,
		"auto-prune":       schema.Omit,
		"next-rotate-time": schema.Omit,
		"acl":              schema.Omit,
		"consumers":        schema.Omit,
		"remote-consumers": schema.Omit,
	}
	return fields, defaults
}

func secretV2Fields() (schema.Fields, schema.Defaults) {
	fields, defaults := secretV1Fields()
	fields["latest-revision-checksum"] = schema.String()
	defaults["latest-revision-checksum"] = schema.Omit
	return fields, defaults
}

func importSecret(source map[string]interface{}, importVersion int, fieldFunc func() (schema.Fields, schema.Defaults)) (*secret, error) {
	fields, defaults := fieldFunc()
	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "secret v%d schema check failed", importVersion)
	}
	valid := coerced.(map[string]interface{})
	secret := &secret{
		ID_:             valid["id"].(string),
		Version_:        int(valid["secret-version"].(int64)),
		Description_:    valid["description"].(string),
		Label_:          valid["label"].(string),
		Owner_:          valid["owner"].(string),
		Created_:        valid["create-time"].(time.Time).UTC(),
		Updated_:        valid["update-time"].(time.Time).UTC(),
		NextRotateTime_: fieldToTimePtr(valid, "next-rotate-time"),
	}

	if policy, ok := valid["rotate-policy"].(string); ok {
		secret.RotatePolicy_ = policy
	}

	if importVersion >= 2 {
		if checksum, ok := valid["latest-revision-checksum"].(string); ok {
			secret.LatestRevisionChecksum_ = checksum
		}
	}

	// This should be in a v2 schema but it's already also in v1.
	if autoPrune, ok := valid["auto-prune"].(bool); ok {
		secret.AutoPrune_ = autoPrune
	}

	secretACL, err := importSecretAccess(valid, importVersion)
	if err != nil {
		return nil, errors.Trace(err)
	}
	secret.ACL_ = secretACL

	revisionList, err := importSecretRevisions(valid, importVersion)
	if err != nil {
		return nil, errors.Trace(err)
	}
	secret.Revisions_ = revisionList

	consumerList, err := importSecretConsumers(valid, importVersion)
	if err != nil {
		return nil, errors.Trace(err)
	}
	secret.Consumers_ = consumerList

	remoteConsumerList, err := importSecretRemoteConsumers(valid, importVersion)
	if err != nil {
		return nil, errors.Trace(err)
	}
	secret.RemoteConsumers_ = remoteConsumerList

	secret.updateComputedFields()
	return secret, nil
}

// SecretAccess represents a secret ACL entry.
type SecretAccess interface {
	Scope() string
	Role() string
}

type secretAccess struct {
	Scope_ string `yaml:"scope"`
	Role_  string `yaml:"role"`
}

// SecretAccessArgs is an argument struct used to create a
// new internal secret access type that supports the secret access interface.
type SecretAccessArgs struct {
	Scope string
	Role  string
}

func newSecretAccess(args map[string]SecretAccessArgs) map[string]*secretAccess {
	var result map[string]*secretAccess
	if len(args) == 0 {
		return result
	}
	result = make(map[string]*secretAccess)
	for subject, access := range args {
		result[subject] = &secretAccess{
			Scope_: access.Scope,
			Role_:  access.Role,
		}
	}
	return result
}

// Scope implements SecretAccess.
func (i *secretAccess) Scope() string {
	return i.Scope_
}

// Role implements SecretAccess.
func (i *secretAccess) Role() string {
	return i.Role_
}

func importSecretAccess(source map[string]interface{}, version int) (map[string]*secretAccess, error) {
	importFunc, ok := secretAccessDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList, ok := source["acl"].(map[interface{}]interface{})
	if !ok {
		return nil, nil
	}
	return importSecretAccessMap(sourceList, importFunc)
}

func importSecretAccessMap(sourceMap map[interface{}]interface{}, importFunc secretAccessDeserializationFunc) (map[string]*secretAccess, error) {
	result := make(map[string]*secretAccess)
	for subject, access := range sourceMap {
		source, ok := access.(map[interface{}]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for subject %v, %T", subject, access)
		}
		access, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "access for %v", subject)
		}
		result[fmt.Sprintf("%v", subject)] = access
	}
	return result, nil
}

type secretAccessDeserializationFunc func(map[interface{}]interface{}) (*secretAccess, error)

var secretAccessDeserializationFuncs = map[int]secretAccessDeserializationFunc{
	1: importSecretAccessV1,
	2: importSecretAccessV2,
}

func importSecretAccessV2(source map[interface{}]interface{}) (*secretAccess, error) {
	return importSecretAccessV1(source)
}

func importSecretAccessV1(source map[interface{}]interface{}) (*secretAccess, error) {
	fields := schema.Fields{
		"scope": schema.String(),
		"role":  schema.String(),
	}

	checker := schema.FieldMap(fields, nil) // no defaults

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "revisions v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	access := &secretAccess{
		Scope_: valid["scope"].(string),
		Role_:  valid["role"].(string),
	}
	return access, nil
}

// SecretConsumer represents a secret consumer.
type SecretConsumer interface {
	Consumer() (names.Tag, error)
	Label() string
	CurrentRevision() int
	LatestRevision() int
}

type secretConsumer struct {
	Consumer_        string `yaml:"consumer"`
	Label_           string `yaml:"label"`
	CurrentRevision_ int    `yaml:"current-revision"`

	// Updated when added to a secret
	// but not exported.
	LatestRevision_ int `yaml:"-"`
}

// SecretConsumerArgs is an argument struct used to create a
// new internal secret consumer type that supports the secret consumer interface.
type SecretConsumerArgs struct {
	Consumer        names.Tag
	Label           string
	CurrentRevision int
}

func newSecretConsumer(args SecretConsumerArgs) *secretConsumer {
	return &secretConsumer{
		Consumer_:        args.Consumer.String(),
		Label_:           args.Label,
		CurrentRevision_: args.CurrentRevision,
	}
}

// Consumer implements SecretConsumer.
func (i *secretConsumer) Consumer() (names.Tag, error) {
	if i.Consumer_ == "" {
		return nil, nil
	}
	tag, err := names.ParseTag(i.Consumer_)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return tag, nil
}

// Label implements SecretConsumer.
func (i *secretConsumer) Label() string {
	return i.Label_
}

// CurrentRevision implements SecretConsumer.
func (i *secretConsumer) CurrentRevision() int {
	return i.CurrentRevision_
}

// LatestRevision implements SecretConsumer.
func (i *secretConsumer) LatestRevision() int {
	return i.LatestRevision_
}

func importSecretConsumers(source map[string]interface{}, version int) ([]*secretConsumer, error) {
	importFunc, ok := secretConsumerDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList, ok := source["consumers"].([]interface{})
	if !ok {
		return nil, nil
	}
	return importSecretConsumersList(sourceList, importFunc)
}

func importSecretConsumersList(sourceList []interface{}, importFunc secretConsumerDeserializationFunc) ([]*secretConsumer, error) {
	result := make([]*secretConsumer, 0, len(sourceList))
	for i, consumer := range sourceList {
		source, ok := consumer.(map[interface{}]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for consumer %d, %T", i, consumer)
		}
		consumer, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "consumer %v", i)
		}
		result = append(result, consumer)
	}
	return result, nil
}

type secretConsumerDeserializationFunc func(map[interface{}]interface{}) (*secretConsumer, error)

var secretConsumerDeserializationFuncs = map[int]secretConsumerDeserializationFunc{
	1: importSecretConsumerV1,
	2: importSecretConsumerV2,
}

func importSecretConsumerV2(source map[interface{}]interface{}) (*secretConsumer, error) {
	return importSecretConsumerV1(source)
}

func importSecretConsumerV1(source map[interface{}]interface{}) (*secretConsumer, error) {
	fields := schema.Fields{
		"consumer":         schema.String(),
		"label":            schema.String(),
		"current-revision": schema.Int(),
	}
	defaults := schema.Defaults{
		"label": schema.Omit,
	}

	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "consumers v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	consumer := &secretConsumer{
		Consumer_:        valid["consumer"].(string),
		Label_:           valid["label"].(string),
		CurrentRevision_: int(valid["current-revision"].(int64)),
	}
	return consumer, nil
}

// SecretRevision represents a secret revision.
type SecretRevision interface {
	Number() int
	Created() time.Time
	Updated() time.Time
	Obsolete() bool
	PendingDelete() bool

	ExpireTime() *time.Time
	ValueRef() SecretValueRef
	Content() map[string]string
}

type secretRevision struct {
	Number_        int       `yaml:"number"`
	Created_       time.Time `yaml:"create-time"`
	Updated_       time.Time `yaml:"update-time"`
	Obsolete_      bool      `yaml:"obsolete,omitempty"`
	PendingDelete_ bool      `yaml:"pending-delete,omitempty"`

	Content_    map[string]string `yaml:"content,omitempty"`
	ValueRef_   *secretValueRef   `yaml:"value-ref,omitempty"`
	ExpireTime_ *time.Time        `yaml:"expire-time,omitempty"`
}

// SecretValueRef represents an external secret revision.
type SecretValueRef interface {
	BackendID() string
	RevisionID() string
}

type secretValueRef struct {
	BackendId_  string `yaml:"backend-id"`
	RevisionId_ string `yaml:"revision-id"`
}

// SecretRevisionArgs is an argument struct used to create a
// new internal secret revision type that supports the secret revision interface.
type SecretRevisionArgs struct {
	Number        int
	Created       time.Time
	Updated       time.Time
	Obsolete      bool
	PendingDelete bool

	Content    map[string]string
	ValueRef   *SecretValueRefArgs
	ExpireTime *time.Time
}

// SecretValueRefArgs is an argument struct used to create a
// new internal secret value reference type.
type SecretValueRefArgs struct {
	BackendID  string
	RevisionID string
}

func newSecretRevision(args SecretRevisionArgs) *secretRevision {
	revision := &secretRevision{
		Number_:        args.Number,
		Created_:       args.Created.UTC(),
		Updated_:       args.Updated.UTC(),
		Obsolete_:      args.Obsolete,
		PendingDelete_: args.PendingDelete,
		Content_:       args.Content,
	}
	if args.ExpireTime != nil {
		expire := args.ExpireTime.UTC()
		revision.ExpireTime_ = &expire
	}
	if args.ValueRef != nil {
		revision.ValueRef_ = &secretValueRef{
			BackendId_:  args.ValueRef.BackendID,
			RevisionId_: args.ValueRef.RevisionID,
		}
	}
	return revision
}

// Number implements SecretRevision.
func (i *secretRevision) Number() int {
	return i.Number_
}

// Created implements SecretRevision.
func (i *secretRevision) Created() time.Time {
	return i.Created_
}

// Updated implements SecretRevision.
func (i *secretRevision) Updated() time.Time {
	return i.Updated_
}

// Obsolete implements SecretRevision.
func (i *secretRevision) Obsolete() bool {
	return i.Obsolete_
}

// PendingDelete implements SecretRevision.
func (i *secretRevision) PendingDelete() bool {
	return i.PendingDelete_
}

// ExpireTime implements SecretRevision.
func (i *secretRevision) ExpireTime() *time.Time {
	return i.ExpireTime_
}

// ValueRef implements SecretRevision.
func (i *secretRevision) ValueRef() SecretValueRef {
	return i.ValueRef_
}

// Content implements SecretRevision.
func (i *secretRevision) Content() map[string]string {
	return i.Content_
}

// BackendID implements SecretValueRef.
func (i *secretValueRef) BackendID() string {
	return i.BackendId_
}

// RevisionID implements SecretValueRef.
func (i *secretValueRef) RevisionID() string {
	return i.RevisionId_
}

func importSecretRevisions(source map[string]interface{}, version int) ([]*secretRevision, error) {
	importFunc, ok := secretRevisionRangeDeserializationFuncs[version]
	if !ok {
		return nil, errors.NotValidf("version %d", version)
	}
	sourceList, ok := source["revisions"].([]interface{})
	if !ok {
		return nil, nil
	}
	return importSecretRevisionList(sourceList, importFunc)
}

func importSecretRevisionList(sourceList []interface{}, importFunc secretRevisionDeserializationFunc) ([]*secretRevision, error) {
	result := make([]*secretRevision, 0, len(sourceList))
	for i, value := range sourceList {
		source, ok := value.(map[interface{}]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected value for revision %d, %T", i, value)
		}
		revisions, err := importFunc(source)
		if err != nil {
			return nil, errors.Annotatef(err, "revision %d", i)
		}
		result = append(result, revisions)
	}
	return result, nil
}

type secretRevisionDeserializationFunc func(map[interface{}]interface{}) (*secretRevision, error)

var secretRevisionRangeDeserializationFuncs = map[int]secretRevisionDeserializationFunc{
	1: importSecretRevisionV1,
	2: importSecretRevisionV2,
}

func importSecretRevisionV2(source map[interface{}]interface{}) (*secretRevision, error) {
	return importSecretRevisionV1(source)
}

func importSecretRevisionV1(source map[interface{}]interface{}) (*secretRevision, error) {
	fields := schema.Fields{
		"number":         schema.Int(),
		"create-time":    schema.Time(),
		"update-time":    schema.Time(),
		"obsolete":       schema.Bool(),
		"pending-delete": schema.Bool(),
		"expire-time":    schema.Time(),
		"value-ref":      schema.StringMap(schema.Any()),
		"content":        schema.StringMap(schema.Any()),
	}
	defaults := schema.Defaults{
		"value-ref":      schema.Omit,
		"content":        schema.Omit,
		"expire-time":    schema.Omit,
		"obsolete":       false,
		"pending-delete": false,
	}

	checker := schema.FieldMap(fields, defaults)

	coerced, err := checker.Coerce(source, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "revisions v1 schema check failed")
	}
	valid := coerced.(map[string]interface{})
	// From here we know that the map returned from the schema coercion
	// contains fields of the right type.

	rev := &secretRevision{
		Number_:        int(valid["number"].(int64)),
		Created_:       valid["create-time"].(time.Time).UTC(),
		Updated_:       valid["update-time"].(time.Time).UTC(),
		Obsolete_:      valid["obsolete"].(bool),
		PendingDelete_: valid["pending-delete"].(bool),
		ExpireTime_:    fieldToTimePtr(valid, "expire-time"),
		Content_:       convertToStringMap(valid["content"]),
	}
	valueRefMap := convertToStringMap(valid["value-ref"])
	if valueRefMap == nil {
		return rev, nil
	}
	rev.ValueRef_ = &secretValueRef{
		BackendId_:  valueRefMap["backend-id"],
		RevisionId_: valueRefMap["revision-id"],
	}
	if rev.ValueRef_.BackendId_ == "" || rev.ValueRef_.RevisionId_ == "" {
		return nil, errors.Errorf("incomplete secret value ref for revision %d", rev.Number_)
	}
	return rev, nil
}
