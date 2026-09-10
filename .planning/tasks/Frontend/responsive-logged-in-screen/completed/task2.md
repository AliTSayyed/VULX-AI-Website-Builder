# Task 2 — Service clients and the project hooks

**Goal.** Reach `ProjectService` and `MessageService` from the browser, and wrap the three project
RPCs in React Query. No component changes.

## 1. The two client hooks

Exactly the existing one-liner pattern (`useUserService.ts`, `useAccountService.ts`). Do not invent
a new shape.

`app/src/hooks/services/useProjectService.ts`:

```ts
import { ProjectService } from "@/gen/api/v1/project_service_pb";
import { useServiceClient } from "./useServiceClient";

export function useProjectService() {
  return useServiceClient(ProjectService);
}
```

`app/src/hooks/services/useMessageService.ts`: the same, with `MessageService` from
`@/gen/api/v1/message_service_pb`.

`useServiceClient` already sends `credentials: "include"` and points at `NEXT_PUBLIC_API_URL`
(falling back to `https://local.api.vulx.ai`), so the session cookie rides along. Nothing there
changes.

## 2. Create `app/src/hooks/useProjects.ts`

```ts
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

export function useProject(id: string | null) {
  const projects = useProjectService();

  return useQuery<Project | null>({
    queryKey: ["project", id],
    enabled: !!id,
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
```

### Points that matter

- **`limit` is a `bigint`.** The proto field is `int64`, which `protoc-gen-es` maps to `bigint` —
  `100` is a type error, `100n` is not. The same applies to `Profile.credits`, which the footer
  already stringifies.
- **`enabled: !!id`** on `useProject`. The Workspace renders in a draft state with no id (§3 of the
  doc); without the guard, React Query fires `GetProject({id: ""})` and gets a `NotFound`.
- **`["project", id]`** with the raw id in the key — task 3 invalidates it by that exact shape.
- **No `timeoutMs` anywhere**, here or on the transport. It would apply to every call including the
  multi-minute Build send in task 3.
- `retry: false` and `staleTime: 30_000` come from the provider defaults in `app/providers.tsx`.
  Do not repeat them.
- `CreateProject` takes `first_prompt` + `provider` only. It does **not** create a sandbox — that
  happens inside the first Build `SendMessage`. A freshly created project has `previewUrl === ""`.

## Done when

- `npm run lint` and `npm run build` pass.
- Nothing imports these hooks yet.
- Optional check: temporarily calling `useProjects()` from a client component shows one
  `ListProjects` request in the Network tab returning `{"projects":[...]}` while logged in. Remove
  the temporary call before committing.
