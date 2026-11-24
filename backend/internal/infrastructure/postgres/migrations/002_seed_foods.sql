INSERT INTO foods (name, category, calories_per_100g, reference_source, reference_detail) VALUES
('Milanesa de carne', 'Carnes', 297, 'Referencia local editable', 'Estimación orientativa para milanesa cocida. Varía según corte, empanado y aceite; ajustá el valor a tu receta o envase.'),
('Puré de papa', 'Guarniciones', 113, 'USDA FoodData Central', 'Referencia para puré preparado con leche y grasa. La receta doméstica puede variar; el valor es editable.'),
('Arroz blanco cocido', 'Cereales', 130, 'USDA FoodData Central', 'Referencia para arroz blanco cocido sin agregados. Aceite y salsas deben registrarse aparte.'),
('Fideos cocidos', 'Cereales', 158, 'USDA FoodData Central', 'Referencia para pasta cocida sin salsa. Registrá queso, aceite y salsa como alimentos adicionales.'),
('Carne vacuna cocida', 'Carnes', 250, 'USDA FoodData Central', 'Referencia genérica para carne vacuna cocida. El corte y la grasa visible modifican el valor.'),
('Pechuga de pollo cocida', 'Carnes', 165, 'USDA FoodData Central', 'Referencia para pechuga cocida sin piel. El aceite o empanado se registra aparte.'),
('Pan francés', 'Panificados', 270, 'Referencia local editable', 'Valor medio orientativo; verificá el peso y la información de la panadería o envase.'),
('Huevo', 'Huevos', 143, 'USDA FoodData Central', 'Referencia por 100 g de huevo entero. Las porciones convierten unidades a gramos.'),
('Banana', 'Frutas', 89, 'USDA FoodData Central', 'Referencia para banana cruda, parte comestible.'),
('Manzana', 'Frutas', 52, 'USDA FoodData Central', 'Referencia para manzana cruda con piel, parte comestible.'),
('Leche entera', 'Lácteos', 61, 'USDA FoodData Central', 'Referencia para leche entera. Preferí el valor del envase cuando esté disponible.'),
('Yogur entero natural', 'Lácteos', 61, 'USDA FoodData Central', 'Referencia para yogur natural entero sin azúcar. Preferí el valor del envase.'),
('Queso cremoso', 'Lácteos', 300, 'Referencia local editable', 'Valor medio orientativo. La composición cambia mucho entre marcas; ajustalo desde el catálogo.'),
('Aceite', 'Grasas', 884, 'USDA FoodData Central', 'Referencia para aceite vegetal. Una cucharada se estima en 14 g.'),
('Papa hervida', 'Guarniciones', 87, 'USDA FoodData Central', 'Referencia para papa hervida sin agregados.'),
('Tomate', 'Verduras', 18, 'USDA FoodData Central', 'Referencia para tomate crudo.'),
('Palta', 'Frutas', 160, 'USDA FoodData Central', 'Referencia para palta cruda, parte comestible.'),
('Dulce de leche', 'Dulces', 315, 'Referencia local editable', 'Valor medio orientativo. Usá la etiqueta de tu marca para mayor precisión.'),
('Mate cocido con leche', 'Bebidas', 45, 'Referencia local editable', 'Estimación orientativa dependiente de la proporción de leche y azúcar. Ajustá o registrá agregados.'),
('Galletitas de agua', 'Panificados', 430, 'Referencia local editable', 'Valor medio orientativo. Preferí el valor nutricional del envase.')
ON CONFLICT DO NOTHING;

INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Unidad pequeña', 'unidad', 'pequeño', 100, FALSE FROM foods WHERE name='Milanesa de carne' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Unidad mediana', 'unidad', 'mediano', 140, TRUE FROM foods WHERE name='Milanesa de carne' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Unidad grande', 'unidad', 'grande', 190, FALSE FROM foods WHERE name='Milanesa de carne' ON CONFLICT DO NOTHING;

INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Poca cantidad', 'porción', 'pequeño', 120, FALSE FROM foods WHERE name='Puré de papa' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Porción normal', 'porción', 'mediano', 200, TRUE FROM foods WHERE name='Puré de papa' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Porción abundante', 'porción', 'grande', 300, FALSE FROM foods WHERE name='Puré de papa' ON CONFLICT DO NOTHING;

INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Porción normal', 'porción', 'mediano', 180, TRUE FROM foods WHERE name IN ('Arroz blanco cocido','Fideos cocidos') ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Porción abundante', 'porción', 'grande', 280, FALSE FROM foods WHERE name IN ('Arroz blanco cocido','Fideos cocidos') ON CONFLICT DO NOTHING;

INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Unidad mediana', 'unidad', 'mediano', 50, TRUE FROM foods WHERE name='Huevo' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Unidad mediana', 'unidad', 'mediano', 120, TRUE FROM foods WHERE name='Banana' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Unidad mediana', 'unidad', 'mediano', 180, TRUE FROM foods WHERE name='Manzana' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Rodaja', 'unidad', 'mediano', 35, TRUE FROM foods WHERE name='Pan francés' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Vaso', 'porción', 'mediano', 200, TRUE FROM foods WHERE name='Leche entera' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Pote', 'unidad', 'mediano', 190, TRUE FROM foods WHERE name='Yogur entero natural' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Cucharada', 'cucharada', 'mediano', 14, TRUE FROM foods WHERE name='Aceite' ON CONFLICT DO NOTHING;
INSERT INTO food_portions (food_id, name, unit, size, grams, is_default)
SELECT id, 'Cucharada', 'cucharada', 'mediano', 20, TRUE FROM foods WHERE name='Dulce de leche' ON CONFLICT DO NOTHING;

