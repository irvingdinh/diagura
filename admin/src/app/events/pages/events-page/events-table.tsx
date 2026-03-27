import { Fragment, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { EventEntry } from "@/lib/types";

function eventVariant(
  name: string,
): "default" | "secondary" | "destructive" | "outline" {
  if (name.includes("deleted") || name.includes("invalidated")) {
    return "destructive";
  }
  if (name.includes("created") || name.includes("login")) {
    return "default";
  }
  if (name.includes("updated") || name.includes("restored")) {
    return "secondary";
  }
  return "outline";
}

function formatTime(timeStr: string): string {
  const d = new Date(timeStr);
  return d.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function tryParseJSON(str: string): Record<string, unknown> | null {
  try {
    const parsed = JSON.parse(str);
    return typeof parsed === "object" && parsed !== null ? parsed : null;
  } catch {
    return null;
  }
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

export const EventsTable = ({
  data,
  isLoading,
}: {
  data: EventEntry[];
  isLoading: boolean;
}) => {
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-36">Time</TableHead>
            <TableHead className="w-44">Event</TableHead>
            <TableHead className="w-32">Actor</TableHead>
            <TableHead className="w-36">Entity</TableHead>
            <TableHead>Data</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading &&
            Array.from({ length: 5 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell>
                  <Skeleton className="h-4 w-28" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-5 w-32" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-24" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-24" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-48" />
                </TableCell>
              </TableRow>
            ))}

          {!isLoading && data.length === 0 && (
            <TableRow>
              <TableCell
                colSpan={5}
                className="text-muted-foreground h-24 text-center"
              >
                No events found.
              </TableCell>
            </TableRow>
          )}

          {!isLoading &&
            data.map((entry, i) => (
              <Fragment key={entry.id}>
                <TableRow
                  className="cursor-pointer"
                  onClick={() =>
                    setExpandedIndex(expandedIndex === i ? null : i)
                  }
                >
                  <TableCell className="font-mono text-sm">
                    {formatTime(entry.created_at)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={eventVariant(entry.name)}>
                      {entry.name}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground truncate font-mono text-sm">
                    {entry.actor_id
                      ? entry.actor_id.slice(0, 8) + "..."
                      : "system"}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {entry.entity_type && entry.entity_id
                      ? `${entry.entity_type}:${entry.entity_id.slice(0, 8)}...`
                      : ""}
                  </TableCell>
                  <TableCell className="max-w-xs truncate font-mono text-sm">
                    {entry.data}
                  </TableCell>
                </TableRow>
                {expandedIndex === i && (
                  <TableRow>
                    <TableCell colSpan={5} className="bg-muted/50 p-4">
                      <div className="space-y-3">
                        <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
                          <span className="text-muted-foreground font-medium">
                            ID
                          </span>
                          <span className="font-mono">{entry.id}</span>
                          <span className="text-muted-foreground font-medium">
                            Actor ID
                          </span>
                          <span className="font-mono">
                            {entry.actor_id || "—"}
                          </span>
                          <span className="text-muted-foreground font-medium">
                            Request ID
                          </span>
                          <span className="font-mono">
                            {entry.request_id || "—"}
                          </span>
                          <span className="text-muted-foreground font-medium">
                            IP
                          </span>
                          <span className="font-mono">{entry.ip || "—"}</span>
                          {entry.entity_type && (
                            <>
                              <span className="text-muted-foreground font-medium">
                                Entity
                              </span>
                              <span className="font-mono">
                                {entry.entity_type}:{entry.entity_id}
                              </span>
                            </>
                          )}
                        </div>
                        {(() => {
                          const parsed = tryParseJSON(entry.data);
                          if (!parsed) return null;
                          return (
                            <div className="border-t pt-3">
                              <p className="text-muted-foreground mb-1 text-xs font-medium uppercase">
                                Event Data
                              </p>
                              <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 font-mono text-sm">
                                {Object.entries(parsed)
                                  .sort(([a], [b]) => a.localeCompare(b))
                                  .map(([key, value]) => (
                                    <Fragment key={key}>
                                      <span className="text-muted-foreground font-medium">
                                        {key}
                                      </span>
                                      <span className="break-all">
                                        {formatValue(value)}
                                      </span>
                                    </Fragment>
                                  ))}
                              </div>
                            </div>
                          );
                        })()}
                      </div>
                    </TableCell>
                  </TableRow>
                )}
              </Fragment>
            ))}
        </TableBody>
      </Table>
    </div>
  );
};
