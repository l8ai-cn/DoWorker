# Worker Component Entrypoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** expose the production Web User Worker as a same-root React component, a self-routed React component, and an imperative web-page mount API without duplicating the Worker UI.

**Architecture:** `AgentCloudApp` remains the same-root integration point and accepts a router supplied by its host. A thin `AgentCloudStandaloneApp` adds `HashRouter` for standalone documents, while `mountAgentCloudApp` owns only `createRoot` and unmount lifecycle for non-React hosts. `worker.html` and `iframe.html` call the same standalone component.

**Tech Stack:** React 19, React Router, Vite, Vitest, Testing Library.

---

## File Structure

| Path | Responsibility |
| --- | --- |
| `clients/web-user/src/standalone.tsx` | Self-routed Worker component for document and isolated React use |
| `clients/web-user/src/mount.tsx` | Imperative mount/unmount entry for non-React hosts |
| `clients/web-user/src/worker.tsx` | Document bootstrap using the self-routed component |
| `clients/web-user/src/embed.tsx` | Re-export the public component entrypoints |
| `clients/web-user/src/standalone.test.tsx` | Router ownership behavior |
| `clients/web-user/src/mount.test.tsx` | Imperative lifecycle behavior |

### Task 1: Self-routed React entry

**Files:**
- Create: `clients/web-user/src/standalone.test.tsx`
- Create: `clients/web-user/src/standalone.tsx`

- [x] **Step 1: Write the failing test**

```tsx
vi.mock("./embed", () => ({
  AgentCloudApp: ({ basename }: { basename?: string }) => (
    <output data-testid="worker-app">{basename ?? "root"}</output>
  ),
}));

it("provides a hash router and forwards AgentCloudApp props", () => {
  render(<AgentCloudStandaloneApp basename="/worker" isDarkMode />);

  expect(screen.getByTestId("worker-app")).toHaveTextContent("/worker");
});
```

- [x] **Step 2: Run test to verify it fails**

Run: `cd clients/web-user && pnpm exec vitest run src/standalone.test.tsx`

Expected: FAIL because `AgentCloudStandaloneApp` does not exist.

- [x] **Step 3: Write minimal implementation**

```tsx
export function AgentCloudStandaloneApp(props: AgentCloudAppProps) {
  return (
    <HashRouter>
      <AgentCloudApp {...props} />
    </HashRouter>
  );
}
```

- [x] **Step 4: Run test to verify it passes**

Run: `cd clients/web-user && pnpm exec vitest run src/standalone.test.tsx`

Expected: PASS.

### Task 2: Imperative page mount API

**Files:**
- Create: `clients/web-user/src/mount.test.tsx`
- Create: `clients/web-user/src/mount.tsx`

- [x] **Step 1: Write the failing test**

```tsx
it("mounts the standalone Worker and removes it on unmount", () => {
  const element = document.createElement("div");
  const mounted = mountAgentCloudApp(element);

  expect(element.querySelector("[data-testid=worker-app]")).not.toBeNull();

  mounted.unmount();
  expect(element.innerHTML).toBe("");
});
```

- [x] **Step 2: Run test to verify it fails**

Run: `cd clients/web-user && pnpm exec vitest run src/mount.test.tsx`

Expected: FAIL because `mountAgentCloudApp` does not exist.

- [x] **Step 3: Write minimal implementation**

```tsx
export function mountAgentCloudApp(element: Element, props: AgentCloudAppProps = {}) {
  const root = createRoot(element);
  root.render(<AgentCloudStandaloneApp {...props} />);
  return { unmount: () => root.unmount() };
}
```

- [x] **Step 4: Run test to verify it passes**

Run: `cd clients/web-user && pnpm exec vitest run src/mount.test.tsx`

Expected: PASS.

### Task 3: Route document entrypoints through the public component

**Files:**
- Modify: `clients/web-user/src/worker.tsx`
- Modify: `clients/web-user/src/embed.tsx`

- [x] **Step 1: Replace the document-local `HashRouter` composition**

```tsx
createRoot(rootElement).render(
  <StrictMode>
    <AgentCloudStandaloneApp />
  </StrictMode>,
);
```

- [x] **Step 2: Re-export the self-routed and imperative entrypoints**

```ts
export { AgentCloudStandaloneApp } from "./standalone";
export { mountAgentCloudApp } from "./mount";
```

- [x] **Step 3: Run focused tests and the Vite build**

Run:

```bash
cd clients/web-user
pnpm exec vitest run src/standalone.test.tsx src/mount.test.tsx
pnpm exec vite build
```

Expected: all tests pass and the build emits `dist/worker.html` and `dist/iframe.html`.

### Task 4: Browser verification

**Files:**
- No source changes.

- [x] **Step 1: Open `worker.html` and `iframe.html`**

Verify both paths render the Agent selector, execution target picker, working directory control, prompt input, and disabled/enabled Start session state.

- [x] **Step 2: Verify an imperative host**

Use the rendered test fixture that calls `mountAgentCloudApp`; confirm `unmount()` removes the root.

## Review Checklist

- [x] The public React same-root path never mounts a nested router.
- [x] The standalone path owns exactly one `HashRouter`.
- [x] The imperative API adds no global state and always returns an unmount handle.
- [x] No generic transcript renderer or transport fallback is reintroduced.
- [x] New files remain under 200 lines.
