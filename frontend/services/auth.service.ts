import { apiClient } from "@/lib/api-client";
import type { User } from "@/types/api";

export interface AuthResponse {
  user: User;
  access_token_expires_at: string;
}

export const authService = {
  login: (email: string, password: string) =>
    apiClient.post<AuthResponse>("/api/v1/auth/login", { email, password }),
  register: (name: string, email: string, password: string) =>
    apiClient.post<User>("/api/v1/auth/register", { name, email, password }),
  logout: () => apiClient.post<void>("/api/v1/auth/logout"),
  refresh: () => apiClient.post<AuthResponse>("/api/v1/auth/refresh"),
  me: () => apiClient.get<User>("/api/v1/me"),
};
