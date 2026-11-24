# CalorieTrack — Estado y roadmap

Última revisión: 23 de noviembre de 2026.

## Objetivo del proyecto

CalorieTrack es una aplicación web local y personal para registrar comidas, estimar calorías sin exigir que todo sea pesado, controlar un objetivo diario configurable y detectar hábitos a partir de información registrada de forma explícita.

Principios del producto:

- Funcionar completamente en localhost y sin servicios externos obligatorios.
- No confundir una comida pendiente con consumo cero.
- No modificar retroactivamente el historial cuando cambia una referencia nutricional.
- Mostrar las estimaciones como aproximaciones editables, no como mediciones exactas.
- Priorizar un registro cotidiano rápido y una arquitectura fácil de mantener.

## Estado actual

### Fase 1 — Base funcional

Estado: funcional de punta a punta, pendiente de cierre documental y operativo.

Implementado:

- Monorepo con `backend` y `frontend`.
- Backend en Go organizado en dominio, aplicación, infraestructura y transporte HTTP.
- API REST con validaciones y manejo centralizado de errores.
- PostgreSQL 17 persistente con migraciones automáticas.
- Docker Compose con servicios `db`, `backend` y `frontend`.
- Tablas e índices para configuración, días, checklist, alimentos, porciones, registros, recetas futuras, peso futuro y fotografías futuras.
- Configuración manual del objetivo diario, inicialmente sin valor predeterminado.
- Snapshot del objetivo por fecha para preservar el historial.
- Zona horaria `America/Argentina/Cordoba`.
- Checklist automático de desayuno, almuerzo, merienda y cena al consultar una fecha.
- Estados de comida pendiente, registrada y omitida voluntariamente.
- Creación de comidas adicionales y snacks.
- Registro manual con múltiples alimentos o ingredientes.
- Cantidad por gramos o porciones habituales.
- Tamaño estimado, método de cocción y notas.
- Cálculo nutricional realizado por el backend.
- Snapshot histórico del alimento, gramaje y calorías aplicadas.
- Catálogo local inicial con fuentes y advertencias visibles.
- Edición personal del valor calórico y detalle de referencia de un alimento.
- Dashboard con fecha, navegación diaria, calorías registradas, objetivo, diferencia y progreso.
- Resumen gráfico de los últimos siete días.
- Días sin registros representados como información ausente, no como cero calorías.
- Eliminación de registros con confirmación y restauración de la tarea a pendiente.
- Interfaz responsive con dark mode, estados de carga, error y vacío.
- TanStack Query para comunicación y caché del frontend.
- Recharts para el resumen semanal.
- Tests de cálculos, porciones, checklist, agregaciones e idempotencia PostgreSQL.
- Build del frontend y tests del backend exitosos.
- Flujo HTTP real y persistencia tras reinicio verificados contra PostgreSQL.
- Inspección visual en 1440, 768 y 390 píxeles con cero errores del inspector.

### Cierre formal de la Fase 1

- [x] Crear un `README.md` con arquitectura, requisitos e instrucciones completas.
- [x] Agregar scripts de preparación y ejecución para Windows y comandos Docker Compose portables.
- [x] Documentar endpoints y ejemplos de uso de la API.
- [x] Documentar cómo respaldar y restaurar PostgreSQL.
- [x] Convertir el recorrido HTTP manual en un test de integración automatizado.
- [ ] Ejecutar una última revisión de los modales de registro y configuración en desktop y celular.
- [x] Separar Recharts del bundle inicial mediante carga diferida.
- [x] Limpiar los datos del recorrido de integración y revisar el contenido final.

La Fase 1 se considera terminada cuando una instalación limpia puede ejecutar los scripts documentados, abrir la aplicación, configurar un objetivo, registrar una comida, reiniciar los servicios y recuperar los mismos datos desde PostgreSQL.

## Orden recomendado a partir de ahora

### Fase 1 cerrada — mantenimiento inmediato

La instalación limpia, las migraciones, la persistencia tras reinicio, los tests, el build y el smoke test están verificados. Antes del primer commit estable sólo queda una revisión visual manual de los modales, porque el host de automatización de navegador no estuvo disponible durante la última ejecución.

El siguiente desarrollo funcional puede comenzar en la Fase 2.

### Primera funcionalidad de Fase 2 — Comidas frecuentes

Es la mejora con mayor impacto sobre el objetivo de registrar una comida en menos de 15 segundos.

Alcance recomendado:

- Crear una receta con nombre, ingredientes y cantidades.
- Calcular sus calorías desde las referencias de cada ingrediente.
- Definir cantidad de porciones producidas.
- Guardar snapshots nutricionales al registrar una receta.
- Registrar una receta desde el dashboard con uno o dos clics.
- Permitir cambiar el multiplicador antes de confirmar.
- Mostrar claramente cuándo una receta usa una referencia modificada desde su última utilización.

Criterios de aceptación:

- Una preparación habitual puede guardarse una vez y reutilizarse.
- Registrar una receta frecuente requiere menos de 15 segundos.
- Cambiar la receta no modifica registros históricos.
- El usuario puede revisar los ingredientes antes de confirmar.
- El registro marca automáticamente la tarea correspondiente como registrada.

### Segunda funcionalidad de Fase 2 — Porciones personalizadas

- Crear, editar y eliminar porciones asociadas a un alimento.
- Guardar nombres como “mi milanesa mediana” o “mi plato normal”.
- Registrar el gramaje corregido por el usuario.
- Ofrecer la porción personal como predeterminada en registros posteriores.
- Mantener las porciones históricas como snapshots.

### Tercera funcionalidad de Fase 2 — Historial y edición

- Calendario mensual con estados por día.
- Pantalla de detalle histórico.
- Edición de una comida registrada.
- Edición de cantidades, porciones, cocción y notas.
- Recalcular únicamente el registro editado.
- Mantener intactos los demás días.

## Fase 3 — Seguimiento y análisis

Pendiente:

- [ ] Peso inicial, peso objetivo opcional y fecha de inicio.
- [ ] Registro de peso por fecha.
- [ ] Evolución del peso.
- [ ] Estadísticas de siete y treinta días.
- [ ] Promedio calculado sólo sobre días con registros válidos.
- [ ] Comparación con el objetivo configurado.
- [ ] Distribución por comida.
- [ ] Comidas y preparaciones de mayor aporte.
- [ ] Días incompletos y comidas pendientes.
- [ ] Frecuencia de alimentos y recetas.
- [ ] Observaciones determinísticas, neutrales y no culpabilizantes.

Las estadísticas no deben tratar un día incompleto como un día de consumo bajo.

## Fase 4 — Fotografías e IA opcional

Pendiente:

- [ ] Carga y captura de fotografías.
- [ ] Conservación local de la foto aunque no haya IA.
- [ ] Interfaz `FoodRecognitionProvider`.
- [ ] `DisabledProvider` como comportamiento predeterminado.
- [ ] `GeminiProvider` activado sólo cuando existe una clave configurada.
- [ ] `OllamaProvider` activado sólo cuando hay un modelo compatible disponible.
- [ ] Resultado estructurado con alimentos, cantidades, cocción, confianza y preguntas.
- [ ] Revisión y corrección obligatorias antes de guardar.
- [ ] Cálculo final realizado por el backend, nunca por la respuesta de la IA.

## Decisiones técnicas vigentes

- PostgreSQL es la fuente de verdad local y persiste en un volumen Docker.
- El frontend nunca calcula el valor histórico definitivo.
- Cada registro almacena los valores nutricionales aplicados.
- Las fechas del dominio usan formato local `YYYY-MM-DD`.
- Las tareas se generan bajo demanda, sin cron.
- No hay autenticación porque la aplicación es personal y local.
- Las integraciones de IA serán adaptadores opcionales y reemplazables.
- Docker Compose es el método principal de ejecución local para `db`, `backend` y `frontend`.

## Próxima tarea concreta

> Implementar comidas frecuentes y recetas con snapshots históricos y registro desde el dashboard en menos de 15 segundos.

No comenzar todavía por IA ni estadísticas. Las recetas son la funcionalidad que más mejora el uso diario con el menor aumento de complejidad.
