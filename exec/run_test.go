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
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/testdata"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"reflect"
	"testing"
)

var (
	testPrivateKeys map[string]interface{}
	testSigners     map[string]ssh.Signer
	testPublicKeys  map[string]ssh.PublicKey
	sshAgentSocket  string
)

func init() {
	var err error

	n := len(testdata.PEMBytes)
	testSigners = make(map[string]ssh.Signer, n)
	testPrivateKeys = make(map[string]interface{}, n)
	testPublicKeys = make(map[string]ssh.PublicKey, n)

	for t, k := range testdata.PEMBytes {
		testPrivateKeys[t], err = ssh.ParseRawPrivateKey(k)
		if err != nil {
			panic(fmt.Sprintf("Unable to parse test key %s: %v", t, err))
		}
		testSigners[t], err = ssh.NewSignerFromKey(testPrivateKeys[t])
		if err != nil {
			panic(fmt.Sprintf("Unable to create signer for test key %s: %v", t, err))
		}
		testPublicKeys[t] = testSigners[t].PublicKey()
	}

	randomStr := fmt.Sprintf("%v", rand.Intn(5000))
	socketFile := "/tmp/gosocket" + randomStr + ".sock"
	setupSshAgent(socketFile)
}

func TestMakeSigner(t *testing.T) {
	tests := []struct {
		name     string
		key      mockSSHKey
		expected ssh.Signer
	}{
		{name: "Basic key signer with valid rsa key",
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["rsa"],
			},
			expected: testSigners["rsa"],
		},
		{name: "Basic key signer with valid dsa key",
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["dsa"],
			},
			expected: testSigners["dsa"],
		},
		{name: "Basic key signer with valid ecdsa key",
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["ecdsa"],
			},
			expected: testSigners["ecdsa"],
		},
		{name: "Basic key signer with valid user key",
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["user"],
			},
			expected: testSigners["user"],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write content of the key to the keyname file
			ioutil.WriteFile(tt.key.keyname, tt.key.content, 0644)
			returned, _ := makeSigner(tt.key.keyname)
			if !reflect.DeepEqual(returned, tt.expected) {
				t.Errorf("Value received: %v expected %v", returned, tt.expected)
			}
			os.Remove(tt.key.keyname)
		})
	}
}

func setupSshAgent(socketFile string) {
	done := make(chan string, 1)
	a := agent.NewKeyring()
	go func(done chan<- string) {
		ln, err := net.Listen("unix", socketFile)
		if err != nil {
			panic(fmt.Sprintf("Couldn't create socket for tests %v", err))
		}
		// Need to wait until the socket is setup
		firstTime := true
		for {
			if firstTime == true {
				done <- socketFile
				firstTime = false
			}
			c, err := ln.Accept()
			defer c.Close()
			if err != nil {
				panic(fmt.Sprintf("Couldn't accept connection to agent tests %v", err))
			}
			go func(c io.ReadWriter) {
				err := agent.ServeAgent(a, c)
				if err != nil {
					panic(fmt.Sprintf("Couldn't serve ssh agent for tests %v", err))
				}

			}(c)
		}

	}(done)
	sshAgentSocket = <-done
}

func addKeytoSSHAgent(key agent.AddedKey) {
	aConn, _ := net.Dial("unix", sshAgentSocket)
	sshAgent := agent.NewClient(aConn)
	sshAgent.Add(key)
}

func removeKeyfromSSHAgent(key ssh.PublicKey) {
	aConn, _ := net.Dial("unix", sshAgentSocket)
	sshAgent := agent.NewClient(aConn)
	sshAgent.Remove(key)
}

func TestMakeKeyring(t *testing.T) {
	tests := []struct {
		name     string
		useagent bool
		key      mockSSHKey
		expected ssh.AuthMethod
	}{
		{name: "Basic key ring with valid rsa key",
			useagent: false,
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["rsa"],
			},
			expected: ssh.PublicKeys(testSigners["rsa"]),
		},
		{name: "Basic key ring with valid dsa key",
			useagent: false,
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["dsa"],
			},
			expected: ssh.PublicKeys(testSigners["dsa"]),
		},
		{name: "Basic key ring with valid ecdsa key",
			useagent: false,
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["ecdsa"],
			},
			expected: ssh.PublicKeys(testSigners["ecdsa"]),
		},
		{name: "Basic key ring with valid user key",
			useagent: false,
			key: mockSSHKey{
				keyname: "/tmp/mockkey",
				content: testdata.PEMBytes["user"],
			},
			expected: ssh.PublicKeys(testSigners["user"]),
		},
		{name: "Basic key ring agent with valid rsa key",
			useagent: true,
			key: mockSSHKey{
				keyname: "",
				content: testdata.PEMBytes["rsa"],
				privkey: agent.AddedKey{PrivateKey: testPrivateKeys["rsa"]},
				pubkey:  testPublicKeys["rsa"],
			},
			expected: ssh.PublicKeys(testSigners["rsa"]),
		},
		{name: "Basic key ring agent with valid dsa key",
			useagent: true,
			key: mockSSHKey{
				keyname: "",
				content: testdata.PEMBytes["dsa"],
				privkey: agent.AddedKey{PrivateKey: testPrivateKeys["dsa"]},
				pubkey:  testPublicKeys["dsa"],
			},
			expected: ssh.PublicKeys(testSigners["dsa"]),
		},
		{name: "Basic key ring agent with valid ecdsa key",
			useagent: true,
			key: mockSSHKey{
				keyname: "",
				content: testdata.PEMBytes["ecdsa"],
				privkey: agent.AddedKey{PrivateKey: testPrivateKeys["ecdsa"]},
				pubkey:  testPublicKeys["ecdsa"],
			},
			expected: ssh.PublicKeys(testSigners["ecdsa"]),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.useagent == true {
				addKeytoSSHAgent(tt.key.privkey)
			}
			// Write content of the key to the keyname file
			if tt.key.keyname != "" {
				ioutil.WriteFile(tt.key.keyname, tt.key.content, 0644)
			}
			returned := makeKeyring(tt.key.keyname, tt.useagent)
			// DeepEqual always returns false for functions unless nil
			// hence converting to string to compare
			check1 := reflect.ValueOf(returned).String()
			check2 := reflect.ValueOf(tt.expected).String()
			if !reflect.DeepEqual(check1, check2) {
				t.Errorf("Value received: %v expected %v", returned, tt.expected)
			}
			if tt.useagent == true {
				removeKeyfromSSHAgent(tt.key.pubkey)
			}
			if tt.key.keyname != "" {
				os.Remove(tt.key.keyname)
			}
		})
	}
}

func TestTearDown(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "Teardown SSH Agent",
			id: "sshAgentTdown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id == "sshAgentTdown" {
				os.Remove(sshAgentSocket)
			}

		})

	}
}
