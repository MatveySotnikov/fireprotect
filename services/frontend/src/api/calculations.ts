import apiClient from './client';
import type { CalculationRequest, CalculationResponse, CalculationRecord } from '../types';

export async function createCalculation(payload: CalculationRequest): Promise<CalculationResponse> {
  const response = await apiClient.post('/calc', payload);
  return response.data;
}

export async function getCalculations(): Promise<CalculationRecord[]> {
  const response = await apiClient.get('/calculations');
  return response.data;
}

export async function getCalculationById(id: number): Promise<CalculationRecord> {
  const response = await apiClient.get(`/calculations/${id}`);
  return response.data;
}

export async function downloadAct(id: number): Promise<Blob> {
  const response = await apiClient.get(`/calculations/${id}/download`, {
    responseType: 'blob',
  });
  return response.data;
}