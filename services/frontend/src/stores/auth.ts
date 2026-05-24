import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import * as authApi from '../api/auth';
import type { User } from '../types';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'));
  const user = ref<User | null>(null);

  const isAuthenticated = computed(() => !!token.value);

  function setToken(newToken: string | null) {
    token.value = newToken;
    if (newToken) {
      localStorage.setItem('auth_token', newToken);
    } else {
      localStorage.removeItem('auth_token');
    }
  }

  async function login(email: string, password: string) {
    const response = await authApi.login({ email, password });
    setToken(response.token);
  }

  async function register(name: string, email: string, password: string) {
    const response = await authApi.register({ name, email, password });
    return response;
  }

  function logout() {
    setToken(null);
    user.value = null;
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    logout,
  };
});