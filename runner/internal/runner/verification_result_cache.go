package runner

import (
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
)

const verificationResultCacheSize = 64

func (h *RunnerMessageHandler) verificationResult(
	command *runnerv1.RunVerificationCommand,
) (*runnerv1.VerificationResultEvent, error) {
	requestID := command.GetRequestId()
	if requestID == "" {
		return h.runVerification(command), nil
	}
	cacheKey := command.GetPodKey() + "\x00" + requestID
	value, err, _ := h.verificationRuns.Do(cacheKey, func() (any, error) {
		if cached := h.cachedVerificationResult(cacheKey); cached != nil {
			return cached, nil
		}
		if h.receipts != nil {
			cached, found, err := h.receipts.LoadVerification(
				command.GetPodKey(),
				requestID,
			)
			if err != nil {
				return nil, err
			}
			if found {
				h.storeVerificationResult(cacheKey, cached)
				return cloneVerificationResult(cached), nil
			}
		}
		result := h.runVerification(command)
		if h.receipts != nil {
			if err := h.receipts.StoreVerification(
				command.GetPodKey(),
				requestID,
				result,
			); err != nil {
				return nil, err
			}
		}
		h.storeVerificationResult(cacheKey, result)
		return cloneVerificationResult(result), nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*runnerv1.VerificationResultEvent), nil
}

func (h *RunnerMessageHandler) cachedVerificationResult(
	requestID string,
) *runnerv1.VerificationResultEvent {
	h.verificationMu.Lock()
	defer h.verificationMu.Unlock()
	return cloneVerificationResult(h.verificationCache[requestID])
}

func (h *RunnerMessageHandler) storeVerificationResult(
	requestID string,
	result *runnerv1.VerificationResultEvent,
) {
	h.verificationMu.Lock()
	defer h.verificationMu.Unlock()
	if h.verificationCache == nil {
		h.verificationCache = make(map[string]*runnerv1.VerificationResultEvent)
	}
	if _, exists := h.verificationCache[requestID]; !exists {
		if len(h.verificationOrder) == verificationResultCacheSize {
			delete(h.verificationCache, h.verificationOrder[0])
			h.verificationOrder = h.verificationOrder[1:]
		}
		h.verificationOrder = append(h.verificationOrder, requestID)
	}
	h.verificationCache[requestID] = cloneVerificationResult(result)
}

func cloneVerificationResult(
	result *runnerv1.VerificationResultEvent,
) *runnerv1.VerificationResultEvent {
	if result == nil {
		return nil
	}
	return proto.Clone(result).(*runnerv1.VerificationResultEvent)
}
