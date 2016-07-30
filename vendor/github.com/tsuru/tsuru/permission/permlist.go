// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package permission

//go:generate bash -c "rm -f permitems.go && go run ./generator/main.go -o permitems.go"

var PermissionRegistry = (&registry{}).addWithCtx(
	"app", []contextType{CtxApp, CtxTeam, CtxPool},
).addWithCtx(
	"app.create", []contextType{CtxTeam},
).add(
	"app.update.description",
	"app.update.log",
	"app.update.pool",
	"app.update.unit.add",
	"app.update.unit.remove",
	"app.update.unit.register",
	"app.update.unit.status",
	"app.update.env.set",
	"app.update.env.unset",
	"app.update.restart",
	"app.update.sleep",
	"app.update.start",
	"app.update.stop",
	"app.update.swap",
	"app.update.grant",
	"app.update.revoke",
	"app.update.teamowner",
	"app.update.cname.add",
	"app.update.cname.remove",
	"app.update.plan",
	"app.update.bind",
	"app.update.events",
	"app.update.unbind",
	"app.deploy",
	"app.deploy.archive-url",
	"app.deploy.build",
	"app.deploy.git",
	"app.deploy.image",
	"app.deploy.rollback",
	"app.deploy.upload",
	"app.read",
	"app.read.deploy",
	"app.read.env",
	"app.read.events",
	"app.read.metric",
	"app.read.log",
	"app.delete",
	"app.run",
	"app.admin.unlock",
	"app.admin.routes",
	"app.admin.quota",
).addWithCtx(
	"node", []contextType{CtxPool},
).add(
	"node.create",
	"node.read",
	"node.update",
	"node.delete",
	"node.autoscale",
).addWithCtx(
	"machine", []contextType{CtxIaaS},
).add(
	"machine.create",
	"machine.delete",
	"machine.read",
	"machine.update.events",
	"machine.read.events",
	"machine.template.create",
	"machine.template.delete",
	"machine.template.update",
	"machine.template.read",
).addWithCtx(
	"team", []contextType{CtxTeam},
).addWithCtx(
	"team.create", []contextType{},
).add(
	"team.read.events",
	"team.delete",
	"team.update.events",
).add(
	"user.create",
	"user.delete",
	"user.read.events",
	"user.update.token",
	"user.update.events",
	"user.update.quota",
	"user.update.password",
	"user.update.reset",
	"user.update.key.add",
	"user.update.key.remove",
).addWithCtx(
	"service", []contextType{CtxService, CtxTeam},
).addWithCtx(
	"service.create", []contextType{CtxTeam},
).add(
	"service.read.doc",
	"service.read.plans",
	"service.read.events",
	"service.update.proxy",
	"service.update.events",
	"service.update.revoke-access",
	"service.update.grant-access",
	"service.update.doc",
	"service.delete",
).addWithCtx(
	"service-instance", []contextType{CtxServiceInstance, CtxTeam},
).addWithCtx(
	"service-instance.create", []contextType{CtxTeam},
).add(
	"service-instance.read.events",
	"service-instance.read.status",
	"service-instance.delete",
	"service-instance.update.events",
	"service-instance.update.proxy",
	"service-instance.update.bind",
	"service-instance.update.unbind",
	"service-instance.update.grant",
	"service-instance.update.revoke",
	"service-instance.update.description",
).add(
	"role.create",
	"role.delete",
	"role.read.events",
	"role.update.events",
	"role.update.assign",
	"role.update.dissociate",
	"role.update.permission.add",
	"role.update.permission.remove",
	"role.default.create",
	"role.default.delete",
).add(
	"platform.create",
	"platform.delete",
	"platform.update",
).add(
	"plan.create",
	"plan.delete",
).addWithCtx(
	"pool", []contextType{CtxPool},
).addWithCtx(
	"pool.create", []contextType{},
).add(
	"pool.read.events",
	"pool.update.events",
	"pool.update.logs",
	"pool.delete",
).add(
	"debug",
).add(
	"healing.read",
).addWithCtx(
	"healing", []contextType{CtxPool},
).add(
	"healing.read",
	"healing.update",
).addWithCtx(
	"nodecontainer", []contextType{CtxPool},
).add(
	"nodecontainer.create",
	"nodecontainer.read",
	"nodecontainer.update",
	"nodecontainer.update.upgrade",
	"nodecontainer.delete",
)
