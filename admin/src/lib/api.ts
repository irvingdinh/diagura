export async function api<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });

  if (!res.ok) {
    if (res.status === 401) {
      window.location.href = "/admin/login";
      throw new Error("Unauthorized");
    }
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Something went wrong");
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}

export async function logout() {
  await fetch("/api/auth/logout", { method: "POST" }).catch(() => {});
  window.location.href = "/admin/login";
}
