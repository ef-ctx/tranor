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
    "name": "proj2-prod",
    "cname": [
      "proj2.example.com"
    ],
    "ip": "proj2-prod.192.168.50.4.nip.io",
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
    "name": "proj1-prod",
    "cname": [
      "proj1.example.com"
    ],
    "ip": "proj1-prod.192.168.50.4.nip.io",
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
    "name": "proj3-prod",
    "cname": [
      "proj3.example.com"
    ],
    "ip": "proj3-prod.192.168.50.4.nip.io",
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

const appInfo1 = `{
  "cname": ["proj1.dev.example.com"],
  "deploys": 35,
  "description": "my nice project",
  "ip": "proj1-dev.tsuru.example.com",
  "lock": {
    "Locked": false,
    "Reason": "",
    "Owner": "",
    "AcquireDate": "0001-01-01T00:00:00Z"
  },
  "name": "proj1-dev",
  "owner": "webmaster@example.com",
  "plan": {
    "name": "medium",
    "memory": 536870912,
    "swap": 1073741824,
    "cpushare": 1024,
    "default": true,
    "router": "hipache"
  },
  "platform": "python",
  "pool": "mypool",
  "repository": "git@example.com:proj1-dev.git",
  "teamowner": "admin",
  "teams": [
    "admin",
    "sysop"
  ],
  "units": [
    {
      "ID": "d7048b11e43dc1699b745dbe5a3a752c776dab37eb648c3c77a7aa2eb17d382f",
      "Name": "",
      "AppName": "proj1-dev",
      "ProcessName": "web"
    }
  ]
}
`

const appInfo2 = `{
  "cname": ["proj1.qa.example.com"],
  "deploys": 35,
  "description": "my nice project",
  "ip": "proj1-qa.tsuru.example.com",
  "lock": {
    "Locked": false,
    "Reason": "",
    "Owner": "",
    "AcquireDate": "0001-01-01T00:00:00Z"
  },
  "name": "proj1-qa",
  "owner": "webmaster@example.com",
  "plan": {
    "name": "medium",
    "memory": 536870912,
    "swap": 1073741824,
    "cpushare": 1024,
    "default": true,
    "router": "hipache"
  },
  "platform": "python",
  "pool": "mypool",
  "repository": "git@example.com:proj1-qa.git",
  "teamowner": "admin",
  "teams": [
    "admin",
    "sysop"
  ],
  "units": [
    {
      "ID": "d7065b11e43dc1699b745dbe5a3a752c776dab37eb648c3c77a7aa20eb17d382f",
      "Name": "",
      "AppName": "proj1-qa",
      "ProcessName": "web"
    },
    {
      "ID": "d7065b11e43dc1699b745dbe5a3a752c776dab37eb648c3c77a7aa20eb17d382f",
      "Name": "",
      "AppName": "proj1-qa",
      "ProcessName": "web"
    }
  ]
}
`

const appInfo3 = `{
  "cname": ["proj1.stage.example.com"],
  "deploys": 35,
  "description": "my nice project",
  "ip": "proj1-stage.tsuru.example.com",
  "lock": {
    "Locked": false,
    "Reason": "",
    "Owner": "",
    "AcquireDate": "0001-01-01T00:00:00Z"
  },
  "name": "proj1-stage",
  "owner": "webmaster@example.com",
  "plan": {
    "name": "medium",
    "memory": 536870912,
    "swap": 1073741824,
    "cpushare": 1024,
    "default": true,
    "router": "hipache"
  },
  "platform": "python",
  "pool": "mypool",
  "repository": "git@example.com:proj1-stage.git",
  "teamowner": "admin",
  "teams": [
    "admin",
    "sysop"
  ],
  "units": [
    {
      "ID": "71281212acdf482318129",
      "Name": "",
      "AppName": "proj1-stage",
      "ProcessName": "web"
    },
    {
      "ID": "d7085b38e43dc1699b745dbe5a3a725c776dab37eb648c3c77a7aa20eb49d382f",
      "Name": "",
      "AppName": "proj1-stage",
      "ProcessName": "web"
    }
  ]
}
`

const appInfo4 = `{
  "cname": ["proj1.example.com"],
  "deploys": 35,
  "description": "my nice project",
  "ip": "proj1-prod.tsuru.example.com",
  "lock": {
    "Locked": false,
    "Reason": "",
    "Owner": "",
    "AcquireDate": "0001-01-01T00:00:00Z"
  },
  "name": "proj1-prod",
  "owner": "webmaster@example.com",
  "plan": {
    "name": "medium",
    "memory": 536870912,
    "swap": 1073741824,
    "cpushare": 1024,
    "default": true,
    "router": "hipache"
  },
  "platform": "python",
  "pool": "mypool",
  "repository": "git@example.com:proj1-prod.git",
  "teamowner": "admin",
  "teams": [
    "admin",
    "sysop"
  ],
  "units": [
    {
      "ID": "71281212acdf482318094afafafafafafabdbedf883197",
      "Name": "",
      "AppName": "proj1-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71281184acdf482318193afafafafafafabdbedf183053",
      "Name": "",
      "AppName": "proj1-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71282437acdf482318155afafafafafafabdbedf182839",
      "Name": "",
      "AppName": "proj1-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71281271acdf482318180afafa0afafafabdbedf182883",
      "Name": "",
      "AppName": "proj1-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71281341acdf482318123afafafafafafabdbedf182932",
      "Name": "",
      "AppName": "proj1-prod",
      "ProcessName": "web"
    }
  ]
}
`

const appInfo5 = `{
  "cname": ["proj3.dev.example.com"],
  "deploys": 35,
  "description": "my nice project",
  "ip": "proj3-dev.tsuru.example.com",
  "lock": {
    "Locked": false,
    "Reason": "",
    "Owner": "",
    "AcquireDate": "0001-01-01T00:00:00Z"
  },
  "name": "proj3-dev",
  "owner": "webmaster@example.com",
  "plan": {
    "name": "autogenerated",
    "memory": 536870912,
    "swap": 1073741824,
    "cpushare": 1024,
    "default": true,
    "router": "hipache"
  },
  "platform": "python",
  "pool": "dev\\dev.example.com",
  "repository": "git@example.com:proj3-dev.git",
  "teamowner": "admin",
  "teams": [
    "admin",
    "sysop"
  ],
  "units": [
    {
      "ID": "d7048b11e43dc1699b745dbe5a3a752c776dab37eb648c3c77a7aa2eb17d382f",
      "Name": "",
      "AppName": "proj3-dev",
      "ProcessName": "web"
    }
  ]
}
`

const appInfo6 = `{
  "cname": ["proj3.example.com"],
  "deploys": 35,
  "description": "my nice project",
  "ip": "proj3-prod.tsuru.example.com",
  "lock": {
    "Locked": false,
    "Reason": "",
    "Owner": "",
    "AcquireDate": "0001-01-01T00:00:00Z"
  },
  "name": "proj3-prod",
  "owner": "webmaster@example.com",
  "plan": {
    "name": "autogenerated",
    "memory": 536870912,
    "swap": 1073741824,
    "cpushare": 1024,
    "default": true,
    "router": "hipache"
  },
  "platform": "python",
  "pool": "prod\\prod.example.com",
  "repository": "git@example.com:proj3-prod.git",
  "teamowner": "admin",
  "teams": [
    "admin",
    "sysop"
  ],
  "units": [
    {
      "ID": "71281212acdf482318094afafafafafafabdbedf883197",
      "Name": "",
      "AppName": "proj3-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71281184acdf482318193afafafafafafabdbedf183053",
      "Name": "",
      "AppName": "proj3-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71282437acdf482318155afafafafafafabdbedf182839",
      "Name": "",
      "AppName": "proj3-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71281271acdf482318180afafa0afafafabdbedf182883",
      "Name": "",
      "AppName": "proj3-prod",
      "ProcessName": "web"
    },
    {
      "ID": "71281341acdf482318123afafafafafafabdbedf182932",
      "Name": "",
      "AppName": "proj3-prod",
      "ProcessName": "web"
    }
  ]
}
`
