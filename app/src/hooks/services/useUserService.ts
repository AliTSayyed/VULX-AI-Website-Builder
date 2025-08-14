import { UserService } from "@/gen/api/v1/user_service_pb";
import { useServiceClient } from "./useServiceClient";

export function useUserService() {
  return useServiceClient(UserService);
}
