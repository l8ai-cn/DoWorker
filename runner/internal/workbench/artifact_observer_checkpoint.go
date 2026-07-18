package workbench

func (o *ArtifactObserver) commitPendingState() error {
	if o.pendingState == nil {
		return nil
	}
	if err := writeArtifactObserverState(o.root, o.pendingState); err != nil {
		return err
	}
	o.pendingState = nil
	return nil
}

func (o *ArtifactObserver) snapshotState() *artifactObserverState {
	state := &artifactObserverState{
		SchemaVersion: artifactObserverStateSchema,
		Files:         make(map[string]artifactObserverFileState),
		Declarations:  make(map[string]artifactObserverDeclaredState),
	}
	for artifactPath, file := range o.baseline {
		state.Files[artifactPath] = persistedArtifactFile(file, 0, false)
	}
	for artifactPath, emitted := range o.emitted {
		state.Files[artifactPath] = persistedArtifactFile(
			emitted.file,
			emitted.revision,
			emitted.deleted,
		)
	}
	for id, artifact := range o.declaredBaseline {
		state.Declarations[id] = persistedDeclaredArtifact(artifact, false)
	}
	for id, emitted := range o.declaredEmitted {
		state.Declarations[id] = persistedDeclaredArtifact(emitted.artifact, true)
	}
	return state
}

func (o *ArtifactObserver) restoreState(state *artifactObserverState) {
	for artifactPath, persisted := range state.Files {
		file := artifactFile{
			path:      artifactPath,
			filename:  persisted.Filename,
			mediaType: persisted.MediaType,
			digest:    persisted.Digest,
			byteSize:  persisted.ByteSize,
		}
		if persisted.Revision == 0 {
			o.baseline[artifactPath] = file
			continue
		}
		o.emitted[artifactPath] = emittedArtifact{
			file:     file,
			revision: persisted.Revision,
			deleted:  persisted.Deleted,
		}
	}
	for id, persisted := range state.Declarations {
		artifact := declaredArtifact{
			artifactID:  id,
			revision:    persisted.Revision,
			producer:    persisted.Producer,
			fingerprint: persisted.Fingerprint,
		}
		if !persisted.Emitted {
			o.declaredBaseline[id] = artifact
			continue
		}
		o.declaredEmitted[id] = emittedDeclaredArtifact{
			artifact: artifact,
			revision: artifact.revision,
		}
	}
}

func persistedArtifactFile(
	file artifactFile,
	revision uint64,
	deleted bool,
) artifactObserverFileState {
	return artifactObserverFileState{
		Filename:  file.filename,
		MediaType: file.mediaType,
		Digest:    file.digest,
		ByteSize:  file.byteSize,
		Revision:  revision,
		Deleted:   deleted,
	}
}

func persistedDeclaredArtifact(
	artifact declaredArtifact,
	emitted bool,
) artifactObserverDeclaredState {
	return artifactObserverDeclaredState{
		Revision:    artifact.revision,
		Fingerprint: artifact.fingerprint,
		Producer:    artifact.producer,
		Emitted:     emitted,
	}
}
