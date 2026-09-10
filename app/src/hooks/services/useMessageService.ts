import { MessageService } from "@/gen/api/v1/message_service_pb";
import { useServiceClient } from "./useServiceClient";

export function useMessageService() {
  return useServiceClient(MessageService);
}
