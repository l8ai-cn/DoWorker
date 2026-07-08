package grpc

func (a *GRPCRunnerAdapter) SetPodEventSink(s PodEventSink) {
	a.podEvents = s
}
