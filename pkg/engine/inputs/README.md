# Engine Inputs Package

This package provides typed, validated, and normalized Inputs structs for each `engine.StepAction`.

## Contract Invariants

### Producer Side (Planner/Adapter)

1. **Create typed struct** from operation metadata
2. **Call `Normalize()`** - sorts set-like fields, normalizes paths, trims strings
3. **Call `Validate()`** - enforces required fields, constraints, path rules
4. **Marshal to JSON** - deterministic output

**Order is critical**: `Normalize()` → `Validate()` → `Marshal()`

### Consumer Side (Agent/Executor)

1. **Unmarshal with `UnmarshalStrict()`** - rejects unknown fields via `DisallowUnknownFields()`
2. **Call `Validate()`** - re-validate after unmarshal (defensive)
3. **Use inputs** - all fields are normalized and validated

## Determinism Guarantees

- **Set-like lists** are sorted lexicographically (tags, services, targets, etc.)
- **KV pairs** are sorted by key (build_args, labels, variables, headers)
- **Paths** are normalized (forward slashes, no `..`, relative only)
- **Hashes** are validated (sha256 = 64 hex chars)
- **JSON output** is deterministic (no map iteration order issues)

## Example Usage

### Producer (Adapter)

```go
in := &inputs.BuildInputs{
    Provider:   "generic",
    Workdir:    "apps/backend",
    Dockerfile: "Dockerfile", // Producer must set explicitly
    Context:    ".",           // Producer must set explicitly
}
if err := in.Normalize(); err != nil {
    return err
}
if err := in.Validate(); err != nil {
    return err
}
jsonBytes, err := json.Marshal(in)
```

### Consumer (Agent)

```go
var in inputs.BuildInputs
if err := inputs.UnmarshalStrict(jsonBytes, &in); err != nil {
    return err
}
if err := in.Validate(); err != nil {
    return err
}
// Use in.* fields safely
```

## See Also

- `spec/engine/plan-actions.md` - Complete schema specification
- `pkg/engine/types.go` - Plan and StepAction definitions

