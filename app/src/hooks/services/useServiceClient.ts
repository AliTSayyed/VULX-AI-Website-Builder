/*
 * This file sets up the rpc connection to the backend
 * will need to change dev / prod urls when defined url is made
 */

import { useMemo } from "react";
import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { GenService } from "@bufbuild/protobuf/codegenv2";

const getBaseUrl = (): string => {
  return "http://localhost:8080";
};

const binaryFormat = getBaseUrl() === "http://localhost:8080" ? false : true;

export function useServiceClient<T extends GenService<any>>(
  service: T
): Client<T> {
  const transport = useMemo(
    () =>
      createConnectTransport({
        baseUrl: getBaseUrl(),
        useBinaryFormat: binaryFormat,
      }),
    []
  );
  return useMemo(() => createClient(service, transport), [service, transport]);
}
