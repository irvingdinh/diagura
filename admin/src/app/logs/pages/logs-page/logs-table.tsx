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
import type { LogEntry } from "@/lib/types";

function levelVariant(
  level: string,
): "default" | "secondary" | "destructive" | "outline" {
  switch (level?.toUpperCase()) {
    case "ERROR":
      return "destructive";
    case "WARN":
    case "WARNING":
      return "default";
    case "DEBUG":
      return "secondary";
    default:
      return "outline";
  }
}

function formatTime(timeStr: string): string {
  const d = new Date(timeStr);
  return d.toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    fractionalSecondDigits: 3,
  });
}

function formatSource(source?: {
  function: string;
  file: string;
  line: number;
}): string {
  if (!source) return "";
  const parts = source.file.split("/");
  const short = parts.slice(-2).join("/");
  return `${short}:${source.line}`;
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

export const LogsTable = ({
  data,
  isLoading,
}: {
  data: LogEntry[];
  isLoading: boolean;
}) => {
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);
  const [prevData, setPrevData] = useState(data);
  if (data !== prevData) {
    setPrevData(data);
    setExpandedIndex(null);
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-32">Time</TableHead>
            <TableHead className="w-20">Level</TableHead>
            <TableHead>Message</TableHead>
            <TableHead className="w-48">Source</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading &&
            Array.from({ length: 5 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell>
                  <Skeleton className="h-4 w-24" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-5 w-14" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-64" />
                </TableCell>
                <TableCell>
                  <Skeleton className="h-4 w-36" />
                </TableCell>
              </TableRow>
            ))}

          {!isLoading && data.length === 0 && (
            <TableRow>
              <TableCell
                colSpan={4}
                className="text-muted-foreground h-24 text-center"
              >
                No log entries found.
              </TableCell>
            </TableRow>
          )}

          {!isLoading &&
            data.map((entry, i) => (
              <Fragment key={i}>
                <TableRow
                  className="cursor-pointer"
                  onClick={() =>
                    setExpandedIndex(expandedIndex === i ? null : i)
                  }
                >
                  <TableCell className="font-mono text-sm">
                    {formatTime(entry.time)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={levelVariant(entry.level)}>
                      {entry.level}
                    </Badge>
                  </TableCell>
                  <TableCell className="max-w-md truncate">
                    {entry.msg}
                  </TableCell>
                  <TableCell className="text-muted-foreground truncate text-sm">
                    {formatSource(
                      entry.source as
                        | {
                            function: string;
                            file: string;
                            line: number;
                          }
                        | undefined,
                    )}
                  </TableCell>
                </TableRow>
                {expandedIndex === i && (
                  <TableRow>
                    <TableCell colSpan={4} className="bg-muted/50 p-4">
                      <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 font-mono text-sm">
                        {Object.entries(entry)
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
