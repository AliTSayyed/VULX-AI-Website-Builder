import { AccountService } from "@/gen/api/v1/account_service_pb";
import { useServiceClient } from "./useServiceClient";

export function useAccountService() {
  return useServiceClient(AccountService);
}
