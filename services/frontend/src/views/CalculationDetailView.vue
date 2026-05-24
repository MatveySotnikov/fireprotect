<template>
  <div class="min-h-screen bg-gray-50 py-10">
    <div class="max-w-2xl mx-auto bg-white p-8 rounded-lg shadow">
      <button @click="$router.back()" class="mb-4 text-indigo-600 hover:text-indigo-800 flex items-center gap-1">
        <ArrowLeft class="h-4 w-4" />
        Назад
      </button>

      <div v-if="loading" class="text-center py-10">
        <Loader2 class="animate-spin h-10 w-10 text-indigo-600 mx-auto" />
        <p class="mt-2 text-gray-600">Загрузка...</p>
      </div>

      <div v-else-if="errorMessage" class="text-red-600 bg-red-50 p-4 rounded-md">
        {{ errorMessage }}
      </div>

      <div v-else-if="calc">
        <h1 class="text-2xl font-bold text-gray-900 mb-6">Детали расчёта #{{ calc.id }}</h1>

        <dl class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
          <div>
            <dt class="text-gray-500">Площадь</dt>
            <dd class="font-medium">{{ calc.area }} м²</dd>
          </div>
          <div>
            <dt class="text-gray-500">Тип площади</dt>
            <dd class="font-medium">{{ calc.area_type === 'projection' ? 'Проекция' : 'Фактическая' }}</dd>
          </div>
          <div v-if="calc.area_type === 'projection'">
            <dt class="text-gray-500">Угол уклона</dt>
            <dd class="font-medium">{{ calc.slope_angle }}°</dd>
          </div>
          <div>
            <dt class="text-gray-500">Группа огнезащиты</dt>
            <dd class="font-medium">{{ calc.target_group === '1_group' ? 'I группа' : 'II группа' }}</dd>
          </div>
          <div>
            <dt class="text-gray-500">Метод нанесения</dt>
            <dd class="font-medium">
              {{ {
                brush: 'Кисть',
                spray_indoor: 'Распыление (в помещении)',
                spray_outdoor: 'Распыление (на открытом воздухе)'
              }[calc.application_method] }}
            </dd>
          </div>
          <div>
            <dt class="text-gray-500">Материал</dt>
            <dd class="font-medium">{{ calc.material?.Title || 'Ручной ввод' }}</dd>
          </div>
          <div>
            <dt class="text-gray-500">Расход (с учётом потерь)</dt>
            <dd class="font-medium">{{ calc.used_normative_rate }} кг/м²</dd>
          </div>
          <div>
            <dt class="text-gray-500">Плотность</dt>
            <dd class="font-medium">{{ calc.used_density }} кг/л</dd>
          </div>
          <div>
            <dt class="text-gray-500">Коэффициент потерь</dt>
            <dd class="font-medium">{{ calc.loss_factor }}</dd>
          </div>
          <div>
            <dt class="text-gray-500">Общая масса</dt>
            <dd class="font-bold text-lg text-green-700">{{ calc.total_mass.toFixed(3) }} кг</dd>
          </div>
          <div>
            <dt class="text-gray-500">Общий объём</dt>
            <dd class="font-bold text-lg text-green-700">{{ calc.total_volume.toFixed(3) }} л</dd>
          </div>
        </dl>

        <button
          @click="downloadAct"
          :disabled="downloading"
          class="mt-6 w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50"
        >
          <Download v-if="!downloading" class="h-5 w-5 mr-2" />
          <Loader2 v-else class="animate-spin h-5 w-5 mr-2" />
          {{ downloading ? 'Скачивание...' : 'Скачать акт (PDF)' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { ArrowLeft, Download, Loader2 } from '@lucide/vue';
import * as calculationsApi from '../api/calculations';
import type { CalculationRecord } from '../types';

const route = useRoute();
const calc = ref<CalculationRecord | null>(null);
const loading = ref(true);
const downloading = ref(false);
const errorMessage = ref('');

onMounted(async () => {
  const id = Number(route.params.id);
  if (isNaN(id)) {
    errorMessage.value = 'Неверный ID расчёта';
    loading.value = false;
    return;
  }
  try {
    calc.value = await calculationsApi.getCalculationById(id);
  } catch (e: any) {
    errorMessage.value = e.response?.data?.error || 'Не удалось загрузить расчёт';
  } finally {
    loading.value = false;
  }
});

async function downloadAct() {
  if (!calc.value) return;
  downloading.value = true;
  try {
    const blob = await calculationsApi.downloadAct(calc.value.id);
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `act_${calc.value.id}.pdf`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (e: any) {
    alert(e.response?.data?.error || 'Не удалось скачать акт');
  } finally {
    downloading.value = false;
  }
}
</script>