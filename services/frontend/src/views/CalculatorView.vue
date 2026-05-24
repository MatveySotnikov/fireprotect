<template>
  <div class="min-h-screen bg-gray-50 py-10">
    <div class="max-w-2xl mx-auto bg-white p-8 rounded-lg shadow">
      <h1 class="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
        <Calculator class="h-7 w-7 text-indigo-600" />
        Расчёт огнезащиты
      </h1>

      <form @submit.prevent="handleCalculate" class="space-y-6">
        <!-- Площадь -->
        <div>
          <label class="block text-sm font-medium text-gray-700">Площадь, м²</label>
          <input
            v-model.number="form.area"
            type="number"
            step="any"
            min="0.01"
            required
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>

        <!-- Тип площади -->
        <div>
          <label class="block text-sm font-medium text-gray-700">Тип площади</label>
          <div class="mt-2 space-y-2">
            <label class="inline-flex items-center mr-4">
              <input type="radio" v-model="form.area_type" value="projection" class="text-indigo-600" />
              <span class="ml-2 text-sm">Проекция</span>
            </label>
            <label class="inline-flex items-center">
              <input type="radio" v-model="form.area_type" value="slope" class="text-indigo-600" />
              <span class="ml-2 text-sm">Фактическая площадь ската</span>
            </label>
          </div>
        </div>

        <div v-if="form.area_type === 'projection'">
          <label class="block text-sm font-medium text-gray-700">Угол уклона, градусы (0–89)</label>
          <input
            v-model.number="form.slope_angle"
            type="number"
            step="any"
            min="0"
            max="90"
            required
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>

        <!-- Группа огнезащиты -->
        <div>
          <label class="block text-sm font-medium text-gray-700">Требуемая группа</label>
          <select
            v-model="form.target_group"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            <option value="1_group">I группа</option>
            <option value="2_group">II группа</option>
          </select>
        </div>

        <!-- Метод нанесения -->
        <div>
          <label class="block text-sm font-medium text-gray-700">Метод нанесения</label>
          <select
            v-model="form.application_method"
            class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            <option value="brush">Кисть</option>
            <option value="spray_indoor">Распыление (в помещении)</option>
            <option value="spray_outdoor">Распыление (на открытом воздухе)</option>
          </select>
        </div>

        <!-- Выбор материала -->
        <div>
          <label class="block text-sm font-medium text-gray-700">Огнезащитный состав</label>
            <select
                v-model="form.material_id"
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                >
                <option :value="undefined">Ручной ввод</option>
                <option
                    v-for="mat in materials"
                    :key="mat.ID"
                    :value="mat.ID"
                >
                    {{ mat.Title }}
                </option>
            </select>
          <!-- Подсказки по выбранному материалу -->
          <div v-if="selectedMaterial" class="mt-2 text-sm text-gray-600 bg-gray-50 p-2 rounded">
            <p>Плотность: <strong>{{ selectedMaterial.DefaultDensity }} кг/л</strong></p>
            <p v-if="form.target_group === '1_group'">
              Расход на I группу: <strong>{{ selectedMaterial.Group1Consumption }} кг/м²</strong>
            </p>
            <p v-else>
              Расход на II группу: <strong>{{ selectedMaterial.Group2Consumption }} кг/м²</strong>
            </p>
          </div>
        </div>

        <!-- Ручной ввод, если материал не выбран -->
        <template v-if="!form.material_id">
          <div>
            <label class="block text-sm font-medium text-gray-700">Нормативный расход, кг/м² (с учётом потерь)</label>
            <input
              v-model.number="form.normative_rate"
              type="number"
              step="any"
              min="0.01"
              required
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">Плотность состава, кг/л</label>
            <input
              v-model.number="form.density"
              type="number"
              step="any"
              min="0.01"
              required
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            />
          </div>
        </template>

        <!-- Ошибка -->
        <div v-if="errorMessage" class="text-red-600 text-sm bg-red-50 p-3 rounded-md">
          {{ errorMessage }}
        </div>

        <!-- Кнопка -->
        <button
          type="submit"
          :disabled="loading"
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          <Loader2 v-if="loading" class="animate-spin h-5 w-5 mr-2" />
          {{ loading ? 'Рассчитываем...' : 'Рассчитать' }}
        </button>
      </form>

      <!-- Результат -->
      <div v-if="result" class="mt-8 p-6 bg-green-50 border border-green-200 rounded-lg">
        <h3 class="text-lg font-semibold text-green-800 flex items-center gap-2">
          <CheckCircle class="h-5 w-5" />
          Результат расчёта
        </h3>
        <div class="mt-3 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="text-gray-600">Общая масса:</span>
            <p class="text-lg font-bold text-gray-900">{{ result.total_mass.toFixed(3) }} кг</p>
          </div>
          <div>
            <span class="text-gray-600">Общий объём:</span>
            <p class="text-lg font-bold text-gray-900">{{ result.total_volume.toFixed(3) }} л</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Calculator, Loader2, CheckCircle } from '@lucide/vue';
import * as materialsApi from '../api/materials';
import * as calcApi from '../api/calculations';
import type { Material, CalculationResponse } from '../types';

const materials = ref<Material[]>([]);
const loading = ref(false);
const errorMessage = ref('');
const result = ref<CalculationResponse | null>(null);

const form = reactive({
  area: undefined as number | undefined,
  area_type: 'projection' as 'projection' | 'slope',
  slope_angle: 0,
  target_group: '1_group' as '1_group' | '2_group',
  application_method: 'brush' as 'brush' | 'spray_indoor' | 'spray_outdoor',
  material_id: undefined as number | undefined,
  normative_rate: undefined as number | undefined,
  density: undefined as number | undefined,
});

const selectedMaterial = computed(() => {
  if (!form.material_id) return null;
  return materials.value.find(m => m.ID === form.material_id) || null;
});

onMounted(async () => {
  try {
    materials.value = await materialsApi.getMaterials();
  } catch (e: any) {
    errorMessage.value = 'Не удалось загрузить список материалов';
  }
});

async function handleCalculate() {
  errorMessage.value = '';
  result.value = null;
  if (!form.area || form.area <= 0) {
    errorMessage.value = 'Площадь должна быть больше нуля';
    return;
  }
  if (form.area_type === 'projection' && (form.slope_angle == null || form.slope_angle < 0 || form.slope_angle >= 90)) {
    errorMessage.value = 'Угол уклона должен быть от 0 до 89 градусов';
    return;
  }
  if (!form.material_id && (!form.normative_rate || !form.density)) {
    errorMessage.value = 'Заполните нормативный расход и плотность, если материал не выбран';
    return;
  }
  loading.value = true;
  try {
    const payload: any = {
      area: form.area,
      area_type: form.area_type,
      slope_angle: form.slope_angle,
      target_group: form.target_group,
      application_method: form.application_method,
    };
    if (form.material_id) {
      payload.material_id = form.material_id;
    } else {
      payload.normative_rate = form.normative_rate;
      payload.density = form.density;
    }
    result.value = await calcApi.createCalculation(payload);
  } catch (e: any) {
    if (e.response?.data) {
      const data = e.response.data;
      errorMessage.value = typeof data === 'string' ? data : (data.error || 'Ошибка расчёта');
    } else {
      errorMessage.value = 'Не удалось выполнить расчёт';
    }
  } finally {
    loading.value = false;
  }
}
</script>