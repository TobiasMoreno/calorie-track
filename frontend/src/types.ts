export type TaskStatus = 'pending' | 'registered' | 'skipped'

export interface UserSettings {
  dailyCalorieTarget: number | null
  timezone: string
  defaultMeals: string[]
  updatedAt: string
}

export interface DailyMealTask {
  id: number
  date: string
  name: string
  slotKey: string
  position: number
  status: TaskStatus
  entryId?: number
}

export interface FoodPortion {
  id: number
  foodId: number
  name: string
  unit: 'unidad' | 'porción' | 'cucharada'
  size: string
  grams: number
  isDefault: boolean
  isCustom: boolean
}

export interface FoodPortionInput {
  name: string
  unit: 'unidad' | 'porción' | 'cucharada'
  size?: string
  grams: number
  isDefault: boolean
}

export interface Food {
  id: number
  name: string
  category: string
  caloriesPer100g: number
  referenceSource: string
  referenceDetail: string
  editable: boolean
  portions: FoodPortion[]
}

export interface MealEntryItem {
  id: number
  foodId: number
  foodName: string
  quantity: number
  unit: string
  portionId?: number
  portionName?: string
  estimatedGrams: number
  caloriesPer100g: number
  calories: number
  cookingMethod?: string
  notes?: string
}

export interface MealEntry {
  id: number
  date: string
  taskId: number
  mealName: string
  recipeId?: number
  recipeName?: string
  recipeMultiplier?: number
  totalCalories: number
  items: MealEntryItem[]
  createdAt: string
}

export interface RecipeIngredient {
  id: number
  recipeId: number
  foodId: number
  foodName: string
  quantity: number
  unit: string
  portionId?: number
  portionName?: string
  estimatedGrams: number
  caloriesPer100g: number
  calories: number
  cookingMethod?: string
  notes?: string
}

export interface Recipe {
  id: number
  name: string
  servings: number
  notes?: string
  ingredients: RecipeIngredient[]
  totalCalories: number
  caloriesPerServing: number
  createdAt: string
  updatedAt: string
}

export interface RecipeInput {
  name: string
  servings: number
  notes?: string
  items: CreateItemInput[]
}

export interface DaySummary {
  date: string
  calories: number
  target: number | null
  taskCount: number
  registeredTasks: number
  skippedTasks: number
  pendingTasks: number
  hasData: boolean
}

export interface CalendarMonth {
  month: string
  days: DaySummary[]
}

export interface Dashboard {
  date: string
  targetCalories: number | null
  consumedCalories: number
  remainingCalories: number | null
  tasks: DailyMealTask[]
  entries: MealEntry[]
  lastSevenDays: DaySummary[]
}

export interface CreateItemInput {
  foodId: number
  quantity: number
  unit: string
  portionId?: number
  cookingMethod?: string
  notes?: string
}

export interface UpdateMealEntryInput {
  items: CreateItemInput[]
}

export interface ApiError { error: { code: string; message: string } }

