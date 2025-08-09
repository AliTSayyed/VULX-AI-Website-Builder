/*
 * This file sets up the rpc connection to the backend
 * will need to change dev / prod urls when defined url is made
 */

import { useMemo } from "react";
import { createClient, type Client } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { GenService } from "@bufbuild/protobuf/codegenv2";

// const getBaseUrl = (): string => {
//   let url = "";
//   if (typeof window !== "undefined") {
//     const hostname = window.location.hostname;
//     if (hostname === "vulx.ai") {
//       url = "https://www.vulx.ai";
//     }
//     const match = hostname.match(/^([^.]+)\.vulx\.ai$/);
//     if (match) {
//       const env = match[1];
//       url = `https://${env}.api.getbrain.ai`;
//     }
//   } else {
//     url = "http://localhost:8080";
//   }
//   return url;
// };

export function useServiceClient<T extends GenService<any>>(
  service: T
): Client<T> {
  const transport = useMemo(
    () =>
      createConnectTransport({
        baseUrl: "http://localhost:8080",
        useBinaryFormat: false,
      }),
    []
  );
  return useMemo(() => createClient(service, transport), [service, transport]);
}
