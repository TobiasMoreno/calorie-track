# CalorieTrack

Aplicación web personal para registrar comidas, estimar porciones y calorías y comparar lo registrado con un objetivo diario configurable.

La Fase 1 y el primer bloque de productividad de la Fase 2 funcionan en local mediante Docker Compose. La aplicación levanta tres servicios:

- `db`: PostgreSQL 17 con volumen persistente.
- `backend`: API REST en Go.
- `frontend`: React compilado y servido por Nginx, que también hace proxy de `/api` al backend.

No tiene autenticación ni requiere servicios cloud.

## Requisitos

- Docker Desktop con Docker Compose v2.
- Puertos libres por defecto: `5173`, `8080` y `5432`.

Go y Node.js sólo son necesarios para desarrollo o para ejecutar verificaciones fuera de Docker.

## Inicio rápido

Desde la raíz del proyecto:

```powershell
Copy-Item .env.example .env
./scripts/start.ps1
```

También se puede usar Docker Compose directamente:

```bash
docker compose up --build --detach
```

Abrir:

- Aplicación: <http://localhost:5173>
- Salud del backend: <http://localhost:8080/api/health>
- PostgreSQL: `localhost:5432`

Las migraciones se ejecutan automáticamente cuando inicia el backend.

## Detener la aplicación

```powershell
./scripts/stop.ps1
```

O bien:

```bash
docker compose down
```

`docker compose down` conserva el volumen con los datos. No usar `docker compose down --volumes` salvo que se quiera borrar definitivamente la base local.

## Configuración

El archivo raíz `.env` es opcional. Docker Compose utiliza valores locales seguros por defecto.

```dotenv
POSTGRES_DB=calorietrack
POSTGRES_USER=calorietrack
POSTGRES_PASSWORD=calorietrack_local
POSTGRES_PORT=5432
BACKEND_PORT=8080
FRONTEND_PORT=5173
```

Si se cambia `FRONTEND_PORT`, la URL será `http://localhost:<puerto>`.

La API acepta estas variables:

```dotenv
CALORIETRACK_ADDR=:8080
CALORIETRACK_DATABASE_URL=postgres://calorietrack:calorietrack_local@localhost:5432/calorietrack?sslmode=disable
CALORIETRACK_ALLOWED_ORIGIN=http://localhost:5173
```

## Persistencia

PostgreSQL guarda la información en el volumen Docker `calorietrack_postgres_data`. Los datos sobreviven a reinicios y a `docker compose down`.

### Crear un respaldo

Con los servicios levantados:

```powershell
docker compose exec -T db pg_dump -U calorietrack -d calorietrack --format=custom > calorietrack.backup
```

### Restaurar un respaldo

La restauración reemplaza información de la base seleccionada. Hacerla únicamente sobre la instalación local deseada:

```powershell
$dbContainer = docker compose ps -q db
docker cp ./calorietrack.backup "${dbContainer}:/tmp/calorietrack.backup"
docker compose exec -T db pg_restore -U calorietrack -d calorietrack --clean --if-exists /tmp/calorietrack.backup
```

## Verificaciones

Con Docker activo:

```powershell
./scripts/test.ps1
docker compose up --build --detach
./scripts/smoke.ps1
```

Verificaciones disponibles:

- Tests unitarios de calorías, porciones, checklist y agregaciones.
- Test del repositorio contra PostgreSQL real.
- Compilación TypeScript y build de producción de Vite.
- Smoke test de base, backend y frontend.

Comandos individuales:

```powershell
cd backend
go test ./...

cd ../frontend
npm ci
npm run build
```

El test del repositorio PostgreSQL se activa al definir `CALORIETRACK_TEST_DATABASE_URL`; sin esa variable se omite para permitir ejecutar los tests unitarios sin infraestructura.

## Arquitectura

```text
Browser
   │
   ▼
Frontend React + Nginx :5173
   │  /api
   ▼
Backend Go :8080
   │
   ▼
PostgreSQL :5432
```

El backend mantiene una clean architecture pragmática:

```text
backend/
├── cmd/api/                         composición y servidor HTTP
└── internal/
    ├── domain/                      entidades y reglas nutricionales
    ├── application/                 casos de uso y puertos
    ├── infrastructure/postgres/     repositorio y migraciones
    └── transport/http/              rutas, DTOs y errores HTTP
```

El frontend separa acceso a datos, tipos y componentes reutilizables:

```text
frontend/src/
├── components/
├── lib/api.ts
├── lib/date.ts
├── types.ts
└── App.tsx
```

## Decisiones de dominio

- El objetivo diario empieza sin valor; la aplicación no asume 2.000 kcal.
- Cada fecha guarda un snapshot del objetivo vigente.
- Las tareas diarias se crean al consultar la fecha, sin cron.
- Pendiente, registrada y omitida son estados diferentes.
- El backend calcula las calorías.
- Cada ítem histórico guarda nombre, gramaje, calorías por 100 g y total aplicados.
- Cambiar una referencia nutricional no modifica registros anteriores.
- Un día sin información no se interpreta como consumo cero.
- Las fechas de negocio se manejan en `America/Argentina/Cordoba`.

## Modelo de datos

Tablas principales implementadas:

- `user_settings`: objetivo, zona horaria y comidas habituales.
- `daily_days`: fecha local y snapshot del objetivo.
- `daily_meal_tasks`: checklist y estado por comida.
- `foods`: catálogo y referencia nutricional editable.
- `food_portions`: equivalencias entre unidades y gramos.
- `meal_entries`: comida registrada y total histórico.
- `meal_entry_items`: alimentos y snapshots nutricionales.
- `recipes`: preparaciones frecuentes y cantidad de porciones.
- `recipe_ingredients`: ingredientes, porciones y gramajes de cada receta.

Las tablas `weight_entries` y `food_photos` preparan las fases siguientes, pero sus funcionalidades todavía no están expuestas.

## API disponible

| Método | Ruta | Uso |
| --- | --- | --- |
| `GET` | `/api/health` | Salud del backend |
| `GET` | `/api/settings` | Consultar configuración |
| `PUT` | `/api/settings` | Actualizar objetivo y checklist habitual |
| `GET` | `/api/dashboard?date=YYYY-MM-DD` | Consultar o inicializar un día |
| `GET` | `/api/calendar?month=YYYY-MM` | Consultar el resumen mensual sin inicializar días vacíos |
| `GET` | `/api/foods?q=` | Buscar alimentos y porciones |
| `PATCH` | `/api/foods/{id}` | Corregir una referencia nutricional |
| `POST` | `/api/foods/{id}/portions` | Crear una porción personal |
| `PUT` | `/api/portions/{id}` | Editar una porción personal |
| `PATCH` | `/api/portions/{id}/default` | Elegir la porción predeterminada de un alimento |
| `DELETE` | `/api/portions/{id}` | Eliminar una porción personal conservando snapshots históricos |
| `GET` | `/api/recipes?q=` | Buscar comidas frecuentes |
| `POST` | `/api/recipes` | Crear una receta con ingredientes |
| `PUT` | `/api/recipes/{id}` | Editar una receta |
| `DELETE` | `/api/recipes/{id}` | Eliminar una receta sin alterar registros históricos |
| `POST` | `/api/recipe-entries` | Registrar porciones de una receta en una comida |
| `POST` | `/api/meal-entries` | Registrar una comida |
| `PUT` | `/api/meal-entries/{id}` | Editar y recalcular un registro histórico |
| `DELETE` | `/api/meal-entries/{id}` | Eliminar un registro |
| `PATCH` | `/api/tasks/{id}` | Marcar pendiente u omitida |
| `POST` | `/api/tasks` | Agregar una comida o snack |

Los errores tienen una forma consistente:

```json
{
  "error": {
    "code": "validation_error",
    "message": "descripción del problema"
  }
}
```

## Desarrollo sin reconstruir las imágenes

Levantar sólo PostgreSQL:

```bash
docker compose up --detach db
```

Backend:

```powershell
cd backend
$env:CALORIETRACK_DATABASE_URL="postgres://calorietrack:calorietrack_local@localhost:5432/calorietrack?sslmode=disable"
go run ./cmd/api
```

Frontend, en otra terminal:

```powershell
cd frontend
npm install
npm run dev
```

## Alcance pendiente

La Fase 1 está cerrada. La Fase 2 incluye comidas frecuentes, porciones personales, calendario mensual, historial por día y edición de registros. El siguiente bloque recomendado es el seguimiento de peso de la Fase 3. El orden completo está en [ROADMAP.md](./ROADMAP.md). Fotografías e IA permanecen fuera de esta entrega.

