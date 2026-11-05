package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"weather-agent/internal/adapters/classifier"
	"weather-agent/internal/adapters/websocket"
)

func main() {
	dllPath := "models/onnxruntime.dll"
	modelPath := "models/climate_model.onnx"

	fmt.Println("⚡ Inicializando Climate Alert Agent...")
	
	// Intentamos inicializar ONNX pero sin detener el programa si falla localmente
	if err := classifier.InitONNX(dllPath); err != nil {
		log.Printf("⚠️ Advertencia: No se pudo cargar ONNX CGO en Windows local: %v", err)
	}

	predictor, err := classifier.NewPredictor(modelPath)
	if err != nil {
		log.Printf("⚠️ No se pudo cargar el modelo ONNX localmente, continuando sin modelo: %v", err)
	}

	// Probamos la función de NOAA para verificar que responda en consola
	fmt.Println("📡 Probando conexión con NOAA NWS API...")
	alertas, err := obtenerAlertasNOAA()
	if err != nil {
		log.Printf("❌ Error al obtener alertas de NOAA: %v", err)
	} else {
		fmt.Printf("✅ NOAA NWS conectado correctamente. Bytes recibidos: %d\n", len(alertas))
	}

	// Inicia el servidor WebSocket y UI Web en http://localhost:8080
	fmt.Println("🚀 Servidor listo en http://localhost:8080")
	websocket.StartServer(":8080", predictor)
}

func obtenerAlertasNOAA() ([]byte, error) {
	req, err := http.NewRequest(
		"GET",
		"https://api.weather.gov/alerts/active?area=CA",
		nil,
	)
	if err != nil {
		return nil, err
	}

	userAgent := os.Getenv("NOAA_USER_AGENT")
	if userAgent == "" {
		userAgent = "Agripilot/1.0 (ocarhvara@gmail.com)"
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/geo+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}