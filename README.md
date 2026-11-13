# AgriPilot — monitoreo agrometeorológico

AgriPilot es un servicio en Go con una interfaz web para mostrar información meteorológica y riesgos agrícolas en ubicaciones de California. Usa un clasificador ONNX cuando el runtime está disponible y reglas de respaldo cuando no lo está.

- **Instancia:** https://weather-agent-agripilot.up.railway.app/
- **Repositorio:** https://github.com/ocarhva/weather-agent

> **Estado del prototipo:** OpenWeather y NOAA se consultan en ejecución cuando corresponde. El NDVI, parte de las mediciones de respaldo y parte del riesgo de granizo son simulados. Planet todavía no está integrado en el código de este repositorio.

## Arquitectura y módulos

```text
weather-agent/
├── cmd/agent/
│   └── main.go
│       Punto de entrada: inicializa el predictor, hace una comprobación
│       de NOAA y arranca el servidor.
├── internal/adapters/
│   ├── classifier/
│   │   ├── onnx_predictor.go
│   │   │   Carga y ejecuta el modelo ONNX; usa reglas de respaldo si
│   │   │   ONNX Runtime o el modelo no están disponibles.
│   │   └── onnx_predictor_test.go
│   │       Prueba la predicción ONNX y una regla de respaldo.
│   └── websocket/
│       └── server.go
│           Expone la interfaz web, el endpoint de chat y el WebSocket;
│           consulta OpenWeather/NOAA y prepara los datos de las granjas.
├── models/
│   └── climate_model.onnx
│       Modelo binario de clasificación usado por el predictor.
├── scripts/
│   └── train_and_export.py
│       Genera datos sintéticos, entrena un Random Forest y lo exporta
│       al formato ONNX.
├── web/
│   └── index.html
│       Interfaz de mapa, filtros, tarjetas de granja y chat.
├── Dockerfile
│   Construye la aplicación y empaqueta ONNX Runtime, el modelo y la web.
├── go.mod
│   Declara el módulo, la versión de Go y sus dependencias.
└── go.sum
    Registra las sumas de verificación de las dependencias.
```

El servidor atiende estas rutas:

| Ruta | Uso |
|---|---|
| `/` | Sirve la interfaz de `web/`. |
| `/ws` | Envía actualizaciones de las granjas por WebSocket cada tres segundos. |
| `/api/chat` | Acepta mensajes JSON mediante `POST` y devuelve una respuesta basada en reglas. |

El proyecto tiene una estructura pequeña inspirada en adaptadores, pero la lógica de datos, riesgos y chat está concentrada actualmente en `internal/adapters/websocket/server.go`; no hay un paquete separado de dominio o de API REST.

## Fuentes de datos y limitaciones

- **OpenWeather:** con `OPENWEATHER_API_KEY`, el servidor consulta temperatura, humedad y viento; también intenta obtener datos del suelo. Si falta la clave o falla la consulta meteorológica, usa valores simulados.
- **NOAA NWS:** consulta alertas activas por ubicación. Un error de consulta se presenta actualmente como “Sin alertas activas”, por lo que la interfaz no distingue el error de una ausencia de alertas.
- **Planet / NDVI:** no hay integración con Planet en el código actual. El NDVI se genera aleatoriamente; las alertas de plaga que dependen de él son demostrativas, no detecciones satelitales reales.
- **Riesgo de granizo:** se deriva de ciertas palabras en el texto de la alerta NOAA o de una probabilidad aleatoria.
- **Clasificador climático:** el modelo se entrena con datos sintéticos y una regla de etiquetado simple. El respaldo en Go considera helada (`temp <= 2 °C`) o calor seco (`temp >= 35 °C` y humedad `< 25 %`).
- **Chat:** responde con reglas de palabras clave; no es un modelo de lenguaje.
- **Mapa:** Leaflet carga teselas de OpenStreetMap y geometrías de condados desde una URL pública externa.

Por estas limitaciones, el proyecto debe considerarse un prototipo y no una fuente validada para decisiones agrícolas.

## Requisitos

- Go `1.27.3`, según `go.mod`.
- Docker, opcional.
- Python 3 y `numpy`, `scikit-learn`, `skl2onnx`, solo para regenerar el modelo.
- Para inferencia ONNX local: la biblioteca compartida ONNX Runtime compatible con `github.com/yalue/onnxruntime_go`. Sin ella, el predictor recurre a reglas.

## Variables de entorno

| Variable | Descripción | Valor por defecto |
|---|---|---|
| `PORT` | Puerto HTTP del servidor. | `8080` |
| `OPENWEATHER_API_KEY` | Clave de OpenWeather. Sin ella se simulan las mediciones meteorológicas. | Vacío |
| `ONNXRUNTIME_LIB` | Ruta a la biblioteca compartida de ONNX Runtime. | `/usr/lib/libonnxruntime.so` |
| `NOAA_USER_AGENT` | Identificador opcional para la comprobación inicial de NOAA en `main.go`. | Valor definido por la aplicación |

El endpoint NOAA dentro de `server.go` establece su propio `User-Agent`.

## Ejecutar localmente

Desde la raíz del repositorio:

```bash
git clone https://github.com/ocarhva/weather-agent.git
cd weather-agent
go mod download
export PORT=8080
export OPENWEATHER_API_KEY="tu_api_key"
go run ./cmd/agent
```

En PowerShell:

```powershell
git clone https://github.com/ocarhva/weather-agent.git
cd weather-agent
go mod download

$env:PORT="8080"
go run ./cmd/agent
```

Abre <http://localhost:8080>. La clave de OpenWeather es opcional; sin ella se muestran mediciones simuladas. Para intentar obtener datos meteorológicos reales, define `OPENWEATHER_API_KEY` antes de iniciar el servicio. Si ONNX Runtime no está disponible, se usan las reglas de respaldo.

## Ejecutar con Docker

El `Dockerfile` compila en `golang:1.27-bookworm` y ejecuta en `debian:bookworm-slim`. Descarga ONNX Runtime `1.19.2` para Linux x86-64.

```bash
docker build -t agripilot-agent .
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e OPENWEATHER_API_KEY="tu_api_key" \
  agripilot-agent
```

Luego abre <http://localhost:8080>.

## Pruebas

La suite actual contiene pruebas del clasificador en `internal/adapters/classifier/onnx_predictor_test.go`. La prueba ONNX se omite si `ONNXRUNTIME_LIB` no está definida; la prueba de respaldo no requiere el runtime.

```bash
go test ./...
go test -race ./...
```

Actualmente no hay pruebas automatizadas para los handlers HTTP/WebSocket, las consultas a OpenWeather o NOAA, el chat ni la interfaz web.

## Regenerar el modelo (opcional)

El script crea datos sintéticos y escribe el archivo en `../models/climate_model.onnx` relativo al directorio de trabajo actual. Ejecútalo desde `scripts/`:

```bash
python -m pip install numpy scikit-learn skl2onnx
cd scripts
python train_and_export.py
```

La precisión que imprime el script se calcula sobre los mismos datos usados para entrenar; no es una medición de rendimiento sobre datos independientes.
