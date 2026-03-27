export const queryKeys = {
  session: ["session"] as const,
  logs: {
    all: ["logs"] as const,
    list: (params: Record<string, unknown>) =>
      ["logs", "list", params] as const,
    dates: ["logs", "dates"] as const,
  },
  users: {
    all: ["users"] as const,
    list: (params: Record<string, unknown>) =>
      ["users", "list", params] as const,
    detail: (id: string) => ["users", "detail", id] as const,
  },
};
