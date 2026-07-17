"use client";

import dynamic from "next/dynamic";
import { useSyncExternalStore } from "react";
import { ServiceWorkerRegistration } from "./ServiceWorkerRegistration";

const PushNotificationManager = dynamic(
  () =>
    import("./PushNotificationManager").then(
      ({ PushNotificationManager: Manager }) => Manager,
    ),
  { ssr: false },
);

interface PWAProviderProps {
  children: React.ReactNode;
}

function subscribe() {
  return () => {};
}
function getSnapshot() {
  return true;
}
function getServerSnapshot() {
  return false;
}
function useIsMounted() {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}

export function PWAProvider({ children }: PWAProviderProps) {
  const mounted = useIsMounted();

  // Don't render PWA components during SSR
  if (!mounted) {
    return <>{children}</>;
  }

  return (
    <>
      <ServiceWorkerRegistration />
      <PushNotificationManager autoSubscribe={false}>
        {children}
      </PushNotificationManager>
    </>
  );
}

export default PWAProvider;
