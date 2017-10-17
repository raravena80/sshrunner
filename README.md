# sshrunner
Run commands across servers using ssh

## Usage
```
Sshrunner runs ssh commands across multiple servers

Usage:
  sshrunner [flags]

Flags:
  -c, --command string         Command to run
      --config string          config file (default is $HOME/.sshrunner.yaml)
  -h, --help                   help for sshrunner
  -k, --key string             Ssh key to use, full path (default "$HOME/.ssh/id_rsa")
  -m, --machines stringArray   Hosts to run command on
  -t, --toggle                 Help message for toggle
  -a, --useagent               Use agent for authentication
  -u, --user string            User to run the command as (default "username")
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
