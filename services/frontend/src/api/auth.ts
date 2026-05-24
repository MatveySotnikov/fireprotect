import apiClient from './client';
import type { LoginResponse } from '../types';

interface RegisterPayload {
  name: string;
  email: string;
  password: string;
}

interface LoginPayload {
  email: string;
  password: string;
}

export async function register(payload: RegisterPayload) {
  const response = await apiClient.post('/auth/register', payload);
  return response.data;
}

export async function login(payload: LoginPayload): Promise<LoginResponse> {
  const response = await apiClient.post('/auth/login', payload);
  return response.data;
}