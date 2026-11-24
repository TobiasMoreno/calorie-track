ALTER TABLE recipe_ingredients
    ADD COLUMN IF NOT EXISTS portion_id BIGINT REFERENCES food_portions(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS portion_name_snapshot TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS cooking_method TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';

ALTER TABLE meal_entries
    ADD COLUMN IF NOT EXISTS recipe_id BIGINT REFERENCES recipes(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS recipe_name_snapshot TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS recipe_multiplier DOUBLE PRECISION;

CREATE INDEX IF NOT EXISTS idx_meal_entries_recipe ON meal_entries(recipe_id);

