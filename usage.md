# tranor usage

## Listing all available commands

In order to list all available commands, one can just invoke ``tranor`` with no
parameters:

```
% tranor
tranor version 0.1.

Usage: tranor command [args]

Available commands:
  env-list             List currently available environments
  envvar-get           Gets environment variables of the project in the given environments
  envvar-set           Defines environment variables for a given project
  envvar-unset         Unset environment variables of the project in the given environments
  help
  login                Initiates a new tsuru session for a user
  logout               Logout will terminate the session with the tsuru server
  plan-list            List available plans that can be used when creating an app
  platform-list        Lists the available platforms
  project-create       Creates a remote project in the tranor server
  project-env-info     Displays information about a project in a specific environment
  project-info         Retrieves and displays information about the given project
  project-list         List the projects on tranor that you has access to
  project-remove       Removes the given project
  project-update       Updates the given project
  target-set           Sets the remote tranor server
  team-create          Create a team for the user
  team-list            List all teams that you are member
  team-remove          Removes a team from tsuru server
  user-info            Displays information about the current user
  version              Display the current version

Use tranor help <commandname> to get more information about a command.

Available topics:
  target

Use tranor help <topicname> to get more information about a topic.
```

## target-set

The command ``tranor target-set`` will define your remote tranor server:

```
% tranor target-set https://gist.githubusercontent.com/fsouza/ac50e24d18bd7c24588164481fa3686b/raw/08d693a2da421dcf5a0ac0e02ba45c050d44fc57
Target successfully defined!
```

## login and user-info

Before using tranor, one needs to login using ``tranor login``:

```
% tranor login
Email: admin@example.com
Password:
Successfully logged in!
```

The command ``tranor user-info`` displays information about the current
session:

```
% tranor user-info
Email: admin@example.com
Roles:
	AllowAll(global)
Permissions:
	*(global)
```

## team-create, team-remove and team-list

The commands ``tranor team-create``, ``tranor team-list`` and ``tranor
team-remove`` can be used to manage teams. Teams are required for creating
projects.

```
% tranor team-create sample
Team "sample" successfully created!
% tranor team-list
+--------+------------------+
| Team   | Permissions      |
+--------+------------------+
| admin  | app              |
|        | team             |
|        | service          |
|        | service-instance |
+--------+------------------+
| sample | app              |
|        | team             |
|        | service          |
|        | service-instance |
+--------+------------------+
% tranor team-remove sample
Are you sure you want to remove team "sample"? (y/n) y
Team "sample" successfully removed!
```

## env-list

The command ``tranor env-list`` lists all environments currently available:

```
% tranor env-list
+-------------+-------------------+
| Environment | DNS Suffix        |
+-------------+-------------------+
| dev         | dev.example.com   |
| qa          | qa.example.com    |
| stage       | stage.example.com |
| prod        | example.com       |
+-------------+-------------------+
```

## platform-list

The command ``tranor platform-list`` lists the available platforms:

```
% tranor platform-list
- python
```

## project-create

The command ``tranor project-create`` creates a new project:

```
% tranor project-create -h
tranor version 0.1.

Usage: tranor

creates a remote project in the tranor server

Flags:

  -d, --description (= "")
      description of the project
  -e, --envs  (= )
      comma-separated list of environments to use
  -h, --help  (= false)
      Display help and exit
  -l, --platform (= "")
      platform of the project
  -n, --name (= "")
      name of the project
  -p, --plan (= "")
      plan to use for the project
  -t, --team (= "")
      team that owns the project
```

The following flags are required: ``-n/--name``, ``-l/--platform`` and
``-t/--team``. Users can also specify a list of environments to use specific
environments (instead of all available ones):

```
% tranor project-create --name myproj --platform python --team admin --envs dev,stage,prod
successfully created the project "myproj"!
```

## project-info

The command ``tranor project-info`` displays information about a project:

```
% tranor project-info --name myproj
Project name: myproj
Description:
Repository:
Platform: python
Teams: admin
Owner: admin@example.com
Team owner: admin

+-------------+--------------------------+-------+--------------+-------------+-------+
| Environment | Address                  | Image | Git hash/tag | Deploy date | Units |
+-------------+--------------------------+-------+--------------+-------------+-------+
| dev         | myproj.dev.example.com   |       |              |             | 0     |
| stage       | myproj.stage.example.com |       |              |             | 0     |
| prod        | myproj.example.com       |       |              |             | 0     |
+-------------+--------------------------+-------+--------------+-------------+-------+
```

## project-env-info

The command ``tranor project-env-info`` gets more details about a project in a
specific environment:


```
% tranor project-env-info --name myproj --env dev
Application: myproj-dev
Description:
Repository:
Platform: python
Teams: admin
Address: myproj.dev.example.com, myproj-dev.192.168.99.100.nip.io
Owner: admin@example.com
Team owner: admin
Deploys: 0
Pool: dev\dev.example.com
Quota: 0/4 units

App Plan:
+---------------+--------+------+-----------+--------+---------+
| Name          | Memory | Swap | Cpu Share | Router | Default |
+---------------+--------+------+-----------+--------+---------+
| autogenerated | 0      | 0    | 100       |        | false   |
+---------------+--------+------+-----------+--------+---------+
```

## project-update

The command ``tranor project-update`` allows users to update informations about
the project:

```
% tranor project-update -h
tranor version 0.1.

Usage: tranor

Updates the given project

Flags:

  --add-envs  (= )
      comma-separated list of environments to add to the project
  -d, --description (= "")
      description of the project
  -h, --help  (= false)
      Display help and exit
  -n, --name (= "")
      name of the project
  -p, --plan (= "")
      plan to use for the project
  --remove-envs  (= )
      comma-separated list of environments to remove from the project
  -t, --team (= "")
      team that owns the project
```

For example, to change the description and the environments in use by the
project:

```
% tranor project-update --name myproj --description "sample project" --add-envs qa --remove-envs stage
adding new environments...
removing old environments...
Deleting from env "stage"... ok
% tranor project-info --name myproj
Project name: myproj
Description: sample project
Repository:
Platform: python
Teams: admin
Owner: admin@example.com
Team owner: admin

+-------------+------------------------+-------+--------------+-------------+-------+
| Environment | Address                | Image | Git hash/tag | Deploy date | Units |
+-------------+------------------------+-------+--------------+-------------+-------+
| dev         | myproj.dev.example.com |       |              |             | 0     |
| qa          | myproj.qa.example.com  |       |              |             | 0     |
| prod        | myproj.example.com     |       |              |             | 0     |
+-------------+------------------------+-------+--------------+-------------+-------+
```

## project-list

The command ``tranor project-list`` displays a list of user projects:

```
% tranor project-list
+---------+--------------+------------------------+
| Project | Environments | Address                |
+---------+--------------+------------------------+
| myproj  | dev          | myproj.dev.example.com |
|         | qa           | myproj.qa.example.com  |
|         | prod         | myproj.example.com     |
+---------+--------------+------------------------+
```

## project-remove

The command ``tranor project-remove`` removes a projects:

```
% tranor project-remove -h
tranor version 0.1.

Usage: tranor

removes the given project

Flags:

  -h, --help  (= false)
      Display help and exit
  -n, --name (= "")
      name of the project to remove
  -y, --assume-yes  (= false)
      Don't ask for confirmation.<Paste>
```

It will ask for confirmation before removing the project. The ``-y`` flag can
be used to skip confirmation.


```
% tranor project-remove --name myproj
Are you sure you want to remove the project "myproj"? (y/n) y
Deleting from env "dev"... ok
Deleting from env "qa"... ok
Deleting from env "stage"... ok
Deleting from env "prod"... ok
```

## envvar-get

The command ``tranor envvar-get`` gets the environment variables defined in the
specified environments:

```
% tranor envvar-get -h
tranor version 0.1.

Usage: tranor

gets environment variables of the project in the given environments

Flags:

  -e, --envs  (= )
      comma-separated list of environments to get the variables
  -h, --help  (= false)
      Display help and exit
  -n, --project-name (= "")
      name of the project
```

To get the variables defined in ``dev`` and ``prod`` for ``myproj``, one would
run:


```
% tranor envvar-get --project-name myproj -e dev,prod
variables in "dev":

 TSURU_APPNAME=*** (private variable)
 TSURU_APPDIR=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "prod":

 TSURU_APP_TOKEN=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APPDIR=*** (private variable)
```

## envvar-set

The command ``tranor envvar-set`` exports environment variables in the given
project in the provided environments:

```
% tranor envvar-set -h
tranor version 0.1.

Usage: tranor envvar-set <NAME=value> [NAME=value]... <-n/--project-name projectname> [-p/--private] [--no-restart]

defines environment variables for a given project

Flags:

  -e, --envs  (= )
      comma-separated list of environments to set the variables
  -h, --help  (= false)
      Display help and exit
  -n, --project-name (= "")
      name of the project
  --no-restart  (= false)
      set the environment variables without restarting the application process
  -p, --private  (= false)
      set the variables to private (not visible through command line)

Minimum # of arguments: 1
```

Variables should be specified in the format ``NAME=value``:

```
% tranor envvar-set --project-name myproj --envs dev,qa,prod DATABASE_USER=root DATABASE_PASSWORD=r00t DATABASE_NAME=mydb
setting variables in environment "dev"... ok
setting variables in environment "qa"... ok
setting variables in environment "prod"... ok
% tranor envvar-get --project-name myproj -e dev,prod
variables in "dev":

 TSURU_APPNAME=*** (private variable)
 TSURU_APPDIR=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)
 DATABASE_USER=root
 DATABASE_PASSWORD=r00t
 DATABASE_NAME=mydb


variables in "prod":

 DATABASE_PASSWORD=r00t
 DATABASE_NAME=mydb
 TSURU_APPNAME=*** (private variable)
 TSURU_APPDIR=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)
 DATABASE_USER=root
```

## envvar-unset

The command ``tranor envvar-unset`` removes environment variables from the
given environments:

```
% tranor envvar-unset -h
tranor version 0.1.

Usage: tranor

unset environment variables of the project in the given environments

Flags:

  -e, --envs  (= )
      comma-separated list of environments to set the variables
  -h, --help  (= false)
      Display help and exit
  -n, --project-name (= "")
      name of the project
  --no-restart  (= false)
      unset environment variables without restarting the application process

Minimum # of arguments: 1
```

To remove the DATABASE_PASSWORD environment variable:

```
% tranor envvar-unset --project-name myproj --envs dev,qa,prod DATABASE_PASSWORD
unsetting variables from environment "dev"... ok
unsetting variables from environment "qa"... ok
unsetting variables from environment "prod"... ok
% tranor envvar-get --project-name myproj -e dev,prod
variables in "dev":

 TSURU_APP_TOKEN=*** (private variable)
 DATABASE_USER=root
 DATABASE_NAME=mydb
 TSURU_APPNAME=*** (private variable)
 TSURU_APPDIR=*** (private variable)


variables in "prod":

 TSURU_APP_TOKEN=*** (private variable)
 DATABASE_USER=root
 DATABASE_NAME=mydb
 TSURU_APPNAME=*** (private variable)
 TSURU_APPDIR=*** (private variable)
```
