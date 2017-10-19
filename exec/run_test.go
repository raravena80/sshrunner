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
	"golang.org/x/crypto/ssh/testdata"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var (
	testPrivateKeys map[string]interface{}
	testSigners     map[string]ssh.Signer
	testPublicKeys  map[string]ssh.PublicKey
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
	}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write content of the key to the keyname file
			ioutil.WriteFile(tt.key.keyname, tt.key.content, 0644)
			returned := makeKeyring(tt.key.keyname, tt.useagent)
			// DeepEqual always returns false for functions unless nil
			// hence converting to string to compare
			check1 := reflect.ValueOf(returned).String()
			check2 := reflect.ValueOf(tt.expected).String()
			if !reflect.DeepEqual(check1, check2) {
				t.Errorf("Value received: %v expected %v", returned, tt.expected)
			}
			os.Remove(tt.key.keyname)
		})
	}
}
