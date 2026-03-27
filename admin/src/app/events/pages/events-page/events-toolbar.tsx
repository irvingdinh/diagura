import { useEffect, useRef, useState } from "react";

import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { EventListParams } from "@/lib/types";

export const EventsToolbar = ({
  params,
  availableNames,
  onChange,
}: {
  params: EventListParams;
  availableNames: string[];
  onChange: (params: EventListParams) => void;
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
        value={params.name ?? "all"}
        onValueChange={(v) =>
          onChange({ ...params, name: v === "all" ? undefined : v })
        }
      >
        <SelectTrigger className="sm:w-52">
          <SelectValue placeholder="All Events" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Events</SelectItem>
          {availableNames.map((n) => (
            <SelectItem key={n} value={n}>
              {n}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={params.entity_type ?? "all"}
        onValueChange={(v) =>
          onChange({ ...params, entity_type: v === "all" ? undefined : v })
        }
      >
        <SelectTrigger className="sm:w-36">
          <SelectValue placeholder="All Entities" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Entities</SelectItem>
          <SelectItem value="user">User</SelectItem>
          <SelectItem value="session">Session</SelectItem>
        </SelectContent>
      </Select>
      <Input
        type="date"
        value={params.date_from ?? ""}
        onChange={(e) =>
          onChange({
            ...params,
            date_from: e.target.value || undefined,
          })
        }
        className="sm:w-40"
      />
      <Input
        type="date"
        value={params.date_to ?? ""}
        onChange={(e) =>
          onChange({
            ...params,
            date_to: e.target.value || undefined,
          })
        }
        className="sm:w-40"
      />
      <Input
        placeholder="Search data..."
        value={searchInput}
        onChange={(e) => handleSearch(e.target.value)}
        className="sm:max-w-xs"
      />
    </div>
  );
};
