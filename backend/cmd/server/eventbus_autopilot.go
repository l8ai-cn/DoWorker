package main

import (
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func protoToEventbusThinking(data *runnerv1.AutopilotThinkingEvent) *eventsv1.AutopilotThinkingEventData {
	result := &eventsv1.AutopilotThinkingEventData{
		AutopilotControllerKey: data.GetAutopilotKey(),
		Iteration:              data.GetIteration(),
		DecisionType:           data.GetDecisionType(),
		Reasoning:              data.GetReasoning(),
		Confidence:             data.GetConfidence(),
	}

	if action := data.GetAction(); action != nil {
		result.Action = &eventsv1.AutopilotActionEventData{
			Type:    action.GetType(),
			Content: action.GetContent(),
			Reason:  action.GetReason(),
		}
	}

	if progress := data.GetProgress(); progress != nil {
		result.Progress = &eventsv1.AutopilotProgressEventData{
			Summary:        progress.GetSummary(),
			CompletedSteps: progress.GetCompletedSteps(),
			RemainingSteps: progress.GetRemainingSteps(),
			Percent:        progress.GetPercent(),
		}
	}

	if helpReq := data.GetHelpRequest(); helpReq != nil {
		result.HelpRequest = &eventsv1.AutopilotHelpRequestEventData{
			Reason:          helpReq.GetReason(),
			Context:         helpReq.GetContext(),
			TerminalExcerpt: helpReq.GetTerminalExcerpt(),
		}
		for _, s := range helpReq.GetSuggestions() {
			result.HelpRequest.Suggestions = append(result.HelpRequest.Suggestions, &eventsv1.AutopilotHelpSuggestionEventData{
				Action: s.GetAction(),
				Label:  s.GetLabel(),
			})
		}
	}

	return result
}
