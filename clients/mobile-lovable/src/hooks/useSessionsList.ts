import { useSessionsListContext } from "@/lib/sessions-list-provider";

export function useSessionsList() {
  return useSessionsListContext();
}
