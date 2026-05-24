<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
      <div class="text-center mb-6">
        <UserPlus class="mx-auto h-12 w-12 text-indigo-600" />
        <h2 class="mt-4 text-3xl font-bold text-gray-900">Регистрация</h2>
        <p class="mt-2 text-sm text-gray-600">
          Уже есть аккаунт?
          <router-link to="/login" class="font-medium text-indigo-600 hover:text-indigo-500">
            Войти
          </router-link>
        </p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-5">
        <div>
          <label for="name" class="block text-sm font-medium text-gray-700">Имя</label>
          <input
            id="name"
            v-model="name"
            type="text"
            required
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="Ваше имя"
          />
        </div>

        <div>
          <label for="email" class="block text-sm font-medium text-gray-700">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="user@example.com"
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700">Пароль</label>
          <div class="relative mt-1">
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              required
              class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Минимум 6 символов"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
            >
              <Eye v-if="!showPassword" class="h-5 w-5" />
              <EyeOff v-else class="h-5 w-5" />
            </button>
          </div>
        </div>

        <div v-if="errorMessage" class="text-red-600 text-sm bg-red-50 p-3 rounded-md">
          {{ errorMessage }}
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          <Loader2 v-if="loading" class="animate-spin h-5 w-5 mr-2" />
          {{ loading ? 'Регистрация...' : 'Зарегистрироваться' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { UserPlus, Eye, EyeOff, Loader2 } from '@lucide/vue';
import { useAuthStore } from '../stores/auth';

const router = useRouter();
const authStore = useAuthStore();

const name = ref('');
const email = ref('');
const password = ref('');
const showPassword = ref(false);
const loading = ref(false);
const errorMessage = ref('');

async function handleRegister() {
  errorMessage.value = '';
  loading.value = true;
  try {
    await authStore.register(name.value, email.value, password.value);
    router.push('/login');
  } catch (error: any) {
    if (error.response && error.response.data) {
      const data = error.response.data;
      errorMessage.value = typeof data === 'string' ? data : (data.error || 'Ошибка регистрации');
    } else {
      errorMessage.value = 'Не удалось соединиться с сервером';
    }
  } finally {
    loading.value = false;
  }
}
</script>