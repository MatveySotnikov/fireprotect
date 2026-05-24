import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import * as authApi from '../api/auth';

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('auth_token'));
  const displayName = ref<string | null>(localStorage.getItem('display_name'));

  const isAuthenticated = computed(() => !!token.value);

  function setToken(newToken: string | null) {
    token.value = newToken;
    if (newToken) {
      localStorage.setItem('auth_token', newToken);
    } else {
      localStorage.removeItem('auth_token');
    }
  }

  function setDisplayName(name: string | null) {
    displayName.value = name;
    if (name) {
      localStorage.setItem('display_name', name);
    } else {
      localStorage.removeItem('display_name');
    }
  }

  async function login(email: string, password: string) {
    const response = await authApi.login({ email, password });
    setToken(response.token);
    setDisplayName(email); // используем email для отображения
  }

  async function register(name: string, email: string, password: string) {
    const response = await authApi.register({ name, email, password });
    setDisplayName(name);
    return response;
  }

  function logout() {
    setToken(null);
    setDisplayName(null);
  }

  return {
    token,
    displayName,
    isAuthenticated,
    login,
    register,
    logout,
  };
});