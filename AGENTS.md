# AGENTS.md

## Tests

Run the full suite:

```sh
go test ./...
```

### Snapshot tests

Power-assert diagram output is captured and compared against a golden txtar snapshot in `testdata/power_assert.txtar`. A normal `go test ./...` run compares against the snapshot and fails on any mismatch.

When a diagram intentionally changes (new assertion added, rendering updated, etc.), regenerate the snapshot:

```sh
ASSERT_UPDATE_SNAPSHOT=1 go test ./...
```
