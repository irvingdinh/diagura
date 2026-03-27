import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";
import type { ApiListResponse, LogEntry, LogListParams } from "@/lib/types";

import { LogsTable } from "./logs-table";
import { LogsToolbar } from "./logs-toolbar";

export const LogsPage = () => {
  const [params, setParams] = useState<LogListParams>({
    page: 1,
    per_page: 20,
    date: new Date().toISOString().slice(0, 10),
  });

  const datesQuery = useQuery({
    queryKey: queryKeys.logs.dates,
    queryFn: () => api<{ data: string[] }>("/api/admin/logs/dates"),
    staleTime: 60 * 1000,
  });

  const query = useQuery({
    queryKey: queryKeys.logs.list(params as Record<string, unknown>),
    queryFn: () => {
      const qs = new URLSearchParams();
      if (params.date) qs.set("date", params.date);
      if (params.page) qs.set("page", String(params.page));
      if (params.per_page) qs.set("per_page", String(params.per_page));
      if (params.level) qs.set("level", params.level);
      if (params.search) qs.set("search", params.search);
      return api<ApiListResponse<LogEntry>>(`/api/admin/logs?${qs}`);
    },
    placeholderData: keepPreviousData,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Logs</h1>

      <LogsToolbar
        params={params}
        availableDates={datesQuery.data?.data ?? []}
        onChange={(next) => setParams({ ...next, page: 1 })}
      />

      <LogsTable data={query.data?.data ?? []} isLoading={query.isLoading} />

      {query.data?.meta && query.data.meta.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground text-sm">
            Showing{" "}
            {(query.data.meta.page - 1) * query.data.meta.per_page + 1}
            &ndash;
            {Math.min(
              query.data.meta.page * query.data.meta.per_page,
              query.data.meta.total,
            )}{" "}
            of {query.data.meta.total}
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={query.data.meta.page <= 1}
              onClick={() =>
                setParams((p) => ({ ...p, page: (p.page ?? 1) - 1 }))
              }
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={query.data.meta.page >= query.data.meta.total_pages}
              onClick={() =>
                setParams((p) => ({ ...p, page: (p.page ?? 1) + 1 }))
              }
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
