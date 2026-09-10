"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { Project } from "@/gen/api/v1/project_service_pb";
import type { AiProvider } from "@/gen/api/v1/enums_pb";
import { useProjectService } from "./services/useProjectService";

/** One page is the whole list at demo scale — see the doc, §8. */
const PAGE = 100n;

export function useProjects() {
  const projects = useProjectService();

  return useQuery<Project[]>({
    queryKey: ["projects"],
    queryFn: async () => {
      const res = await projects.listProjects({ limit: PAGE });
      return res.projects;
    },
  });
}

export function useProject(id: string | null, opts?: { pollForTitle?: boolean }) {
  const projects = useProjectService();

  return useQuery<Project | null>({
    queryKey: ["project", id],
    enabled: !!id,
    // Only while a just-created project's title might still be generating —
    // see LoggedInScreen's pendingTitleId. Never on for an already-open project.
    refetchInterval: opts?.pollForTitle ? 2000 : false,
    queryFn: async () => {
      const res = await projects.getProject({ id: id! });
      return res.project ?? null;
    },
  });
}

export function useCreateProject() {
  const projects = useProjectService();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { firstPrompt: string; provider: AiProvider }) => {
      const res = await projects.createProject(input);
      if (!res.project) throw new Error("CreateProject returned no project");
      return res.project;
    },
    onSuccess: (project) => {
      // Seed the detail cache so the Workspace can read the project without a
      // second round trip, then let the list refetch in the background.
      queryClient.setQueryData(["project", project.id], project);
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}
