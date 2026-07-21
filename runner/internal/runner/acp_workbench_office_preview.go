package runner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/l8ai-cn/agentcloud/runner/internal/safego"
	"github.com/l8ai-cn/agentcloud/runner/internal/workbench"
)

func (f *acpWorkbenchForwarder) queueOfficePreview(
	artifact *workbench.ArtifactDescriptor,
) {
	source, ok := workbench.ResolveOfficePreviewSource(artifact)
	if !ok {
		return
	}
	key := artifact.GetArtifactId() + ":" +
		strconv.FormatUint(artifact.GetRevision(), 10)
	f.previewMu.Lock()
	if _, exists := f.converting[key]; exists {
		f.previewMu.Unlock()
		return
	}
	f.converting[key] = struct{}{}
	f.previewMu.Unlock()

	processing := workbench.OfficePreviewProcessing(artifact, source)
	f.send(f.mapper.Artifacts([]*workbench.ArtifactDescriptor{processing}))
	safego.Go("office-artifact-preview", func() {
		defer f.finishOfficePreview(key)
		f.convertAndPublishOfficePreview(processing, source)
	})
}

func (f *acpWorkbenchForwarder) convertAndPublishOfficePreview(
	processing *workbench.ArtifactDescriptor,
	source workbench.OfficePreviewSource,
) {
	pdf, err := f.convertOffice(context.Background(), f.workDir, source.Path)
	if err != nil {
		f.failOfficePreview(processing, err)
		return
	}
	digest, err := officePreviewSourceDigest(f.workDir, source.Path)
	if err != nil {
		f.failOfficePreview(processing, err)
		return
	}
	if digest != source.Digest {
		f.failOfficePreview(processing, fmt.Errorf(
			"office preview source changed during conversion",
		))
		return
	}
	previewPath, previewDigest, previewBytes, err := writeOfficePreviewArtifact(
		f.workDir,
		processing.GetArtifactId(),
		processing.GetRevision(),
		pdf,
	)
	if err != nil {
		f.failOfficePreview(processing, err)
		return
	}
	if !f.isCurrentArtifact(processing) {
		return
	}
	ready := workbench.OfficePreviewReady(
		processing,
		"workspace:"+previewPath,
		previewDigest,
		previewBytes,
	)
	f.send(f.mapper.Artifacts([]*workbench.ArtifactDescriptor{ready}))
}

func (f *acpWorkbenchForwarder) failOfficePreview(
	processing *workbench.ArtifactDescriptor,
	err error,
) {
	if !f.isCurrentArtifact(processing) {
		return
	}
	f.send(f.mapper.Artifacts([]*workbench.ArtifactDescriptor{
		workbench.OfficePreviewFailed(processing),
	}))
	f.send(f.mapper.Unsupported("artifact.office_preview.failed", map[string]string{
		"artifact_id": processing.GetArtifactId(),
		"error":       err.Error(),
	}))
}

func (f *acpWorkbenchForwarder) recordArtifactRevision(
	artifact *workbench.ArtifactDescriptor,
) {
	f.previewMu.Lock()
	f.latestRevision[artifact.GetArtifactId()] = artifact.GetRevision()
	f.previewMu.Unlock()
}

func (f *acpWorkbenchForwarder) isCurrentArtifact(
	artifact *workbench.ArtifactDescriptor,
) bool {
	f.previewMu.Lock()
	defer f.previewMu.Unlock()
	return f.latestRevision[artifact.GetArtifactId()] == artifact.GetRevision()
}

func (f *acpWorkbenchForwarder) finishOfficePreview(key string) {
	f.previewMu.Lock()
	delete(f.converting, key)
	f.previewMu.Unlock()
}
