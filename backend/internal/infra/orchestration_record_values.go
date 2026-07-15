package infra

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func corruptRecord(record string) error {
	return fmt.Errorf("%w: %s", orchestrationcontrol.ErrCorrupt, record)
}

func structTypeMeta(apiVersion, kind string) orchestrationresource.TypeMeta {
	return orchestrationresource.TypeMeta{APIVersion: apiVersion, Kind: kind}
}

func stringSlug(value string) slugkit.Slug {
	return slugkit.Slug(value)
}

func decodeStrictJSON(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return orchestrationcontrol.ErrCorrupt
		}
		return err
	}
	return nil
}
