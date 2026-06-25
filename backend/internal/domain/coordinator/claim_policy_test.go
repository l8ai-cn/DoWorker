package coordinator

import "testing"

func TestClaimPolicyMatches(t *testing.T) {
	cases := []struct {
		name   string
		policy ClaimPolicy
		cand   Candidate
		want   bool
	}{
		{
			name:   "empty policy matches anything",
			policy: ClaimPolicy{},
			cand:   Candidate{State: "open"},
			want:   true,
		},
		{
			name:   "label must be present",
			policy: ClaimPolicy{Labels: []string{"bug"}},
			cand:   Candidate{Labels: []string{"BUG", "p1"}},
			want:   true,
		},
		{
			name:   "missing label rejects",
			policy: ClaimPolicy{Labels: []string{"bug"}},
			cand:   Candidate{Labels: []string{"feature"}},
			want:   false,
		},
		{
			name:   "state mismatch rejects",
			policy: ClaimPolicy{States: []string{"open"}},
			cand:   Candidate{State: "closed"},
			want:   false,
		},
		{
			name:   "unassigned only rejects assigned",
			policy: ClaimPolicy{UnassignedOnly: true},
			cand:   Candidate{Assignees: []string{"alice"}},
			want:   false,
		},
		{
			name:   "title keyword matches case-insensitively",
			policy: ClaimPolicy{TitleKeywords: []string{"crash"}},
			cand:   Candidate{Title: "App CRASHes on login"},
			want:   true,
		},
		{
			name:   "body keyword absent rejects",
			policy: ClaimPolicy{BodyKeywords: []string{"stacktrace"}},
			cand:   Candidate{Description: "no logs here"},
			want:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := tc.policy.Matches(tc.cand)
			if got != tc.want {
				t.Fatalf("Matches = %v (reason %q), want %v", got, reason, tc.want)
			}
		})
	}
}

func TestDecodeClaimPolicyMergesLabelFilter(t *testing.T) {
	p := Project{
		LabelFilter: []string{"auto"},
		ClaimPolicy: []byte(`{"states":["open"],"labels":["bug"]}`),
	}
	policy := p.DecodeClaimPolicy()
	if !containsFold(policy.Labels, "auto") || !containsFold(policy.Labels, "bug") {
		t.Fatalf("labels = %v, want bug+auto merged", policy.Labels)
	}
	if len(policy.States) != 1 || policy.States[0] != "open" {
		t.Fatalf("states = %v, want [open]", policy.States)
	}
}
