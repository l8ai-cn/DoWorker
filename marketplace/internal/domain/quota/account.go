package quota

import (
	"errors"
	"strings"
)

var (
	ErrInvalidAccount = errors.New("invalid quota account")
	ErrInvalidAmount  = errors.New("invalid quota amount")
	ErrInsufficient   = errors.New("insufficient quota")
)

type Account struct {
	id        string
	available int64
	reserved  int64
	consumed  int64
}

func NewAccount(id string, available int64) (*Account, error) {
	if strings.TrimSpace(id) == "" || available < 0 {
		return nil, ErrInvalidAccount
	}
	return &Account{id: id, available: available}, nil
}

func (a *Account) Reserve(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.available {
		return ErrInsufficient
	}
	a.available -= amount
	a.reserved += amount
	return nil
}

func (a *Account) Settle(amount int64) error {
	if amount <= 0 || amount > a.reserved {
		return ErrInvalidAmount
	}
	a.reserved -= amount
	a.consumed += amount
	return nil
}

func (a *Account) Release(amount int64) error {
	if amount <= 0 || amount > a.reserved {
		return ErrInvalidAmount
	}
	a.reserved -= amount
	a.available += amount
	return nil
}

func (a Account) ID() string       { return a.id }
func (a Account) Available() int64 { return a.available }
func (a Account) Reserved() int64  { return a.reserved }
func (a Account) Consumed() int64  { return a.consumed }
