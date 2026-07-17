package client

import "context"

func (r *REST) SendPodPrompt(ctx context.Context, orgSlug, podKey, prompt string) error {
	request := map[string]string{
		"orgSlug": orgSlug,
		"podKey":  podKey,
		"prompt":  prompt,
	}
	return r.connectCall(ctx, "/proto.pod.v1.PodService/SendPodPrompt", request, nil)
}
