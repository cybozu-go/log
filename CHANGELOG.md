# Change Log

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.2.0] - 2016-08-26
### Added
- `Logger.WriteThrough` to write arbitrary bytes to the underlying writer.

## [1.1.2] - 2016-08-25
### Changed
- These interfaces are formatted nicer in logs.
    - [`encoding.TextMarshaler`](https://golang.org/pkg/encoding/#TextMarshaler)
    - [`json.Marshaler`](https://golang.org/pkg/encoding/json/#Marshaler)
    - [`error`](https://golang.org/pkg/builtin/#error)

## [1.1.1] - 2016-08-24
### Added
- `FnError` field name constant for error strings.
- [SPEC] add "exec" and "http" log types.

### Changed
- Invalid UTF-8 string no longer results in an error.

## [1.1.0] - 2016-08-22
### Changed
- `Logger.Writer`: fixed a minor bug.
- "id" log field is renamed to "request_id" (API is not changed).

## [1.0.1] - 2016-08-20
### Changed
- [Logger.Writer](https://godoc.org/github.com/cybozu-go/log#Logger.Writer) is rewritten for better performance.

[Unreleased]: https://github.com/cybozu-go/log/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/cybozu-go/log/compare/v1.1.2...v1.2.0
[1.1.2]: https://github.com/cybozu-go/log/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/cybozu-go/log/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/cybozu-go/log/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/cybozu-go/log/compare/v1.0.0...v1.0.1
