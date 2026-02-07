---
applyTo: "go.mod"
---

- **Patch version in go directive is intentional**: Since Go 1.21, the `go` directive supports patch versions (e.g., `go 1.24.7`). This project intentionally specifies a patch version to set the minimum required Go version. Do not suggest removing the patch version or replacing it with major.minor only.
