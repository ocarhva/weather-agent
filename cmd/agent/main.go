package main

import (
	"fmt"
	"log"
	"weather-agent/internal/adapters/classifier"
	"weather-agent/internal/adapters/websocket"
)

func main() {
	dllPath := "models/onnxruntime.dll"
	modelPath := "models/climate_model.onnx"

	fmt.Println("⚡ Inicializando Climate Alert Agent...")
	_ = classifier.InitONNX(dllPath)

	predictor, err := classifier.NewPredictor(modelPath)
	if err != nil {
		log.Fatalf("❌ Error cargando predictor: %v", err)
	}

	// Inicia el servidor WebSocket y UI Web en http://localhost:8080
	websocket.StartServer(":8080", predictor)
}