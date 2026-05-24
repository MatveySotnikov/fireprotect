<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Навигационная панель -->
    <nav class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <!-- Левая часть: логотип и ссылки -->
          <div class="flex items-center space-x-8">
            <router-link to="/" class="flex items-center gap-2 text-indigo-600 font-bold text-lg">
              <Shield class="h-6 w-6" />
              FireProtect
            </router-link>
            <router-link
                to="/"
                class="inline-flex items-center gap-1 px-1 pt-1 text-sm font-medium text-gray-700 hover:text-indigo-600 border-b-2 border-transparent hover:border-indigo-600 transition"
                exact-active-class="border-indigo-600 text-indigo-600"
                >
                <Calculator class="h-4 w-4" />
                Калькулятор
            </router-link>
            <router-link
                to="/history"
                class="inline-flex items-center gap-1 px-1 pt-1 text-sm font-medium text-gray-700 hover:text-indigo-600 border-b-2 border-transparent hover:border-indigo-600 transition"
                exact-active-class="border-indigo-600 text-indigo-600"
                >
                <History class="h-4 w-4" />
                История
            </router-link>
          </div>

          <!-- Правая часть: кнопка выхода -->
             <div class="flex items-center gap-4">
                <span class="text-sm text-gray-600">
                    {{ authStore.displayName || 'Пользователь' }}
                </span>
                <button
                    @click="handleLogout"
                    class="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 hover:text-red-600 transition"
                >
                    <LogOut class="h-4 w-4" />
                    Выход
                </button>
            </div>
        </div>
      </div>
    </nav>

    <!-- Основной контент страницы -->
    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { Shield, Calculator, History, LogOut } from '@lucide/vue';
import { useAuthStore } from '../../stores/auth';

const router = useRouter();
const authStore = useAuthStore();

function handleLogout() {
  authStore.logout();
  router.push('/login');
}
</script>