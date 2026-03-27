import { useEffect, useRef, useState } from "react";

import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { LogListParams } from "@/lib/types";

function formatDateLabel(dateStr: string): string {
  const [year, month, day] = dateStr.split("-").map(Number);
  return new Date(year, month - 1, day).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export const LogsToolbar = ({
  params,
  availableDates,
  onChange,
}: {
  params: LogListParams;
  availableDates: string[];
  onChange: (params: LogListParams) => void;
}) => {
  const [searchInput, setSearchInput] = useState(params.search ?? "");
  const timerRef = useRef<ReturnType<typeof setTimeout>>(null);
  const paramsRef = useRef(params);
  useEffect(() => {
    paramsRef.current = params;
  });

  const handleSearch = (value: string) => {
    setSearchInput(value);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      onChange({ ...paramsRef.current, search: value || undefined });
    }, 300);
  };

  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
      <Select
        value={params.date ?? ""}
        onValueChange={(v) => onChange({ ...params, date: v })}
      >
        <SelectTrigger className="sm:w-48">
          <SelectValue placeholder="Select date" />
        </SelectTrigger>
        <SelectContent>
          {availableDates.map((d) => (
            <SelectItem key={d} value={d}>
              {formatDateLabel(d)}
            </SelectItem>
          ))}
          {availableDates.length === 0 && (
            <SelectItem value={params.date ?? ""} disabled>
              No logs available
            </SelectItem>
          )}
        </SelectContent>
      </Select>
      <Select
        value={params.level ?? "all"}
        onValueChange={(v) =>
          onChange({ ...params, level: v === "all" ? undefined : v })
        }
      >
        <SelectTrigger className="sm:w-36">
          <SelectValue placeholder="All Levels" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Levels</SelectItem>
          <SelectItem value="DEBUG">DEBUG+</SelectItem>
          <SelectItem value="INFO">INFO+</SelectItem>
          <SelectItem value="WARN">WARN+</SelectItem>
          <SelectItem value="ERROR">ERROR</SelectItem>
        </SelectContent>
      </Select>
      <Input
        placeholder="Search logs..."
        value={searchInput}
        onChange={(e) => handleSearch(e.target.value)}
        className="sm:max-w-xs"
      />
    </div>
  );
};
