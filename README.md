[![GoDoc](https://godoc.org/github.com/cybozu-go/log?status.png)][godoc]
[![Build Status](https://travis-ci.org/cybozu-go/log.png)](https://travis-ci.org/cybozu-go/log)

Logging framework for Go
========================

This is a logging framework mainly for our Go products.

Be warned that this is a _framework_ rather than a library.
Most features cannot be configured.

Features
--------

* Light-weight.

    Hard-coded maximum log buffer size and 1-pass formatters
    help cybozu/log be memory- and CPU- efficient.

* Less configuration.

    In fact, nothing need to be configured for most cases.

* Built-in logfmt formatter.

    By default, logs are formatted in [logfmt][], and goes out
    to the standard error.

* Automatic redirect for Go standard logs.

    The framework automatically redirects [Go standard logs][golog]
    to itself.

Usage
-----

Read [the documentation][godoc].

License
-------

[MIT](https://opensource.org/licenses/MIT)

[logfmt]: https://brandur.org/logfmt
[fluentd]: http://www.fluentd.org/
[golog]: https://golang.org/pkg/log/
[godoc]: https://godoc.org/github.com/cybozu-go/log
