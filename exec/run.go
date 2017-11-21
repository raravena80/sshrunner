// Copyright © 2017 Ricardo Aravena <raravena@branch.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io/ioutil"
	"net"
	"os"
	"time"
)

// Options Type that holds all the options for sshrunner
type Options struct {
	machines    []string
	port        int
	user        string
	cmd         string
	key         string
	agentSocket string
	useAgent    bool
}

type executeResult struct {
	result string
	err    error
}

// User Sets the user option
func User(u string) func(*Options) {
	return func(e *Options) {
		e.user = u
	}
}

// Port Sets the port option
func Port(p int) func(*Options) {
	return func(e *Options) {
		e.port = p
	}
}

// Cmd Sets the command to be run option
func Cmd(c string) func(*Options) {
	return func(e *Options) {
		e.cmd = c
	}
}

// Machines Sets the list of machines where to run the command
func Machines(m []string) func(*Options) {
	return func(e *Options) {
		e.machines = m
	}
}

// Key Sets the ssh key filename used to authenticate
func Key(k string) func(*Options) {
	return func(e *Options) {
		e.key = k
	}
}

// UseAgent Sets whether we want to use the ssh agent to authenticate
func UseAgent(u bool) func(*Options) {
	return func(e *Options) {
		e.useAgent = u
	}
}

// AgentSocket Sets the agent socket file
// Default is the value for env var SSH_AUTH_SOCK
func AgentSocket(s string) func(*Options) {
	return func(e *Options) {
		e.agentSocket = s
	}
}

func makeSigner(keyname string) (signer ssh.Signer, err error) {
	fp, err := os.Open(keyname)
	if err != nil {
		return
	}
	defer fp.Close()

	buf, _ := ioutil.ReadAll(fp)
	signer, _ = ssh.ParsePrivateKey(buf)
	return
}

func makeKeyring(key, socket string, useAgent bool) ssh.AuthMethod {
	signers := []ssh.Signer{}

	if useAgent == true {
		aConn, _ := net.Dial("unix", socket)
		sshAgent := agent.NewClient(aConn)
		aSigners, _ := sshAgent.Signers()
		for _, signer := range aSigners {
			signers = append(signers, signer)
		}
	}

	keys := []string{key}

	for _, keyname := range keys {
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}
	return ssh.PublicKeys(signers...)
}

func executeCmd(opt Options, hostname string, config *ssh.ClientConfig) executeResult {

	port := fmt.Sprint(opt.port)
	conn, err := ssh.Dial("tcp", hostname+":"+port, config)

	if err != nil {
		return executeResult{result: "Connection refused",
			err: err}
	}

	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	err = session.Run(opt.cmd)

	return executeResult{result: hostname + ":\n" + stdoutBuf.String(),
		err: err}
}

// Run Main function that kicks off the run
func Run(options ...func(*Options)) bool {
	opt := Options{}
	for _, option := range options {
		option(&opt)
	}

	// in 10 seconds the message will come to timeout channel
	timeout := time.After(10 * time.Second)
	results := make(chan executeResult, len(opt.machines))

	config := &ssh.ClientConfig{
		User: opt.user,
		Auth: []ssh.AuthMethod{
			makeKeyring(
				opt.key,
				opt.agentSocket,
				opt.useAgent),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for _, m := range opt.machines {
		go func(hostname string) {
			// we’ll write results into the buffered channel of strings
			results <- executeCmd(opt, hostname, config)
		}(m)
	}

	retval := true

	fmt.Println(len(opt.machines))
	for i := 0; i < len(opt.machines); i++ {
		select {
		case res := <-results:
			if res.err == nil {
				fmt.Print(res.result)
			} else {
				fmt.Println(res.result, "\n", res.err)
				retval = false
			}
		case <-timeout:
			fmt.Println(fmt.Sprintf("%v:", opt.machines[i]))
			fmt.Println("Server timed out!")
			retval = false
		}
	}
	return retval
}
