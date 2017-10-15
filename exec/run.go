// Copyright © 2017 Ricardo Aravena <raravena80@gmail.com>
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
	"io/ioutil"
	"os"
	"time"
)

type SignerContainer struct {
	signers []ssh.Signer
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

func makeKeyring() ssh.AuthMethod {
	signers := []ssh.Signer{}
	keys := []string{
		os.Getenv("HOME") + "/.ssh/id_rsa",
		os.Getenv("HOME") + "/.ssh/id_dsa"}

	for _, keyname := range keys {
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}
	return ssh.PublicKeys(signers...)
}

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	fmt.Println(hostname)
	conn, err := ssh.Dial("tcp", hostname+":22", config)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return hostname + ": " + stdoutBuf.String()
}

// Run the ssh command
func Run(machines []string, cmd string, user string) {
	// in 5 seconds the message will come to timeout channel
	timeout := time.After(5 * time.Second)
	results := make(chan string, 10)

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{makeKeyring()},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for _, m := range machines {
		go func(hostname string) {
			results <- executeCmd(cmd, hostname, config)
			// we’ll write results into the buffered channel of strings
		}(m)
	}

	for i := 0; i < len(machines); i++ {
		select {
		case res := <-results:
			fmt.Print(res)
		case <-timeout:
			fmt.Println("Timed out!")
			return
		}
	}
}
