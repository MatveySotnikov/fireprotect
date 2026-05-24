export interface User {
  id: number;
  email: string;
  name: string;
}

export interface LoginResponse {
  token: string;
}

export interface Material {
  ID: number;
  Title: string;
  DefaultDensity: number;
  Group1Consumption: number;
  Group2Consumption: number;
  BrushLoss: number;
  SprayIndoorLoss: number;
  SprayOutdoorLoss: number;
}

export interface CalculationRequest {
  area: number;
  area_type: 'projection' | 'slope';
  slope_angle: number;
  target_group: '1_group' | '2_group';
  application_method: 'brush' | 'spray_indoor' | 'spray_outdoor';
  material_id?: number;
  normative_rate?: number;
  density?: number;
}

export interface CalculationResponse {
  total_mass: number;
  total_volume: number;
}

export interface CalculationRecord {
  id: number;
  user_id: number;
  material_id: number | null;
  area: number;
  area_type: string;
  slope_angle: number;
  target_group: string;
  application_method: string;
  loss_factor: number;
  layers: number;
  used_normative_rate: number;
  used_density: number;
  total_mass: number;
  total_volume: number;
  user: User;
  material: Material | null;
}