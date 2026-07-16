import type { RelayStatusInfo } from "@/stores/relayConnection";

import type { WebAcpRuntimeDeps } from "./webAcpRuntimeTypes";

interface PodSubscriptionConsumer {
  onStatus(status: RelayStatusInfo): void;
}

interface SharedPodSubscription {
  consumers: Set<PodSubscriptionConsumer>;
  dispose(): void;
  latestStatus: RelayStatusInfo | null;
  ready: Promise<void>;
}

const relaySubscriptions = new WeakMap<
  WebAcpRuntimeDeps["relay"],
  Map<string, SharedPodSubscription>
>();

export function retainWebAcpPodSubscription(
  deps: WebAcpRuntimeDeps,
  podKey: string,
  subscriptionId: string,
  consumer: PodSubscriptionConsumer,
): { ready: Promise<void>; release(): void } {
  const subscriptions = subscriptionsFor(deps);
  let shared = subscriptions.get(podKey);
  if (!shared) {
    shared = createSharedSubscription(deps, podKey, subscriptionId);
    subscriptions.set(podKey, shared);
  }
  shared.consumers.add(consumer);
  if (shared.latestStatus) consumer.onStatus(shared.latestStatus);
  let released = false;
  return {
    ready: shared.ready,
    release() {
      if (released) return;
      released = true;
      shared?.consumers.delete(consumer);
      if (shared?.consumers.size !== 0) return;
      shared?.dispose();
      if (subscriptions.get(podKey) === shared) subscriptions.delete(podKey);
    },
  };
}

function subscriptionsFor(
  deps: WebAcpRuntimeDeps,
): Map<string, SharedPodSubscription> {
  let subscriptions = relaySubscriptions.get(deps.relay);
  if (!subscriptions) {
    subscriptions = new Map();
    relaySubscriptions.set(deps.relay, subscriptions);
  }
  return subscriptions;
}

function createSharedSubscription(
  deps: WebAcpRuntimeDeps,
  podKey: string,
  subscriptionId: string,
): SharedPodSubscription {
  const consumers = new Set<PodSubscriptionConsumer>();
  const shared: SharedPodSubscription = {
    consumers,
    dispose: () => undefined,
    latestStatus: null,
    ready: Promise.resolve(),
  };
  const cleanups = [
    deps.relay.onAcpMessage(podKey, (messageType, payload) => {
      deps.dispatchRelayEvent(podKey, messageType, payload);
    }),
    deps.relay.onStatusChange(podKey, (status) => {
      shared.latestStatus = status;
      consumers.forEach((consumer) => consumer.onStatus(status));
    }),
  ];
  let disposed = false;
  shared.ready = deps.relay
    .subscribe(podKey, subscriptionId, () => undefined)
    .then(() => undefined);
  shared.dispose = () => {
      if (disposed) return;
      disposed = true;
      cleanups.forEach((cleanup) => cleanup());
      deps.relay.unsubscribe(podKey, subscriptionId);
  };
  return shared;
}
