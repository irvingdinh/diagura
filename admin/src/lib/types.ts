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

export interface UserListParams {
  page?: number;
  per_page?: number;
  search?: string;
  role?: string;
  status?: "active" | "deleted";
}
