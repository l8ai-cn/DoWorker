package runner

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type commandReceiptStore struct {
	root string
}

func newCommandReceiptStore(root string) *commandReceiptStore {
	return &commandReceiptStore{root: root}
}

func commandReceiptStoreForRunner(runner MessageHandlerContext) *commandReceiptStore {
	if runner == nil {
		return nil
	}
	cfg := runner.GetConfig()
	if cfg == nil || (cfg.Workspace == "" && cfg.WorkspaceRoot == "") {
		return nil
	}
	return newCommandReceiptStore(
		filepath.Join(cfg.GetSandboxesDir(), ".command-receipts"),
	)
}

func (s *commandReceiptStore) ClaimPrompt(podKey, commandID string) (bool, error) {
	path, err := s.receiptPath(podKey, "prompt", commandID)
	if err != nil {
		return false, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := file.WriteString(commandID); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return false, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return false, err
	}
	return true, nil
}

func (s *commandReceiptStore) ReleasePrompt(podKey, commandID string) error {
	path, err := s.receiptPath(podKey, "prompt", commandID)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *commandReceiptStore) LoadVerification(
	podKey, requestID string,
) (*runnerv1.VerificationResultEvent, bool, error) {
	path, err := s.receiptPath(podKey, "verification", requestID)
	if err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	result := &runnerv1.VerificationResultEvent{}
	if err := protojson.Unmarshal(data, result); err != nil {
		return nil, false, err
	}
	return result, true, nil
}

func (s *commandReceiptStore) StoreVerification(
	podKey, requestID string,
	result *runnerv1.VerificationResultEvent,
) error {
	path, err := s.receiptPath(podKey, "verification", requestID)
	if err != nil {
		return err
	}
	data, err := protojson.Marshal(result)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".verification-*")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func (s *commandReceiptStore) DeletePod(podKey string) error {
	return os.RemoveAll(filepath.Join(s.root, receiptHash(podKey)))
}

func (s *commandReceiptStore) receiptPath(
	podKey, kind, commandID string,
) (string, error) {
	if s == nil || s.root == "" {
		return "", fmt.Errorf("command receipt root is not configured")
	}
	dir := filepath.Join(s.root, receiptHash(podKey))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, kind+"-"+receiptHash(commandID)+".json"), nil
}

func receiptHash(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}
