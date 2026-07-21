package knowledgebase

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"

	kbdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
	"golang.org/x/crypto/ssh"
)

const (
	readOnlyDeployKeyConfig  = "_agentcloud_deploy_key_ro"
	readWriteDeployKeyConfig = "_agentcloud_deploy_key_rw"
)

type mountDeployKeys struct {
	readOnlyPrivate  string
	readOnlyPublic   string
	readWritePrivate string
	readWritePublic  string
}

func newMountDeployKeys() (*mountDeployKeys, error) {
	roPrivate, roPublic, err := newDeployKeyPair()
	if err != nil {
		return nil, err
	}
	rwPrivate, rwPublic, err := newDeployKeyPair()
	if err != nil {
		return nil, err
	}
	return &mountDeployKeys{
		readOnlyPrivate:  roPrivate,
		readOnlyPublic:   roPublic,
		readWritePrivate: rwPrivate,
		readWritePublic:  rwPublic,
	}, nil
}

func newDeployKeyPair() (string, string, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("knowledgebase: generate deploy key: %w", err)
	}
	privateBlock, err := ssh.MarshalPrivateKey(private, "")
	if err != nil {
		return "", "", fmt.Errorf("knowledgebase: marshal deploy private key: %w", err)
	}
	sshPublic, err := ssh.NewPublicKey(public)
	if err != nil {
		return "", "", fmt.Errorf("knowledgebase: marshal deploy public key: %w", err)
	}
	return string(pem.EncodeToMemory(privateBlock)), string(ssh.MarshalAuthorizedKey(sshPublic)), nil
}

func addMountDeployKeys(raw json.RawMessage, keys *mountDeployKeys) (json.RawMessage, error) {
	var config map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, fmt.Errorf("%w: source_config must be a JSON object: %v", ErrInvalidInput, err)
		}
	}
	if config == nil {
		config = map[string]any{}
	}
	config[readOnlyDeployKeyConfig] = keys.readOnlyPrivate
	config[readWriteDeployKeyConfig] = keys.readWritePrivate
	return json.Marshal(config)
}

func (s *Service) mountPrivateKey(raw json.RawMessage, mode string) (string, error) {
	decrypted, err := s.decryptSourceSecrets(raw)
	if err != nil {
		return "", err
	}
	var config map[string]any
	if err := json.Unmarshal(decrypted, &config); err != nil {
		return "", fmt.Errorf("%w: stored source_config is invalid", ErrInvalidInput)
	}
	keyName := readOnlyDeployKeyConfig
	if mode == kbdomain.MountModeReadWrite {
		keyName = readWriteDeployKeyConfig
	}
	privateKey, _ := config[keyName].(string)
	if privateKey == "" {
		return "", fmt.Errorf(
			"%w: knowledge base is missing its %s repository credential",
			ErrNotConfigured,
			mode,
		)
	}
	return privateKey, nil
}
