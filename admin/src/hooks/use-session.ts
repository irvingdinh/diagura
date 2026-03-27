import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";
import type { ApiResponse, SessionUser } from "@/lib/types";

export function useSession() {
  return useQuery({
    queryKey: queryKeys.session,
    queryFn: () =>
      api<ApiResponse<SessionUser>>("/api/auth/session").then((r) => r.data),
    retry: false,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: true,
  });
}
