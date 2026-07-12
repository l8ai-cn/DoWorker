package quota

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountReserveSettleAndRelease(t *testing.T) {
	account, err := NewAccount("account-1", 100_000_000)
	require.NoError(t, err)

	require.NoError(t, account.Reserve(30_000_000))
	require.Equal(t, int64(70_000_000), account.Available())
	require.Equal(t, int64(30_000_000), account.Reserved())

	require.NoError(t, account.Settle(20_000_000))
	require.Equal(t, int64(70_000_000), account.Available())
	require.Equal(t, int64(10_000_000), account.Reserved())
	require.Equal(t, int64(20_000_000), account.Consumed())

	require.NoError(t, account.Release(10_000_000))
	require.Equal(t, int64(80_000_000), account.Available())
	require.Zero(t, account.Reserved())
}

func TestAccountRejectsInsufficientQuota(t *testing.T) {
	account, err := NewAccount("account-1", 5_000_000)
	require.NoError(t, err)

	err = account.Reserve(6_000_000)
	require.ErrorIs(t, err, ErrInsufficient)
	require.Equal(t, int64(5_000_000), account.Available())
	require.Zero(t, account.Reserved())
}
