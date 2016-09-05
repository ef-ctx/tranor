// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

const listOfApps = `[
  {
    "name": "tsuru-dashboard",
    "cname": [],
    "ip": "tsuru-dashboard.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj2-dev",
    "cname": [
      "proj2.dev.example.com"
    ],
    "ip": "proj2-dev.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj2-qa",
    "cname": [
      "proj2.qa.example.com"
    ],
    "ip": "proj2-qa.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj2-stage",
    "cname": [
      "proj2.stage.example.com"
    ],
    "ip": "proj2-stage.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj2-production",
    "cname": [
      "proj2.example.com"
    ],
    "ip": "proj2-production.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "myblog-qa",
    "cname": [
      "myblog.qa.example.com",
      "myblog.qa2.example.com"
    ],
    "ip": "myblog-dev.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj1-dev",
    "cname": [
      "proj1.dev.example.com"
    ],
    "ip": "proj1-dev.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj1-qa",
    "cname": [
      "proj1.qa.example.com"
    ],
    "ip": "proj1-qa.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj1-stage",
    "cname": [
      "proj1.stage.example.com"
    ],
    "ip": "proj1-stage.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "myblog-dev",
    "cname": [],
    "ip": "myblog-dev.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj1-production",
    "cname": [
      "proj1.example.com"
    ],
    "ip": "proj1-production.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj3-dev",
    "cname": [
      "proj3.dev.example.com"
    ],
    "ip": "proj3-dev.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  },
  {
    "name": "proj3-production",
    "cname": [
      "proj3.example.com"
    ],
    "ip": "proj3-production.192.168.50.4.nip.io",
    "lock": {
      "Locked": false,
      "Reason": "",
      "Owner": "",
      "AcquireDate": "0001-01-01T00:00:00Z"
    }
  }
]`

const deployments = `[
  {
    "ID": "57ccc9490640fd3def98b157",
    "App": "proj1-dev",
    "Timestamp": "2016-09-05T01:24:25.706Z",
    "Duration": 125995000000,
    "Commit": "40244ff2866eba7e2da6eee8a6fc51464c9f604f",
    "Error": "",
    "Image": "v938",
    "Log": "",
    "User": "admin@example.com",
    "Origin": "",
    "CanRollback": true,
    "RemoveDate": "0001-01-01T00:00:00Z",
    "Diff": ""
  }
]`
