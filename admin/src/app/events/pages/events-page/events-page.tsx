import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import { queryKeys } from "@/lib/query-keys";
import type { ApiListResponse, EventEntry, EventListParams } from "@/lib/types";

import { EventsTable } from "./events-table";
import { EventsToolbar } from "./events-toolbar";

export const EventsPage = () => {
  const [params, setParams] = useState<EventListParams>({
    page: 1,
    per_page: 20,
  });

  const namesQuery = useQuery({
    queryKey: queryKeys.events.names,
    queryFn: () => api<{ data: string[] }>("/api/admin/events/names"),
    staleTime: 60 * 1000,
  });

  const query = useQuery({
    queryKey: queryKeys.events.list(params as Record<string, unknown>),
    queryFn: () => {
      const qs = new URLSearchParams();
      if (params.page) qs.set("page", String(params.page));
      if (params.per_page) qs.set("per_page", String(params.per_page));
      if (params.name) qs.set("name", params.name);
      if (params.entity_type) qs.set("entity_type", params.entity_type);
      if (params.search) qs.set("search", params.search);
      if (params.date_from) qs.set("date_from", params.date_from);
      if (params.date_to) qs.set("date_to", params.date_to);
      return api<ApiListResponse<EventEntry>>(`/api/admin/events?${qs}`);
    },
    placeholderData: keepPreviousData,
  });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Events</h1>

      <EventsToolbar
        params={params}
        availableNames={namesQuery.data?.data ?? []}
        onChange={(next) => setParams({ ...next, page: 1 })}
      />

      <EventsTable data={query.data?.data ?? []} isLoading={query.isLoading} />

      {query.data?.meta && query.data.meta.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground text-sm">
            Showing {(query.data.meta.page - 1) * query.data.meta.per_page + 1}
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
