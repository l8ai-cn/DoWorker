# Resource-Native Phase 1A Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `subagent-driven-development` task-by-task. Steps use checkbox (`- [ ]`).

**Goal:** Add the strict resource metadata, reference, condition, and
submission contracts without persistence or domain actions.

**Architecture:** Create a focused `orchestrationresource` domain package.
Authoring manifests and resolved resources use different validation methods so
server-owned fields cannot enter Apply input.
**Tech Stack:** Go 1.25, `slugkit`, testify.

---

### Task 1: API Version And Kind

**Files:**
- Create: `backend/internal/domain/orchestrationresource/type_meta.go`
- Test: `backend/internal/domain/orchestrationresource/type_meta_test.go`

- [x] **Step 1: Write failing validation tests**

Cover the exact supported version and Kind grammar:

```go
func TestTypeMetaAcceptsInitialAPIGroup(t *testing.T) {
	meta := TypeMeta{APIVersion: "agentsmesh.io/v1alpha1", Kind: "WorkerTemplate"}
	require.NoError(t, meta.Validate())
}

func TestTypeMetaRejectsUnknownVersionAndMalformedKind(t *testing.T) {
	require.Error(t, (TypeMeta{APIVersion: "v1", Kind: "worker-template"}).Validate())
}
```

- [x] **Step 2: Implement the contract**

```go
const APIVersionV1Alpha1 = "agentsmesh.io/v1alpha1"

type TypeMeta struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
}

func (meta TypeMeta) Validate() error
```

`Kind` must match `^[A-Z][A-Za-z0-9]{1,99}$`.
- [x] **Step 3: Verify**

```bash
go test ./backend/internal/domain/orchestrationresource \
  -run TestTypeMeta -count=1
```

### Task 2: Metadata And Submission Boundary

**Files:**
- Create: `backend/internal/domain/orchestrationresource/metadata.go`
- Create: `backend/internal/domain/orchestrationresource/manifest.go`
- Test: `backend/internal/domain/orchestrationresource/manifest_test.go`

- [x] **Step 1: Write failing metadata tests**

```go
func TestManifestSubmissionRejectsServerOwnedFields(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Metadata.UID = "res_123"
	require.ErrorIs(t, manifest.ValidateSubmission(), ErrServerManagedField)
}

func TestManifestSubmissionRejectsStatus(t *testing.T) {
	manifest := validManifestForTest()
	manifest.Status = json.RawMessage(`{"ready":true}`)
	require.ErrorIs(t, manifest.ValidateSubmission(), ErrServerManagedField)
}
```

- [x] **Step 2: Implement metadata**

```go
type Metadata struct {
	Name            slugkit.Slug    `json:"name" yaml:"name"`
	Namespace       slugkit.Slug    `json:"namespace" yaml:"namespace"`
	DisplayName     string          `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Labels          map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	UID             string          `json:"uid,omitempty" yaml:"uid,omitempty"`
	ResourceVersion string          `json:"resourceVersion,omitempty" yaml:"resourceVersion,omitempty"`
	Generation      int64           `json:"generation,omitempty" yaml:"generation,omitempty"`
}
```

Validate name, namespace, label keys, label values, display-name length, and
non-negative generation. Keep server-field rejection in submission validation.

- [x] **Step 3: Implement manifest**

```go
type Manifest struct {
	TypeMeta
	Metadata Metadata        `json:"metadata" yaml:"metadata"`
	Spec     json.RawMessage `json:"spec" yaml:"spec"`
	Status   json.RawMessage `json:"status,omitempty" yaml:"status,omitempty"`
}

func (manifest Manifest) ValidateSubmission() error
func (manifest Manifest) ValidateStored() error
```

Submission requires a non-empty JSON object Spec and rejects UID,
resourceVersion, generation, and Status. Stored validation permits them.
- [x] **Step 4: Verify**

```bash
go test ./backend/internal/domain/orchestrationresource \
  -run 'TestManifest|TestMetadata' -count=1
```

### Task 3: Draft And Resolved References

**Files:**
- Create: `backend/internal/domain/orchestrationresource/reference.go`
- Test: `backend/internal/domain/orchestrationresource/reference_test.go`

- [x] **Step 1: Write failing reference tests**

```go
func TestReferenceDraftUsesScopedName(t *testing.T) {
	ref := Reference{Kind: "ModelBinding", Namespace: mustSlug("acme"), Name: mustSlug("coding-primary")}
	require.NoError(t, ref.ValidateDraft("acme"))
}

func TestReferenceResolvedRequiresImmutableIdentity(t *testing.T) {
	ref := validResolvedReferenceForTest()
	ref.Digest = ""
	require.Error(t, ref.ValidateResolved("acme"))
}
```

- [x] **Step 2: Implement reference validation**

```go
type Reference struct {
	APIVersion string       `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string       `json:"kind" yaml:"kind"`
	Namespace  slugkit.Slug `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name       slugkit.Slug `json:"name" yaml:"name"`
	UID        string       `json:"uid,omitempty" yaml:"uid,omitempty"`
	Revision   int64        `json:"revision,omitempty" yaml:"revision,omitempty"`
	Digest     string       `json:"digest,omitempty" yaml:"digest,omitempty"`
}

func (ref Reference) ValidateDraft(defaultNamespace string) error
func (ref Reference) ValidateResolved(defaultNamespace string) error
```

Draft permits name lookup. Resolved requires UID, positive revision, and
lowercase `sha256:` digest. Both reject cross-namespace references.
### Task 4: Secret Reference And Conditions

**Files:**
- Create: `backend/internal/domain/orchestrationresource/secret_reference.go`
- Create: `backend/internal/domain/orchestrationresource/condition.go`
- Test: `backend/internal/domain/orchestrationresource/secret_reference_test.go`
- Test: `backend/internal/domain/orchestrationresource/condition_test.go`

- [x] **Step 1: Implement reference-only Secret shape**

```go
type SecretReference struct {
	Name     slugkit.Slug `json:"name" yaml:"name"`
	Key      slugkit.Slug `json:"key" yaml:"key"`
	Revision int64        `json:"revision,omitempty" yaml:"revision,omitempty"`
}
```

The type has no value field. Validate positive revision when supplied.
- [x] **Step 2: Implement status conditions**

```go
type Condition struct {
	Type               string    `json:"type" yaml:"type"`
	Status             string    `json:"status" yaml:"status"`
	Reason             string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message            string    `json:"message,omitempty" yaml:"message,omitempty"`
	ObservedGeneration int64     `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime" yaml:"lastTransitionTime"`
}
```

Status is exactly `True`, `False`, or `Unknown`.

- [x] **Step 3: Run terminal verification**

```bash
go test ./backend/internal/domain/orchestrationresource -count=1
go test ./backend/internal/domain/workerspec ./backend/internal/service/workercreation -count=1
```
