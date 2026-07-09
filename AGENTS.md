## Required actions after changes

1. After changing code:
- Run: `make fix`

## Testing private symbols

Tests that need unexported API: add `export_test.go` (`package foo`), not `package foo` in `*_test.go` and not `//nolint:testpackage`.

Example:

```go
// export_test.go
package foo

var DoThing = doThing
```

```go
// thing_test.go
package foo_test
```

## exhaustruct

If `{pkg}.{Type} is missing fields {...}`: for lazy-init or intentionally omitted fields add `exhaustruct:"optional"` on the struct field — not `nolint`.
