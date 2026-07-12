package installation

import (
	"errors"
	"strings"
)

type Status string
type OperationStatus string

const (
	StatusPlanning    Status = "planning"
	StatusInstalling  Status = "installing"
	StatusActive      Status = "active"
	StatusFailed      Status = "failed"
	StatusSuspended   Status = "suspended"
	StatusUninstalled Status = "uninstalled"
)

const (
	OperationPlanned   OperationStatus = "planned"
	OperationRunning   OperationStatus = "running"
	OperationSucceeded OperationStatus = "succeeded"
	OperationFailed    OperationStatus = "failed"
)

var (
	ErrInvalidInstallation = errors.New("invalid installation")
	ErrInvalidOperation    = errors.New("invalid installation operation")
	ErrInvalidTransition   = errors.New("invalid installation transition")
)

type Installation struct {
	id                   string
	marketplaceID        int64
	listingID            int64
	listingVersionID     int64
	entitlementID        string
	targetOrganizationID int64
	installedByUserID    int64
	status               Status
	runtimeRef           string
}

func New(
	id string,
	marketplaceID, listingID, listingVersionID int64,
	entitlementID string,
	targetOrganizationID, installedByUserID int64,
) (*Installation, error) {
	if strings.TrimSpace(id) == "" || marketplaceID <= 0 || listingID <= 0 ||
		listingVersionID <= 0 || strings.TrimSpace(entitlementID) == "" ||
		targetOrganizationID <= 0 || installedByUserID <= 0 {
		return nil, ErrInvalidInstallation
	}
	return &Installation{
		id: id, marketplaceID: marketplaceID, listingID: listingID,
		listingVersionID: listingVersionID, entitlementID: entitlementID,
		targetOrganizationID: targetOrganizationID,
		installedByUserID:    installedByUserID, status: StatusPlanning,
	}, nil
}

func (i *Installation) Plan(
	operationID, planID, planDigest string,
	plan []byte,
) (*Operation, error) {
	if i.status != StatusPlanning || strings.TrimSpace(operationID) == "" ||
		strings.TrimSpace(planID) == "" || strings.TrimSpace(planDigest) == "" ||
		len(plan) == 0 {
		return nil, ErrInvalidOperation
	}
	return &Operation{
		id: operationID, planID: planID, planDigest: planDigest,
		plan: append([]byte(nil), plan...), status: OperationPlanned,
		installation: i,
	}, nil
}

func (i *Installation) Activate(runtimeRef string) error {
	if i.status != StatusInstalling || strings.TrimSpace(runtimeRef) == "" {
		return ErrInvalidTransition
	}
	i.runtimeRef = runtimeRef
	i.status = StatusActive
	return nil
}

func (i *Installation) Fail() error {
	if i.status != StatusInstalling {
		return ErrInvalidTransition
	}
	i.status = StatusFailed
	return nil
}

func (i Installation) Status() Status { return i.status }

type Operation struct {
	id           string
	planID       string
	planDigest   string
	plan         []byte
	status       OperationStatus
	result       []byte
	errorCode    string
	errorMessage string
	installation *Installation
}

func (o *Operation) Start() error {
	if o.status != OperationPlanned || o.installation.status != StatusPlanning {
		return ErrInvalidTransition
	}
	o.status = OperationRunning
	o.installation.status = StatusInstalling
	return nil
}

func (o *Operation) Succeed(result []byte) error {
	if o.status != OperationRunning || len(result) == 0 {
		return ErrInvalidTransition
	}
	o.status = OperationSucceeded
	o.result = append([]byte(nil), result...)
	return nil
}

func (o *Operation) Fail(code, message string) error {
	if o.status != OperationRunning || strings.TrimSpace(code) == "" ||
		strings.TrimSpace(message) == "" {
		return ErrInvalidTransition
	}
	o.status = OperationFailed
	o.errorCode = code
	o.errorMessage = message
	return nil
}

func (o Operation) Status() OperationStatus { return o.status }
