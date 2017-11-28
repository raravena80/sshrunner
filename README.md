# sshrunner [![CircleCI Build Status](https://circleci.com/gh/raravena80/sshrunner.svg?style=shield)](https://circleci.com/gh/raravena80/sshrunner) [![Coverage Status](https://coveralls.io/repos/github/raravena80/sshrunner/badge.svg?branch=master)](https://coveralls.io/github/raravena80/sshrunner?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/raravena80/sshrunner)](https://goreportcard.com/report/github.com/raravena80/sshrunner) [![Documentation](https://godoc.org/github.com/raravena80/sshrunner?status.svg)](http://godoc.org/github.com/raravena80/sshrunner) [![Apache Licensed](https://img.shields.io/badge/license-Apache2.0-blue.svg)](https://raw.githubusercontent.com/raravena80/sshrunner/master/LICENSE)
Run commands across servers using ssh

## Usage
```
Sshrunner runs an ssh command across multiple servers

For example:
$ sshrunner -c "mkdir /tmp/tmpdir" -m 17.2.2.2,17.2.3.2

Makes /tmp/tmpdir in 17.2.2.2 and 17.2.3.2 (It can also take dns names)

Usage:
  sshrunner [flags]

Flags:
  -s, --agentsocket string   Socket for the ssh agent (default "/private/tmp/com.apple.launchd.xxx/Listeners")
  -c, --command string       Command to run
      --config string        config file (default is $HOME/.sshrunner.yaml)
  -h, --help                 help for sshrunner
  -k, --key string           Ssh key to use for authentication, full path (default "/Users/raravena/.ssh/id_rsa")
  -m, --machines strings     Hosts to run command on
  -p, --port int             Ssh port to connect to (default 22)
  -a, --useagent             Use agent for authentication
  -u, --user string          User to run the command as (default "raravena")
```

## Config

Sample `~/.sshrunner.yaml`

```
sshrunner:
  user: ubuntu
  key: /Users/username/.ssh/id_rsa
  useagent: true
  machines:
    - 172.1.1.1
    - 172.1.1.2
    - 172.1.1.3
    - 172.1.1.4
    - 172.1.1.5
  command: sudo rm -f /var/log/syslog.*
```
