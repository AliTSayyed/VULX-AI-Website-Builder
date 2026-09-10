import { ProjectService } from "@/gen/api/v1/project_service_pb";
import { useServiceClient } from "./useServiceClient";

export function useProjectService() {
  return useServiceClient(ProjectService);
}
