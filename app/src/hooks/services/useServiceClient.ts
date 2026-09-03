/*
 * This file sets up the rpc connection to the backend.
 * The base URL comes from NEXT_PUBLIC_API_URL, inlined at build time.
 */

import { useMemo } from "react";
import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { GenService, GenServiceMethods } from "@bufbuild/protobuf/codegenv2";

const LOCAL_API_URL = "https://local.api.vulx.ai";

const getBaseUrl = (): string => {
  return process.env.NEXT_PUBLIC_API_URL ?? LOCAL_API_URL;
};

// binary protobuf is faster on the wire; staying on JSON against the local
// API so requests are readable in the Network tab when debugging
const binaryFormat = getBaseUrl() === LOCAL_API_URL ? false : true;

export function useServiceClient<T extends GenService<GenServiceMethods>>(
  service: T,
): Client<T> {
  const transport = useMemo(
    () =>
      createConnectTransport({
        baseUrl: getBaseUrl(),
        useBinaryFormat: binaryFormat,
        fetch: (input, init) =>
          fetch(input, { ...init, credentials: "include" }),
      }),
    [],
  );
  return useMemo(() => createClient(service, transport), [service, transport]);
}
