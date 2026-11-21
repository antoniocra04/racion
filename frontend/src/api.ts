export interface Nutrition {
  protein: number;
  fat: number;
  carbs: number;
  kcal: number;
}

export interface Product {
  id: number;
  revision: number;
  name: string;
  state: string;
  per100: Nutrition;
}

export interface MealItem {
  productId: number;
  productName: string;
  productState: string;
  grams: number;
  nutrition: Nutrition;
}

export interface Meal {
  name: string;
  items: MealItem[];
  nutrition: Nutrition;
}

export interface Plan {
  id: number;
  revision: number;
  name: string;
  active: boolean;
  meals: Meal[];
  nutrition: Nutrition;
}

export interface RationState {
  plans: Plan[];
  products: Product[];
}

export type Reload = () => Promise<boolean>;

export class ApiError extends Error {
  fields: Record<string, string>;
  constructor(message: string, fields: Record<string, string> = {}) {
    super(message);
    this.fields = fields;
  }
}

export async function request<T>(path: string, options?: { method: string; body?: unknown }): Promise<T> {
  const hasBody = options?.body !== undefined;
  let response: Response;
  try {
    response = await fetch('/api' + path, {
      method: options?.method,
      headers: hasBody ? { 'Content-Type': 'application/json' } : undefined,
      body: hasBody ? JSON.stringify(options?.body) : undefined,
    });
  } catch {
    throw new ApiError('Связь с сервером прервалась. Проверь подключение и попробуй ещё раз.');
  }
  if (response.status === 204) return undefined as T;
  if (!(response.headers.get('Content-Type')?.includes('application/json'))) {
    throw new ApiError('Сервер не смог выполнить запрос. Попробуй ещё раз.');
  }
  const data = await response.json();
  if (!response.ok) throw new ApiError(data.error ?? 'Не удалось сохранить изменения.', data.fields);
  return data as T;
}

const whole = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 });
const fractional = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 1 });
export const kcal = (n: number) => whole.format(n);
export const grams = (n: number) => fractional.format(n);
export function decimal(n: number) {
  const [digits, exponent] = String(n).split('e-');
  return exponent ? '0,' + '0'.repeat(Number(exponent) - 1) + digits.replace('.', '') : digits.replace('.', ',');
}

export function creationToken() {
  return Array.from(crypto.getRandomValues(new Uint8Array(16)), (n) => n.toString(16).padStart(2, '0')).join('');
}
