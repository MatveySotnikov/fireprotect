<template>
  <div class="min-h-screen bg-gray-50 py-10">
    <div class="max-w-4xl mx-auto bg-white p-8 rounded-lg shadow">
      <h1 class="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
        <History class="h-7 w-7 text-indigo-600" />
        История расчётов
      </h1>

      <!-- Индикатор загрузки -->
      <div v-if="loading" class="text-center py-10">
        <Loader2 class="animate-spin h-10 w-10 text-indigo-600 mx-auto" />
        <p class="mt-2 text-gray-600">Загрузка...</p>
      </div>

      <!-- Ошибка -->
      <div v-else-if="errorMessage" class="text-red-600 bg-red-50 p-4 rounded-md">
        {{ errorMessage }}
      </div>

      <!-- Пустой список -->
      <div v-else-if="calculations.length === 0" class="text-center py-10 text-gray-500">
        <p>Нет сохранённых расчётов</p>
      </div>

      <!-- Таблица с расчётами -->
      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Площадь</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Материал</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Масса, кг</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Объём, л</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Действия</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="calc in calculations" :key="calc.id" class="hover:bg-gray-50">
              <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ calc.id }}</td>
              <td class="px-4 py-3 text-sm text-gray-700">{{ calc.area }} м²</td>
              <td class="px-4 py-3 text-sm text-gray-700">
                {{ calc.material?.Title || 'Ручной ввод' }}
              </td>
              <td class="px-4 py-3 text-sm text-gray-700">{{ calc.total_mass.toFixed(2) }}</td>
              <td class="px-4 py-3 text-sm text-gray-700">{{ calc.total_volume.toFixed(2) }}</td>
              <td class="px-4 py-3 text-sm space-x-2">
                <router-link
                  :to="`/calculations/${calc.id}`"
                  class="text-indigo-600 hover:text-indigo-800 font-medium"
                >
                  Детали
                </router-link>
                <button
                  @click="downloadAct(calc.id)"
                  class="text-green-600 hover:text-green-800 font-medium"
                >
                  Скачать
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { History, Loader2 } from '@lucide/vue';
import * as calculationsApi from '../api/calculations';
import type { CalculationRecord } from '../types';

const calculations = ref<CalculationRecord[]>([]);
const loading = ref(true);
const errorMessage = ref('');

onMounted(async () => {
  try {
    calculations.value = await calculationsApi.getCalculations();
  } catch (e: any) {
    errorMessage.value = e.response?.data?.error || 'Не удалось загрузить историю';
  } finally {
    loading.value = false;
  }
});

async function downloadAct(id: number) {
  try {
    const blob = await calculationsApi.downloadAct(id);
    // Создаём временную ссылку для скачивания
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `act_${id}.pdf`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (e: any) {
    alert(e.response?.data?.error || 'Не удалось скачать акт');
  }
}
</script>