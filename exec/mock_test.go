package exec

import "golang.org/x/crypto/ssh"
import "golang.org/x/crypto/ssh/agent"

type mockSSHKey struct {
	keyname string
	content []byte
	privkey agent.AddedKey
	pubkey  ssh.PublicKey
}
