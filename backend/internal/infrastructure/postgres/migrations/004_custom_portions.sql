ALTER TABLE food_portions
    ADD COLUMN IF NOT EXISTS is_custom BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_food_portions_custom_name
    ON food_portions(food_id, lower(name))
    WHERE is_custom = TRUE;

CREATE INDEX IF NOT EXISTS idx_food_portions_custom
    ON food_portions(food_id, is_custom, is_default);
