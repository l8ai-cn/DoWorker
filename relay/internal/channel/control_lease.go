package channel

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/l8ai-cn/agentcloud/relay/internal/protocol"
)

func (c *Channel) handleControlLeaseRequest(subscriberID string, request protocol.ControlLeaseRequest) {
	switch request.Action {
	case protocol.ControlLeaseActionAcquire:
		c.acquireControlLease(subscriberID)
	case protocol.ControlLeaseActionRenew:
		c.renewControlLease(subscriberID, request.LeaseID)
	case protocol.ControlLeaseActionRelease:
		c.releaseControlLease(subscriberID, request.LeaseID)
	default:
		c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusRequired, "", time.Time{})
	}
}

func (c *Channel) acquireControlLease(subscriberID string) {
	leaseID, err := newControlLeaseID()
	if err != nil {
		c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusRequired, "", time.Time{})
		return
	}

	now := time.Now()
	c.controlMu.Lock()
	if c.IsClosed() {
		c.controlMu.Unlock()
		return
	}
	if c.controlOwner != "" && now.Before(c.controlExpiresAt) {
		if c.controlOwner == subscriberID {
			leaseID = c.controlLeaseID
			expiresAt := c.controlExpiresAt
			c.controlMu.Unlock()
			c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusGranted, leaseID, expiresAt)
			return
		}
		c.controlMu.Unlock()
		c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusBusy, "", time.Time{})
		return
	}
	c.setControlLeaseLocked(subscriberID, leaseID, now)
	expiresAt := c.controlExpiresAt
	c.controlMu.Unlock()
	c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusGranted, leaseID, expiresAt)
}

func (c *Channel) renewControlLease(subscriberID, leaseID string) {
	now := time.Now()
	c.controlMu.Lock()
	if c.IsClosed() ||
		c.controlOwner != subscriberID ||
		c.controlLeaseID != leaseID ||
		!now.Before(c.controlExpiresAt) {
		c.controlMu.Unlock()
		c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusRequired, "", time.Time{})
		return
	}
	c.setControlLeaseLocked(subscriberID, leaseID, now)
	expiresAt := c.controlExpiresAt
	c.controlMu.Unlock()
	c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusGranted, leaseID, expiresAt)
}

func (c *Channel) releaseControlLease(subscriberID, leaseID string) {
	c.controlMu.Lock()
	if c.controlOwner != subscriberID || c.controlLeaseID != leaseID {
		c.controlMu.Unlock()
		c.sendControlStatus(subscriberID, protocol.ControlLeaseStatusRequired, "", time.Time{})
		return
	}
	c.clearControlLeaseLocked()
	c.controlMu.Unlock()
	c.Broadcast(protocol.EncodeControlLeaseStatus(protocol.ControlLeaseStatusReleased, "", 0))
}

func (c *Channel) writeToPublisherWithControlLease(
	subscriberID string,
	data []byte,
) (bool, error) {
	c.controlMu.Lock()
	now := time.Now()
	if c.controlOwner == subscriberID && now.Before(c.controlExpiresAt) {
		err := c.writeToPublisher(data)
		c.controlMu.Unlock()
		return true, err
	}
	expired := c.controlOwner != "" && !now.Before(c.controlExpiresAt)
	if expired {
		c.clearControlLeaseLocked()
	}
	c.controlMu.Unlock()
	if expired {
		c.Broadcast(protocol.EncodeControlLeaseStatus(protocol.ControlLeaseStatusExpired, "", 0))
	}
	return false, nil
}

func (c *Channel) releaseControlLeaseForSubscriber(subscriberID string) {
	c.controlMu.Lock()
	if c.controlOwner != subscriberID {
		c.controlMu.Unlock()
		return
	}
	c.clearControlLeaseLocked()
	c.controlMu.Unlock()
	c.Broadcast(protocol.EncodeControlLeaseStatus(protocol.ControlLeaseStatusReleased, "", 0))
}

func (c *Channel) stopControlLease() {
	c.controlMu.Lock()
	c.clearControlLeaseLocked()
	c.controlMu.Unlock()
}

func (c *Channel) setControlLeaseLocked(subscriberID, leaseID string, now time.Time) {
	if c.controlTimer != nil {
		c.controlTimer.Stop()
	}
	duration := c.config.ControlLeaseDuration
	if duration <= 0 {
		duration = DefaultChannelConfig().ControlLeaseDuration
	}
	c.controlOwner = subscriberID
	c.controlLeaseID = leaseID
	c.controlExpiresAt = now.Add(duration)
	c.controlEpoch++
	epoch := c.controlEpoch
	c.controlTimer = time.AfterFunc(duration, func() {
		c.expireControlLease(epoch)
	})
}

func (c *Channel) expireControlLease(epoch uint64) {
	c.controlMu.Lock()
	if c.controlOwner == "" || c.controlEpoch != epoch {
		c.controlMu.Unlock()
		return
	}
	c.clearControlLeaseLocked()
	c.controlMu.Unlock()
	if c.IsClosed() {
		return
	}
	c.Broadcast(protocol.EncodeControlLeaseStatus(protocol.ControlLeaseStatusExpired, "", 0))
}

func (c *Channel) clearControlLeaseLocked() {
	if c.controlTimer != nil {
		c.controlTimer.Stop()
		c.controlTimer = nil
	}
	c.controlOwner = ""
	c.controlLeaseID = ""
	c.controlExpiresAt = time.Time{}
	c.controlEpoch++
}

func (c *Channel) sendControlStatus(subscriberID, status, leaseID string, expiresAt time.Time) {
	c.subscribersMu.RLock()
	subscriber := c.subscribers[subscriberID]
	c.subscribersMu.RUnlock()
	if subscriber == nil {
		return
	}
	var expiresAtMillis int64
	if !expiresAt.IsZero() {
		expiresAtMillis = expiresAt.UnixMilli()
	}
	if err := subscriber.WriteMessage(protocol.EncodeControlLeaseStatus(status, leaseID, expiresAtMillis)); err != nil {
		c.RemoveSubscriber(subscriberID)
	}
}

func newControlLeaseID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
