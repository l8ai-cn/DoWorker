package workerdependency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRejectsMaterializedSecretLikeData(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Document)
		match  string
	}{
		{
			name: "sensitive runtime field",
			mutate: func(document *Document) {
				document.RuntimeBundles[0].Values[0].Name = "SENTRY_AUTH_TOKEN"
				document.RuntimeBundles[0].ContentDigest = mustRuntimeValuesDigest(
					t,
					document.RuntimeBundles[0].Values,
				)
			},
			match: "must remain a Secret reference",
		},
		{
			name: "strong token in runtime value",
			mutate: func(document *Document) {
				document.RuntimeBundles[0].Values[0].Value = "sk-live-do-not-store"
				document.RuntimeBundles[0].ContentDigest = mustRuntimeValuesDigest(
					t,
					document.RuntimeBundles[0].Values,
				)
			},
			match: "contains raw secret-like data",
		},
		{
			name: "nested Secret key in config JSON",
			mutate: func(document *Document) {
				document.RuntimeBundles[2].Values[0].Value =
					`{"apiToken":"plain-text-secret"}`
				document.RuntimeBundles[2].ContentDigest = mustRuntimeValuesDigest(
					t,
					document.RuntimeBundles[2].Values,
				)
			},
			match: "contains raw secret-like data",
		},
		{
			name: "strong token in AgentFile",
			mutate: func(document *Document) {
				document.Worker.AgentfileSource =
					"AGENT codex\nCONFIG note STRING = \"sk-live-do-not-store\"\nMODE pty\n"
				document.Worker.AgentfileSourceDigest = TextDigest(
					document.Worker.AgentfileSource,
				)
			},
			match: "agentfile source contains raw secret-like data",
		},
		{
			name: "sensitive undeclared AgentFile field",
			mutate: func(document *Document) {
				document.Worker.AgentfileSource =
					"AGENT codex\nENV UNDECLARED_TOKEN = \"opaque-value\"\nMODE pty\n"
				document.Worker.AgentfileSourceDigest = TextDigest(
					document.Worker.AgentfileSource,
				)
			},
			match: "must use a live reference",
		},
		{
			name: "repository token query",
			mutate: func(document *Document) {
				document.Repository.HTTPCloneURL =
					"https://git.example.com/repo.git?token=plain-text"
			},
			match: "clone endpoint contains raw secret-like data",
		},
		{
			name: "repository URL user info",
			mutate: func(document *Document) {
				document.Repository.HTTPCloneURL =
					"https://user:hunter2@git.example.com/repo.git"
			},
			match: "clone endpoint contains raw secret-like data",
		},
		{
			name: "model token query",
			mutate: func(document *Document) {
				document.Models.Primary.BaseURL =
					"https://api.example.com/v1?token=plain-text"
			},
			match: "model base url contains raw secret-like data",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := validDocument(t)
			test.mutate(&document)

			err := Validate(document)

			require.ErrorContains(t, err, test.match)
		})
	}
}
