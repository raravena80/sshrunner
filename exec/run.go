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

type executeResult struct {
	result string
	err    error
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

func makeKeyring(key string, useAgent bool) ssh.AuthMethod {
	signers := []ssh.Signer{}

	if useAgent == true {
		aConn, _ := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
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

func executeCmd(cmd, hostname, port string, config *ssh.ClientConfig) executeResult {
	conn, err := ssh.Dial("tcp", hostname+":"+port, config)

	if err != nil {
		return executeResult{result: "",
			err: err}
	}

	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	err = session.Run(cmd)

	return executeResult{result: hostname + ":\n" + stdoutBuf.String(),
		err: err}
}

// Run the ssh command
func Run(machines []string, port, cmd, user, key string, useAgent bool) bool {
	// in 5 seconds the message will come to timeout channel
	timeout := time.After(5 * time.Second)
	results := make(chan executeResult, len(machines))

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{makeKeyring(key, useAgent)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for _, m := range machines {
		go func(hostname string) {
			results <- executeCmd(cmd, hostname, port, config)
			// we’ll write results into the buffered channel of strings
		}(m)
	}

	for i := 0; i < len(machines); i++ {
		select {
		case res := <-results:
			if res.err == nil {
				fmt.Print(res.result)
			} else {
				fmt.Println(res.err)
				return false
			}
		case <-timeout:
			fmt.Println("Timed out!")
			return false
		}
	}
	return true
}
