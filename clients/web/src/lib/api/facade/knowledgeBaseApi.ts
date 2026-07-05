// Facade for the knowledgebase Connect-RPC adapter. Business code imports
// from here so the wire-shape layer stays internal to the facade boundary.
// Tests mock this path.

export interface KnowledgeBase {
  id: number;
  slug: string;
  name: string;
  description: string;
  http_clone_url: string;
  default_branch: string;
  source_type: string;
  sync_status: string;
  sync_error?: string;
  last_synced_at?: string;
  created_at: string;
  updated_at: string;
}

export interface KbAgentMount {
  agent_slug: string;
  mode: "ro" | "rw";
}

// Pod-creation-time mount selection (emitted as the Agentfile
// `KNOWLEDGE slug [rw]` declaration; the backend resolves slugs to git
// clone specs at orchestration time).
export interface KnowledgeMountSelection {
  slug: string;
  mode: "ro" | "rw";
}

export interface KbDirEntry {
  name: string;
  path: string;
  type: "file" | "dir";
  size: number;
}

export interface KbFile {
  path: string;
  content: string;
  size: number;
}

export {
  listKnowledgeBases,
  getKnowledgeBase,
  createKnowledgeBase,
  updateKnowledgeBase,
  deleteKnowledgeBase,
  listKbAgentMounts,
  setKbAgentMounts,
  getKbFile,
  listKbDir,
} from "../connect/knowledgeBaseConnect";
