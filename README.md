# AgriPilot — Agente Inteligente para Monitoreo Agrometeorológico

**AgriPilot** es un agente autónomo desarrollado en **Go (Golang)** diseñado para el monitoreo agrometeorológico de precisión mediante telemetría en tiempo real, alertas gubernamentales oficiales e inferencia local de aprendizaje automático en el borde (*Edge Computing*) con **ONNX Runtime** vía CGO.

- **Autora:** Jimena Huara Ocar
- **Institución:** Sociedad Argentina de Informática (SADIO) — *Golang para Sistemas Inteligentes (2026)*
- **Instancia en Producción:** https://weather-agent-agripilot.up.railway.app/
- **Repositorio:** https://github.com/ocarhva/weather-agent

---

## 1. Arquitectura del Sistema

El proyecto implementa el patrón de **Arquitectura Hexagonal (Ports and Adapters)** para desacoplar el núcleo de evaluación de riesgo agrometeorológico de los adaptadores de red e inferencia nativa:

```text
weather-agent/
├── cmd/
│   └── agent/
│       └── main.go                  # Punto de entrada e inicialización (Bootstrapping)
├── internal/
│   └── adapters/
│       ├── classifier/
│       │   ├── onnx_predictor.go    # Motor de inferencia CGO (ONNX Runtime) y fallback
│       │   └── onnx_predictor_test.go # Capa 1: Tests unitarios (R^4, failover y concurrencia)
│       ├── httpapi/
│       │   └── handler_test.go      # Capa 2: Tests de integración REST y mocking externo
│       └── websocket/
│           └── server.go            # Bucle de eventos (Ticker 3s), ingestión APIs y difusión WS
├── models/
│   └── climate_model.onnx           # Grafo estático serializado (Random Forest)
├── scripts/
│   └── train_and_export.py          # Pipeline MLOps de entrenamiento y exportación ONNX
├── static/ & web/
│   └── index.html                   # Interfaz cartográfica GIS (Leaflet.js + GeoJSON)
├── Dockerfile                       # Construcción multietapa (Multi-Stage Build)
├── go.mod                           # Definición del módulo y dependencias Go
└── go.sum                           # Sumas de verificación criptográfica
```

---

## 2. Requisitos Previos

* **Go:** Versión `1.22` o superior.
* **Docker:** Opcional, recomendado para ejecutar el entorno completo con la biblioteca nativa `libonnxruntime.so` preconfigurada.
* **Python 3.10+:** Opcional, únicamente requerido si se desea reentrenar y exportar nuevamente el archivo `models/climate_model.onnx` (`scikit-learn`, `skl2onnx`, `numpy`).

---

## 3. Variables de Entorno

Antes de ejecutar la aplicación, configure las siguientes variables de entorno (en local o en su proveedor *Cloud*):

| Variable | Descripción | Valor por Defecto / Ejemplo |
| :--- | :--- | :--- |
| `PORT` | Puerto HTTP en el que escucha el servidor | `8080` |
| `OPENWEATHER_API_KEY` | Clave de API de OpenWeatherMap para telemetría climática | *(Requerida para datos en vivo)* |
| `ONNXRUNTIME_LIB` | Ruta a la biblioteca dinámica C++ `libonnxruntime.so` (opcional en local; el sistema activa *failover* determinista si no está presente) | `/usr/lib/libonnxruntime.so` |

---

## 4. Instalación y Ejecución Local

### Paso 1: Clonar el repositorio e instalar dependencias
```bash
git clone [https://github.com/ocarhva/weather-agent.git](https://github.com/ocarhva/weather-agent.git)
cd weather-agent
go mod download
```

### Paso 2: Configurar variables de entorno
En **Linux / macOS**:
```bash
export PORT=8080
export OPENWEATHER_API_KEY="tu_api_key_aqui"
```

En **Windows (PowerShell)**:
```powershell
$env:PORT="8080"
$env:OPENWEATHER_API_KEY="tu_api_key_aqui"
```

### Paso 3: Ejecutar el servidor
```bash
go run ./cmd/agent/main.go
```
Una vez iniciado, acceda desde su navegador a: `http://localhost:8080`

> **Nota sobre Tolerancia a Fallos (*Failover*):** Si ejecuta el proyecto localmente en Windows/macOS sin tener compilada la librería compartida `libonnxruntime.so`, **AgriPilot** conmutará automáticamente y sin interrupciones hacia su motor determinista de reglas agrometeorológicas ($f_{\text{rules}}$).

---

## 5. Ejecución con Docker (Entorno Completo con ONNX CGO)

El proyecto incluye un `Dockerfile` con compilación multietapa (*Multi-Stage Build*) basado en Go Alpine y Debian Slim que empaqueta el binario junto con **ONNX Runtime v1.14.1**:

```bash
# 1. Construir la imagen de contenedor
docker build -t agripilot-agent .

# 2. Ejecutar el contenedor exponiendo el puerto 8080
docker run -d -p 8080:8080 \
  -e PORT=8080 \
  -e OPENWEATHER_API_KEY="tu_api_key_aqui" \
  --name agripilot \
  agripilot-agent
```

---

## 6. Ejecución de la Suite de Pruebas (QA)

El sistema cuenta con dos capas de verificación automatizada (detalladas en el Apéndice A del artículo):

1. **Capa 1 (Unitaria):** Validación en la frontera $\mathbb{R}^4$, tolerancia a fallos (*failover*) ante punteros nulos y seguridad concurrente con 50 *goroutines*.
2. **Capa 2 (Integración API):** Simulación HTTP (`httptest`) de endpoints internos (`/api/evaluate`, `/api/chat`) y consumo de APIs meteorológicas externas.

Para ejecutar todas las pruebas:
```bash
go test -v ./...
```

Para verificar la ausencia de condiciones de carrera (*data races*):
```bash
go test -race -v ./...
```

---

## 7. Reentrenamiento del Modelo MLOps (Opcional)

Si desea regenerar el grafo estático `models/climate_model.onnx`:
```bash
pip install numpy scikit-learn skl2onnx
python scripts/train_and_export.py
```
