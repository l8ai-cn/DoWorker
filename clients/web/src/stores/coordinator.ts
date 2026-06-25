import { create } from "zustand";
import {
  coordinatorApi,
  type CoordinatorExecution,
  type CoordinatorProject,
  type CreateProjectInput,
  type UpdateProjectInput,
} from "@/lib/api/coordinatorApi";

interface CoordinatorState {
  projects: CoordinatorProject[];
  executions: Record<number, CoordinatorExecution[]>;
  loading: boolean;
  error: string | null;

  loadProjects: () => Promise<void>;
  createProject: (data: CreateProjectInput) => Promise<void>;
  updateProject: (id: number, data: UpdateProjectInput) => Promise<void>;
  deleteProject: (id: number) => Promise<void>;
  loadExecutions: (id: number) => Promise<void>;
  runNow: (id: number) => Promise<void>;
}

function message(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

export const useCoordinatorStore = create<CoordinatorState>((set, get) => ({
  projects: [],
  executions: {},
  loading: false,
  error: null,

  loadProjects: async () => {
    set({ loading: true, error: null });
    try {
      set({ projects: await coordinatorApi.listProjects(), loading: false });
    } catch (e) {
      set({ error: message(e), loading: false });
    }
  },

  createProject: async (data) => {
    set({ error: null });
    try {
      await coordinatorApi.createProject(data);
      await get().loadProjects();
    } catch (e) {
      set({ error: message(e) });
      throw e;
    }
  },

  updateProject: async (id, data) => {
    try {
      const updated = await coordinatorApi.updateProject(id, data);
      set({ projects: get().projects.map((p) => (p.id === id ? updated : p)) });
    } catch (e) {
      set({ error: message(e) });
    }
  },

  deleteProject: async (id) => {
    try {
      await coordinatorApi.deleteProject(id);
      set({ projects: get().projects.filter((p) => p.id !== id) });
    } catch (e) {
      set({ error: message(e) });
    }
  },

  loadExecutions: async (id) => {
    try {
      const executions = await coordinatorApi.listExecutions(id);
      set({ executions: { ...get().executions, [id]: executions } });
    } catch (e) {
      set({ error: message(e) });
    }
  },

  runNow: async (id) => {
    set({ error: null });
    try {
      await coordinatorApi.runNow(id);
      await get().loadExecutions(id);
    } catch (e) {
      set({ error: message(e) });
    }
  },
}));
