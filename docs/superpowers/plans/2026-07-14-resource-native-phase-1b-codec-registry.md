# Resource-Native Phase 1B Codec And Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `subagent-driven-development` task-by-task. Steps use checkbox (`- [ ]`).

**Goal:** Add strict JSON/YAML resource codecs and a typed Spec registry.

**Architecture:** Codecs produce the same canonical `Manifest`. The registry
owns strict Kind-specific Spec decoding, while business services remain outside
the package.

**Tech Stack:** Go 1.25, `encoding/json`, `gopkg.in/yaml.v3`, testify.

---

### Task 1: Strict JSON Codec

**Files:**
- Create: `backend/internal/domain/orchestrationresource/json_codec.go`
- Test: `backend/internal/domain/orchestrationresource/json_codec_test.go`

- [x] **Step 1: Write failing strictness tests**

Test valid decoding, duplicate object keys at any depth, nesting over 64,
trailing documents, unknown top-level and metadata fields, non-object Spec,
and server-owned submission fields.

- [x] **Step 2: Implement**

```go
func DecodeJSONSubmission(source []byte) (Manifest, error)
func EncodeJSON(resource Manifest) ([]byte, error)
```

Enforce the 1 MiB source and 64-level nesting limits and reject duplicate
object keys before typed decoding. Use `json.Decoder.DisallowUnknownFields()`,
reject trailing tokens, validate submission, and encode normalized JSON with
one trailing newline.

### Task 2: Bounded YAML Codec

**Files:**
- Create: `backend/internal/domain/orchestrationresource/yaml_codec.go`
- Create: `backend/internal/domain/orchestrationresource/yaml_node_validation.go`
- Test: `backend/internal/domain/orchestrationresource/yaml_codec_test.go`

- [ ] **Step 1: Write YAML attack and parity tests**

Cover duplicate keys, aliases, merge/custom tags, multiple documents, unknown
envelope fields, unknown metadata fields, depth over 64, more than 10,000
nodes, source over 1 MiB, and JSON/YAML normalized equality.

- [ ] **Step 2: Implement bounded parsing**

```go
const maxManifestBytes = 1 << 20
const maxYAMLManifestBytes = 256 << 10
const maxYAMLDepth = 64
const maxYAMLNodes = 10_000
const maxYAMLLineBytes = 64 << 10

func DecodeYAMLSubmission(source []byte) (Manifest, error)
func EncodeYAML(resource Manifest) ([]byte, error)
```

YAML is the human editing format and is limited to 256 KiB plus 64 KiB per
physical line; larger machine-generated documents use JSON and large content is
carried by referenced resources. Parse one `yaml.Node`, reject duplicate
mapping keys, enforce exact node/depth limits, reject aliases, anchors,
merge/custom tags, convert scalar tags explicitly to JSON-compatible values,
and reuse strict JSON submission decoding. Do not silently coerce non-string
mapping keys. Numeric tags must match their JSON integer/float lexemes;
timestamp, binary, and other YAML-only semantics are rejected. Encoding counts
nodes directly from JSON tokens before building a generic tree and uses a
256 KiB limited writer. It must preserve scalar content exactly and must not
trim document suffixes after `yaml.Encoder` writes them.

### Task 3: Typed Spec Registry

**Files:**
- Create: `backend/internal/domain/orchestrationresource/schema_registry.go`
- Test: `backend/internal/domain/orchestrationresource/schema_registry_test.go`

- [ ] **Step 1: Write registry tests**

Use this test-only schema:

```go
type registrySpec struct {
	ModelRef      Reference       `json:"modelRef"`
	CredentialRef SecretReference `json:"credentialRef"`
}
```

Tests cover duplicate registration, unknown Kind, unknown Spec fields,
Secret `value` rejection, and successful typed decoding.

- [ ] **Step 2: Implement**

```go
type Schema struct {
	NewSpec  func() any
	Validate func(Metadata, any) error
}

type Registry struct {
	schemas map[TypeMeta]Schema
}

func NewRegistry() *Registry
func (registry *Registry) Register(meta TypeMeta, schema Schema) error
func (registry *Registry) DecodeAndValidate(manifest Manifest) (any, error)
```

Decode Spec with `DisallowUnknownFields`, reject trailing tokens, then call the
typed validator with manifest Metadata so namespace-scoped references can be
validated. Registration validates TypeMeta, freezes a pointer-to-struct root
type, and rejects duplicates. Decoding creates an independent root instance,
caches immutable reflected field sets, and is safe alongside registration.

### Task 4: Canonical Round Trip

**Files:**
- Create: `backend/internal/domain/orchestrationresource/round_trip_test.go`
- Modify: `docs/superpowers/plans/2026-07-14-resource-native-orchestration-goal.md`

- [x] **Step 1: Prove JSON/YAML parity**

The same WorkerTemplate-shaped test manifest must produce identical TypeMeta,
Metadata, canonical Spec JSON, and Reference values through both codecs.

- [ ] **Step 2: Run terminal verification**

```bash
go test ./backend/internal/domain/orchestrationresource -count=1
go test ./backend/internal/domain/workerspec ./backend/internal/service/workerspec \
  ./backend/internal/service/workercreation -count=1
git diff --check
```

- [ ] **Step 3: Record Phase 1 evidence**

Mark Phase 1 complete only when all codec attack tests and registry strictness
tests pass without fallback or ignored fields.
