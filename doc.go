/*
Package log provides the standard logging framework for cybozu products.

As this is a framework rather than a library, most features are hard-coded
and non-customizable.

cybozu/log is a structured logger, that is, every log entry consists of
mandatory and optional fields.  Mandatory fields are:

    "tag" is by default the executables file name w/o directory path.
    "logged_at" is generated automatically by the framework.
    "severity" corresponds to each logging method such as "log.Error".
    "utsname" is generated automatically by the framework.
    "message" is provided by the argument for logging methods.

To help development, logs go to standard error by default.  This can be
changed to any io.Writer.  The log is formatted in so-called logfmt
described in: https://gist.github.com/kr/0e8d5ee4b954ce604bb2

If the framework detects that fluentd is running on the same host,
it sends logs to fluentd as well.  You may specify the UNIX domain
socket path by "FLUENTD_SOCK" environment variable.

The standard field names are defined as constants in this package.
For example, "secret" is defined as FnSecret.

Field data can be one of the following types:

    nil, bool, int, int64, string (UTF-8 string), []byte (binary data),
    []int, []int64, []string, time.Time

The framework automatically redirects Go's standard log output to
the default logger provided by this framework.
*/
package log
