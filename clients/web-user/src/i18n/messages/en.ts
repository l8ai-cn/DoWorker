const englishMessages = {
  brand: "Agent Cloud",
  shell: {
    newSession: "New session",
    searchSessions: "Search sessions",
    sessions: "Sessions",
    settings: "Settings",
    backToApp: "Back to Agent Cloud",
  },
  composer: {
    heading: "What should we do?",
    signIn: "Sign in",
    signingIn: "Signing in…",
    username: "Username",
    password: "Password",
    chooseHostWorkspace: "Please choose a host and working directory",
    enterMessage: "Enter a message to get started",
    workingDirectory: "Working directory",
    noWorktree: "No worktree",
    loading: "Loading…",
    send: "Send",
    selectHost: "Select host",
    noHosts: "No hosts",
    connecting: "Connecting…",
    repository: "Repository",
    selectAgent: "Select agent",
    filesystemError: "Failed to load directory",
    emptyDirectory: "(empty directory)",
    noMatchingEntries: "No matching entries",
    filesystemTimeout: "Directory listing timed out — type a path directly (e.g. /workspace)",
    placeholder: "Describe a task to start a new session…",
    placeholderSkills: "Describe a task, or try a skill",
  },
  auth: {
    welcome: "Welcome to Agent Cloud.",
    devAccounts: "Local dev accounts from deploy/dev/seed/seed.sql",
    devUser: "Dev user",
    admin: "Admin",
    use: "Use",
  },
} as const;

type StringCatalog<T> = {
  [K in keyof T]: T[K] extends string ? string : StringCatalog<T[K]>;
};

export type MessageTree = StringCatalog<typeof englishMessages>;
export const en: MessageTree = englishMessages;
