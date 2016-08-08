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
			"name": "production",
			"dnsSuffix": "example.com"
		}
	]
}
```

##Contributing and running tests

Contributions are welcome! In order to run tests locally, you need to be have
Go 1.6+ and run:

```
% go test
```

You can also get a coverage report with the `-cover` flag:

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
