#tranor

[![Build Status](https://travis-ci.org/ef-ctx/tranor.svg?branch=master)](https://travis-ci.org/ef-ctx/tranor)
[![codecov](https://codecov.io/gh/ef-ctx/tranor/branch/master/graph/badge.svg)](https://codecov.io/gh/ef-ctx/tranor)
[![Go Report Card](https://goreportcard.com/badge/github.com/ef-ctx/tranor)](https://goreportcard.com/report/github.com/ef-ctx/tranor)

Plugin which adds the "environment" concept (eg. dev, QA, prod) to Tsuru PaaS.

##Using

In order to use tranor, one needs to install and configure it. Assuming that
[Go is installed and configured](https://golang.org/doc/install), the two
commands below should get tranor up and running:

```
% go get github.com/ef-ctx/tranor
% tranor target-set <remote-tranor-config-server>
```

The config server is an HTTP endpoint that contains the Tsuru target server and
the list of environment names and DNS suffixes, in the following format:


```json
{
	"target": "http://tsuru.example.com",
	"envs": [
		{
			"name": "dev",
			"dnsSuffix": "dev.example.com"
		},
		{
			"name": "stage",
			"dnsSuffix": "stage.example.com"
		},
		{
			"name": "prod",
			"dnsSuffix": "example.com"
		}
	]
}
```

For more details and some terminal session examples, check the
[usage.md](https://github.com/ef-ctx/tranor/blob/master/usage.md) page.

##Contributing and running tests

Contributions are welcome! In order to run tests locally, you need to be have
Go 1.7+ and run:

```
% go test
```

You can run integration tests against a real tsuru server. The test suite
assumes that the platform ``python`` is available, as well as the pools
``dev\dev.example.com``, ``qa\qa.example.com``, ``stage\stage.example.com`` and
``prod\example.com``, the plan ``medium`` and the team ``myteam``. Having all
requirements satisfied, one can run:

```
% TSURU_TEST_HOST=<tsuru-server> TSURU_TEST_TOKEN=<token-value> go test
```

If you have tsuru-admin available, you can run ``make prepare-test-server``:

```
% TSURU_TEST_HOST=<tsuru-server> TSURU_TEST_TOKEN=<token-value> make prepare-test-server
```

You can get the value of TSURU_TEST_HOST with the command ``tsuru target-list``
and the value of TSURU_TEST_TOKEN with the command ``tsuru token-show``.

You can also enable coverage report with the `-cover` flag:

```
% go test -cover
```

It's possible to get a nicer report in the browser by using the cover tool:

```
% go test -coverprofile cover.out
% go tool cover -html cover.out
```

We also have a script to check code formatting and lint the code:

```
% ./check-fmt.sh
```
