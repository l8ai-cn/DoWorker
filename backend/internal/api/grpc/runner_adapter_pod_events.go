package grpc

func (a *GRPCRunnerAdapter) SetPodEventSink(s PodEventSink) {
	a.podEvents = s
}

func (a *GRPCRunnerAdapter) SetWorkbenchEventSink(s WorkbenchEventSink) {
	a.workbenchEvents = s
}
