import type { CalendarMonth, CreateItemInput, Dashboard, Food, FoodPortion, FoodPortionInput, Recipe, RecipeInput, TaskStatus, UpdateMealEntryInput, UserSettings } from '../types'

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
  calendar: (month: string) => request<CalendarMonth>(`/calendar?month=${encodeURIComponent(month)}`),
  settings: () => request<UserSettings>('/settings'),
  saveSettings: (settings: Pick<UserSettings, 'dailyCalorieTarget' | 'timezone' | 'defaultMeals'>) =>
    request<UserSettings>('/settings', { method: 'PUT', body: JSON.stringify(settings) }),
  foods: (query = '') => request<Food[]>(`/foods?q=${encodeURIComponent(query)}`),
  updateFood: (id: number, input: { caloriesPer100g: number; referenceDetail: string }) =>
    request<Food>(`/foods/${id}`, { method: 'PATCH', body: JSON.stringify(input) }),
  createPortion: (foodId: number, input: FoodPortionInput) => request<FoodPortion>(`/foods/${foodId}/portions`, { method: 'POST', body: JSON.stringify(input) }),
  updatePortion: (id: number, input: FoodPortionInput) => request<FoodPortion>(`/portions/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  setDefaultPortion: (id: number) => request<FoodPortion>(`/portions/${id}/default`, { method: 'PATCH' }),
  deletePortion: (id: number) => request<void>(`/portions/${id}`, { method: 'DELETE' }),
  recipes: (query = '') => request<Recipe[]>(`/recipes?q=${encodeURIComponent(query)}`),
  createRecipe: (input: RecipeInput) => request<Recipe>('/recipes', { method: 'POST', body: JSON.stringify(input) }),
  updateRecipe: (id: number, input: RecipeInput) => request<Recipe>(`/recipes/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  deleteRecipe: (id: number) => request<void>(`/recipes/${id}`, { method: 'DELETE' }),
  createRecipeEntry: (input: { date: string; taskId: number; recipeId: number; multiplier: number }) =>
    request('/recipe-entries', { method: 'POST', body: JSON.stringify(input) }),
  createEntry: (input: { date: string; taskId: number; items: CreateItemInput[] }) =>
    request('/meal-entries', { method: 'POST', body: JSON.stringify(input) }),
  updateEntry: (id: number, input: UpdateMealEntryInput) => request(`/meal-entries/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  deleteEntry: (id: number) => request<void>(`/meal-entries/${id}`, { method: 'DELETE' }),
  updateTask: (id: number, status: Exclude<TaskStatus, 'registered'>) =>
    request<void>(`/tasks/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  createTask: (date: string, name: string) => request('/tasks', { method: 'POST', body: JSON.stringify({ date, name }) }),
}

