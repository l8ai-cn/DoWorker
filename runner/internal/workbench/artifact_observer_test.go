package workbench

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestArtifactObserverEmitsOnlyChangedDeliverables(t *testing.T) {
	root := t.TempDir()
	writeArtifactFile(t, root, "output/existing.png", "baseline")
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)

	writeArtifactFile(t, root, "output/demo.mp4", "video")
	writeArtifactFile(t, root, "src/ignored.go", "package ignored")

	artifacts, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(t, "workspace:output/demo.mp4", artifacts[0].GetArtifactId())
	require.Equal(t, "video/mp4", artifacts[0].GetMediaType())
	require.Equal(t, "workspace:output/demo.mp4",
		artifacts[0].GetRepresentations()[0].GetTransport().GetResourceId())
	require.Nil(t, artifacts[0].GetManifest())

	unchanged, err := observer.Scan()
	require.NoError(t, err)
	require.Empty(t, unchanged)
}

func TestArtifactObserverTracksRevisionAndDeletion(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "deliverables/result.html", "<h1>one</h1>")

	created, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, created, 1)
	require.Equal(t, uint64(1), created[0].GetRevision())
	require.Equal(t, agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_READY,
		created[0].GetStatus())

	writeArtifactFile(t, root, "deliverables/result.html", "<h1>two</h1>")
	updated, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, uint64(2), updated[0].GetRevision())
	require.NotEqual(t,
		created[0].GetRevisions()[0].GetDigest(),
		updated[0].GetRevisions()[0].GetDigest())

	require.NoError(t, os.Remove(filepath.Join(root, "deliverables/result.html")))
	deleted, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, deleted, 1)
	require.Equal(t, uint64(3), deleted[0].GetRevision())
	require.Equal(t, agentworkbenchv2.ArtifactStatus_ARTIFACT_STATUS_DELETED,
		deleted[0].GetStatus())
}

func TestArtifactObserverReportsModifiedBaselineFile(t *testing.T) {
	root := t.TempDir()
	writeArtifactFile(t, root, "output/existing.png", "baseline")
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)

	writeArtifactFile(t, root, "output/existing.png", "changed")
	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(t, "workspace:output/existing.png", artifacts[0].GetArtifactId())
}

func TestArtifactObserverProjectsWordDocument(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "outputs/report.docx", "word")

	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(t, "workspace:outputs/report.docx", artifacts[0].GetArtifactId())
	require.Equal(t,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		artifacts[0].GetMediaType())
}

func TestArtifactObserverProjectsSpreadsheetCsvAndAudio(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "outputs/data.xlsx", "sheet")
	writeArtifactFile(t, root, "outputs/results.csv", "name,value\nnorth,10")
	writeArtifactFile(t, root, "outputs/briefing.mp3", "audio")

	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, 3)
	mediaTypes := map[string]string{}
	for _, artifact := range artifacts {
		mediaTypes[artifact.GetFilename()] = artifact.GetMediaType()
	}
	require.Equal(t,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		mediaTypes["data.xlsx"])
	require.Equal(t, "text/csv", mediaTypes["results.csv"])
	require.Equal(t, "audio/mpeg", mediaTypes["briefing.mp3"])
}

func TestArtifactObserverIgnoresDerivedPreviewStorage(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(
		t,
		root,
		".do-worker/workbench/previews/report-r1.pdf",
		"preview",
	)

	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Empty(t, artifacts)
}

func TestArtifactObserverProjectsDeclaredImageEditArtifact(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/source.png", "source")
	writeArtifactFile(t, root, "output/result.png", "result")
	writeArtifactDeclaration(t, root, "image-edit.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"image-edit-result",
		"revision":1,
		"role":"image_edit",
		"primary_representation_id":"result",
		"producer":{"namespace":"openai.codex","type":"image.edit"},
		"representations":[
			{"representation_id":"source","path":"output/source.png","media_type":"image/png","role":"source"},
			{"representation_id":"result","path":"output/result.png","media_type":"image/png","role":"result"}
		],
		"manifest":{
			"kind":"image_edit",
			"source_representation_id":"source",
			"result_representation_id":"result",
			"candidate_representation_ids":["result"],
			"source_width":1280,
			"source_height":720
		}
	}`)

	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	artifact := artifacts[0]
	require.Equal(t, "image-edit-result", artifact.GetArtifactId())
	require.Equal(t, "result.png", artifact.GetFilename())
	require.Equal(t, "image/png", artifact.GetMediaType())
	require.Empty(t, artifact.GetProvenance().GetToolExecutionId())
	require.Len(t, artifact.GetRepresentations(), 2)
	require.Equal(t,
		"workspace:output/source.png",
		artifactRepresentation(t, artifact, "source").GetTransport().GetResourceId())
	require.Equal(t,
		"result",
		artifact.GetManifest().GetImageEdit().GetResultRepresentationId())
	require.Equal(t, uint32(1280),
		artifact.GetManifest().GetImageEdit().GetSourceWidth())
}

func TestArtifactObserverProjectsDeclaredPresentationAndVideo(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	for path, content := range map[string]string{
		"output/deck.pptx":         "deck",
		"output/slide-1.png":       "page",
		"output/slide-1-thumb.png": "thumb",
		"output/original.mov":      "original",
		"output/playable.mp4":      validMP4Fixture("playable"),
		"output/poster.png":        "poster",
	} {
		writeArtifactFile(t, root, path, content)
	}
	writeArtifactDeclaration(t, root, "presentation.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"quarterly-review",
		"revision":1,
		"role":"presentation",
		"primary_representation_id":"deck",
		"producer":{"namespace":"openai.codex","type":"presentation.create"},
		"representations":[
			{"representation_id":"deck","path":"output/deck.pptx","media_type":"application/vnd.openxmlformats-officedocument.presentationml.presentation","role":"source"},
			{"representation_id":"page-1","path":"output/slide-1.png","media_type":"image/png","role":"page"},
			{"representation_id":"thumb-1","path":"output/slide-1-thumb.png","media_type":"image/png","role":"thumbnail"}
		],
		"manifest":{
			"kind":"presentation",
			"deck_revision":1,
			"slides":[{"slide_id":"slide-1","position":1,"title":"Overview","page_representation_id":"page-1","thumbnail_representation_id":"thumb-1"}],
			"versions":[{"version_id":"v1","revision":1,"label":"Initial"}],
			"selected_version_id":"v1"
		}
	}`)
	writeArtifactDeclaration(t, root, "video.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"launch-video",
		"revision":1,
		"role":"video",
		"primary_representation_id":"playable",
		"producer":{"namespace":"seedance","type":"video.generate","id":"task-1"},
		"representations":[
			{"representation_id":"original","path":"output/original.mov","media_type":"video/quicktime","role":"original"},
			{"representation_id":"playable","path":"output/playable.mp4","media_type":"video/mp4","role":"playable","duration_millis":2000},
			{"representation_id":"poster","path":"output/poster.png","media_type":"image/png","role":"poster"}
		],
		"manifest":{
			"kind":"video",
			"stage":"ready",
			"duration_millis":2000,
			"original_representation_id":"original",
			"playable_representation_id":"playable",
			"poster_representation_id":"poster"
		}
	}`)

	artifacts, err := observer.Scan()

	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	require.NotNil(t, artifacts[0].GetManifest().GetVideo())
	require.NotNil(t, artifacts[1].GetManifest().GetPresentation())
	require.Equal(t, "page-1",
		artifacts[1].GetManifest().GetPresentation().GetSlides()[0].GetPageRepresentationId())
}

func TestArtifactObserverRejectsInvalidDeclaration(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactDeclaration(t, root, "invalid.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v0",
		"artifact_id":"invalid-artifact",
		"revision":1,
		"role":"preview",
		"primary_representation_id":"file",
		"producer":{"namespace":"test","type":"invalid"},
		"representations":[
			{"representation_id":"file","path":"../secret.png","media_type":"image/png"}
		]
	}`)

	_, err = observer.Scan()

	require.ErrorContains(t, err, "schema_version")
}

func TestArtifactObserverRejectsEmptyDeclaredFile(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/empty.mp4", "")
	writeArtifactDeclaration(t, root, "video.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"empty-video",
		"revision":1,
		"role":"video",
		"primary_representation_id":"playable",
		"producer":{"namespace":"seedance","type":"video.generate","id":"task-1"},
		"representations":[
			{"representation_id":"playable","path":"output/empty.mp4","media_type":"video/mp4","role":"playable"}
		],
		"manifest":{
			"kind":"video",
			"stage":"ready",
			"playable_representation_id":"playable"
		}
	}`)

	_, err = observer.Scan()

	require.ErrorContains(t, err, `workspace path "output/empty.mp4" is empty`)
}

func TestArtifactObserverRejectsSeedanceVideoWithoutProducerID(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/video.mp4", validMP4Fixture("video"))
	writeArtifactDeclaration(t, root, "video.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"video",
		"revision":1,
		"role":"video",
		"primary_representation_id":"playable",
		"producer":{"namespace":"seedance","type":"video.generate"},
		"representations":[
			{"representation_id":"playable","path":"output/video.mp4","media_type":"video/mp4"}
		]
	}`)

	_, err = observer.Scan()

	require.ErrorContains(t, err, "seedance video.generate requires producer.id")
}

func TestArtifactObserverRejectsTextDisguisedAsMP4(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/video.mp4", "not a video file")
	writeArtifactDeclaration(t, root, "video.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"video",
		"revision":1,
		"role":"video",
		"primary_representation_id":"playable",
		"producer":{"namespace":"seedance","type":"video.generate","id":"task-1"},
		"representations":[
			{"representation_id":"playable","path":"output/video.mp4","media_type":"video/mp4"}
		]
	}`)

	_, err = observer.Scan()

	require.ErrorContains(t, err, `workspace path "output/video.mp4" is not a decodable MP4 file`)
}

func TestArtifactObserverRejectsMP4WithoutVideoTrack(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(
		t,
		root,
		"output/video.mp4",
		"\x00\x00\x00\x18ftypisom\x00\x00\x02\x00isomiso2"+
			"\x00\x00\x00\x08moov\x00\x00\x00\x0cmdatdata",
	)
	writeArtifactDeclaration(t, root, "video.json", `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"video",
		"revision":1,
		"role":"video",
		"primary_representation_id":"playable",
		"producer":{"namespace":"seedance","type":"video.generate","id":"task-1"},
		"representations":[
			{"representation_id":"playable","path":"output/video.mp4","media_type":"video/mp4"}
		],
		"manifest":{
			"kind":"video",
			"stage":"ready",
			"playable_representation_id":"playable"
		}
	}`)

	_, err = observer.Scan()

	require.ErrorContains(t, err, `workspace path "output/video.mp4" is not a decodable MP4 file`)
}

func TestArtifactObserverRequiresDeclaredRevisionIncrement(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.html", "<h1>one</h1>")
	writeArtifactDeclaration(t, root, "page.json", declaredPageArtifact(1))

	first, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, uint64(1), first[0].GetRevision())

	writeArtifactFile(t, root, "output/result.html", "<h1>two</h1>")
	_, err = observer.Scan()
	require.ErrorContains(t, err, "revision")

	writeArtifactDeclaration(t, root, "page.json", declaredPageArtifact(2))
	second, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Equal(t, uint64(2), second[0].GetRevision())
}

func TestArtifactObserverRejectsProducerChangeAcrossRevisions(t *testing.T) {
	root := t.TempDir()
	observer, err := NewArtifactObserver(root)
	require.NoError(t, err)
	writeArtifactFile(t, root, "output/result.html", "<h1>one</h1>")
	writeArtifactDeclaration(t, root, "page.json", declaredPageArtifact(1))
	first, err := observer.Scan()
	require.NoError(t, err)
	require.Len(t, first, 1)

	writeArtifactFile(t, root, "output/result.html", "<h1>two</h1>")
	writeArtifactDeclaration(t, root, "page.json", declaredPageArtifactForProducer(
		2,
		"web.update",
	))

	_, err = observer.Scan()

	require.ErrorContains(t, err, "producer must remain stable")
}

func writeArtifactFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func validMP4Fixture(payload string) string {
	file, err := os.CreateTemp("", "agentsmesh-test-*.mp4")
	if err != nil {
		panic(err)
	}
	path := file.Name()
	file.Close()
	defer os.Remove(path)
	color := testVideoColor(payload)
	output, err := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c="+color+":s=16x16:d=0.2:r=5",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		path,
	).CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("generate MP4 fixture: %s", string(output)))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func testVideoColor(seed string) string {
	sum := sha1.Sum([]byte(seed))
	return "0x" + hex.EncodeToString(sum[:3])
}

func writeArtifactDeclaration(t *testing.T, root, name, content string) {
	t.Helper()
	writeArtifactFile(
		t,
		root,
		filepath.ToSlash(filepath.Join(".do-worker/workbench/artifacts", name)),
		content,
	)
}

func declaredPageArtifact(revision int) string {
	return declaredPageArtifactForProducer(revision, "web.create")
}

func declaredPageArtifactForProducer(revision int, producerType string) string {
	return fmt.Sprintf(`{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"rendered-page",
		"revision":%d,
		"role":"preview",
		"primary_representation_id":"page",
		"producer":{"namespace":"openai.codex","type":%q},
		"representations":[
			{"representation_id":"page","path":"output/result.html","media_type":"text/html","role":"primary"}
		]
	}`, revision, producerType)
}

func artifactRepresentation(
	t *testing.T,
	artifact *agentworkbenchv2.ArtifactDescriptor,
	id string,
) *agentworkbenchv2.ArtifactRepresentation {
	t.Helper()
	for _, representation := range artifact.GetRepresentations() {
		if representation.GetRepresentationId() == id {
			return representation
		}
	}
	t.Fatalf("representation %q not found", id)
	return nil
}
