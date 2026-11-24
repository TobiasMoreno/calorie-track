import type { CreateItemInput, Dashboard, Food, TaskStatus, UserSettings } from '../types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (!response.ok) {
    const payload = await response.json().catch(() => null)
    throw new Error(payload?.error?.message ?? `Error ${response.status}`)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const api = {
  dashboard: (date: string) => request<Dashboard>(`/dashboard?date=${encodeURIComponent(date)}`),
  settings: () => request<UserSettings>('/settings'),
  saveSettings: (settings: Pick<UserSettings, 'dailyCalorieTarget' | 'timezone' | 'defaultMeals'>) =>
    request<UserSettings>('/settings', { method: 'PUT', body: JSON.stringify(settings) }),
  foods: (query = '') => request<Food[]>(`/foods?q=${encodeURIComponent(query)}`),
  updateFood: (id: number, input: { caloriesPer100g: number; referenceDetail: string }) =>
    request<Food>(`/foods/${id}`, { method: 'PATCH', body: JSON.stringify(input) }),
  createEntry: (input: { date: string; taskId: number; items: CreateItemInput[] }) =>
    request('/meal-entries', { method: 'POST', body: JSON.stringify(input) }),
  deleteEntry: (id: number) => request<void>(`/meal-entries/${id}`, { method: 'DELETE' }),
  updateTask: (id: number, status: Exclude<TaskStatus, 'registered'>) =>
    request<void>(`/tasks/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  createTask: (date: string, name: string) => request('/tasks', { method: 'POST', body: JSON.stringify({ date, name }) }),
}

