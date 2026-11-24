CREATE TABLE IF NOT EXISTS user_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    daily_calorie_target DOUBLE PRECISION CHECK (daily_calorie_target IS NULL OR daily_calorie_target BETWEEN 500 AND 10000),
    timezone TEXT NOT NULL DEFAULT 'America/Argentina/Cordoba',
    default_meals_json TEXT NOT NULL DEFAULT '["Desayuno","Almuerzo","Merienda","Cena"]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS daily_days (
    local_date DATE PRIMARY KEY,
    calorie_target_snapshot DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS daily_meal_tasks (
    id BIGSERIAL PRIMARY KEY,
    local_date DATE NOT NULL REFERENCES daily_days(local_date) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (char_length(trim(name)) BETWEEN 2 AND 60),
    slot_key TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','registered','skipped')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(local_date, slot_key)
);
CREATE INDEX IF NOT EXISTS idx_daily_meal_tasks_date_position ON daily_meal_tasks(local_date, position);
CREATE INDEX IF NOT EXISTS idx_daily_meal_tasks_status ON daily_meal_tasks(status);

CREATE TABLE IF NOT EXISTS foods (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    calories_per_100g DOUBLE PRECISION NOT NULL CHECK (calories_per_100g >= 0),
    reference_source TEXT NOT NULL,
    reference_detail TEXT NOT NULL,
    editable BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_foods_name_lower ON foods(lower(name));
CREATE INDEX IF NOT EXISTS idx_foods_category_name ON foods(category, name);

CREATE TABLE IF NOT EXISTS food_portions (
    id BIGSERIAL PRIMARY KEY,
    food_id BIGINT NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    unit TEXT NOT NULL CHECK (unit IN ('unidad','porción','cucharada')),
    size TEXT NOT NULL DEFAULT 'mediano',
    grams DOUBLE PRECISION NOT NULL CHECK (grams > 0),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(food_id, name)
);
CREATE INDEX IF NOT EXISTS idx_food_portions_food ON food_portions(food_id);

CREATE TABLE IF NOT EXISTS recipes (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    servings DOUBLE PRECISION NOT NULL DEFAULT 1 CHECK (servings > 0),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_recipes_name_lower ON recipes(lower(name));

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id BIGSERIAL PRIMARY KEY,
    recipe_id BIGINT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    food_id BIGINT NOT NULL REFERENCES foods(id),
    quantity DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    unit TEXT NOT NULL,
    estimated_grams DOUBLE PRECISION NOT NULL CHECK (estimated_grams > 0),
    position INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe ON recipe_ingredients(recipe_id, position);

CREATE TABLE IF NOT EXISTS meal_entries (
    id BIGSERIAL PRIMARY KEY,
    local_date DATE NOT NULL REFERENCES daily_days(local_date),
    task_id BIGINT NOT NULL UNIQUE REFERENCES daily_meal_tasks(id),
    meal_name_snapshot TEXT NOT NULL,
    total_calories DOUBLE PRECISION NOT NULL CHECK (total_calories >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_meal_entries_date ON meal_entries(local_date, created_at);

CREATE TABLE IF NOT EXISTS meal_entry_items (
    id BIGSERIAL PRIMARY KEY,
    meal_entry_id BIGINT NOT NULL REFERENCES meal_entries(id) ON DELETE CASCADE,
    food_id BIGINT NOT NULL REFERENCES foods(id),
    food_name_snapshot TEXT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    unit TEXT NOT NULL,
    portion_id BIGINT REFERENCES food_portions(id) ON DELETE SET NULL,
    portion_name_snapshot TEXT NOT NULL DEFAULT '',
    estimated_grams DOUBLE PRECISION NOT NULL CHECK (estimated_grams > 0),
    calories_per_100g_snapshot DOUBLE PRECISION NOT NULL CHECK (calories_per_100g_snapshot >= 0),
    calculated_calories DOUBLE PRECISION NOT NULL CHECK (calculated_calories >= 0),
    cooking_method TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_meal_entry_items_entry ON meal_entry_items(meal_entry_id, position);

CREATE TABLE IF NOT EXISTS weight_entries (
    id BIGSERIAL PRIMARY KEY,
    local_date DATE NOT NULL UNIQUE,
    weight_kg DOUBLE PRECISION NOT NULL CHECK (weight_kg BETWEEN 20 AND 500),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS food_photos (
    id BIGSERIAL PRIMARY KEY,
    meal_entry_id BIGINT REFERENCES meal_entries(id) ON DELETE SET NULL,
    local_date DATE NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    mime_type TEXT NOT NULL,
    recognition_provider TEXT NOT NULL DEFAULT 'disabled',
    recognition_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_food_photos_date ON food_photos(local_date);

INSERT INTO user_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

