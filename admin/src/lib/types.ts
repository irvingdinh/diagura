export interface SessionUser {
  id: string;
  email: string;
  name: string;
  role: string;
  force_password_change: boolean;
}

export interface User {
  id: string;
  email: string;
  name: string;
  force_password_change: boolean;
  deleted_at?: string | null;
  created_at: string;
  updated_at: string;
  role_slug: string;
  role_name: string;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface ApiResponse<T> {
  data: T;
}

export interface ApiListResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export interface LogEntry {
  time: string;
  level: string;
  msg: string;
  source?: { function: string; file: string; line: number };
  [key: string]: unknown;
}

export interface LogListParams {
  date?: string;
  page?: number;
  per_page?: number;
  level?: string;
  search?: string;
}

export interface UserListParams {
  page?: number;
  per_page?: number;
  search?: string;
  role?: string;
  status?: "active" | "deleted";
}

export interface EventEntry {
  id: string;
  name: string;
  actor_id: string;
  request_id: string;
  ip: string;
  entity_type: string;
  entity_id: string;
  data: string;
  created_at: string;
}

export interface EventListParams {
  page?: number;
  per_page?: number;
  name?: string;
  entity_type?: string;
  search?: string;
  date_from?: string;
  date_to?: string;
}
