import { useEffect, useState } from "react";
import { readAuthToken } from "@/lib/auth-store";

export function useIsAuthed(): boolean {
  const [authed, setAuthed] = useState(false);
  useEffect(() => {
    setAuthed(Boolean(readAuthToken()));
  }, []);
  return authed;
}
