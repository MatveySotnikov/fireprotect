import apiClient from './client';
import type { Material } from '../types';

export async function getMaterials(): Promise<Material[]> {
  const response = await apiClient.get('/materials');
  return response.data;
}