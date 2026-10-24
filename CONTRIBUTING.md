# Contributing

Thanks for helping improve Goralens.

1. Open an issue first for larger changes so the approach can be agreed on.
2. Keep the code idiomatic Go and the comments short; run `gofmt`.
3. Add or update tests. Run every module's tests before opening a pull request:

   ```bash
   go vet ./... && go test -race ./...
   (cd examples/basic && go test -race ./...)
   (cd redishook && go test ./...)
   ```

4. The dashboard (`ui/`) is plain HTML, CSS and JavaScript with no build step and no external
   resources; keep it that way. Insert untrusted values as text, never as HTML.
5. Describe user-visible changes in `CHANGELOG.md` under "Unreleased".

Please report security issues privately to the maintainer instead of opening a public issue.
