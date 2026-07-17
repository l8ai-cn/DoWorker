export interface AgentArtifactItem {
  id: string;
  kind: "artifact";
  artifactId: string;
  filename: string;
  mimeType: string | null;
  status: "completed" | "failed";
}
