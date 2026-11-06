package classifier

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

// Predictor evalua el modelo ONNX (RandomForest exportado con skl2onnx).
// Si el runtime o el modelo no estan disponibles, usa las mismas reglas con las
// que se entreno el modelo, para que el servicio siga funcionando.
type Predictor struct {
	modelPath string
	session   *ort.DynamicAdvancedSession
	mu        sync.Mutex
}

var initOnce sync.Once
var initErr error

// InitONNX carga la biblioteca compartida de ONNX Runtime (libonnxruntime.so en
// Linux) e inicializa el entorno. libPath vacio usa ONNXRUNTIME_LIB o la ruta por defecto.
func InitONNX(libPath string) error {
	initOnce.Do(func() {
		if libPath == "" {
			libPath = os.Getenv("ONNXRUNTIME_LIB")
		}
		if libPath == "" {
			libPath = "/usr/lib/libonnxruntime.so"
		}
		if _, err := os.Stat(libPath); err != nil {
			initErr = fmt.Errorf("no se encontro %s: %w", libPath, err)
			return
		}
		ort.SetSharedLibraryPath(libPath)
		if err := ort.InitializeEnvironment(); err != nil {
			initErr = fmt.Errorf("no se pudo inicializar ONNX Runtime: %w", err)
		}
	})
	return initErr
}

// NewPredictor abre el modelo. Siempre devuelve un Predictor usable; si hay error,
// el Predictor usa las reglas de respaldo y el error describe por que.
func NewPredictor(modelPath string) (*Predictor, error) {
	p := &Predictor{modelPath: modelPath}
	if err := InitONNX(""); err != nil {
		return p, err
	}
	// Solo se pide output_label: output_probability es una secuencia de mapas
	// (ZipMap) que el wrapper no soporta.
	s, err := ort.NewDynamicAdvancedSession(modelPath,
		[]string{"float_input"}, []string{"output_label"}, nil)
	if err != nil {
		return p, fmt.Errorf("no se pudo cargar el modelo %s: %w", modelPath, err)
	}
	p.session = s
	log.Printf("ONNX cargado: modelo=%s", filepath.Base(modelPath))
	return p, nil
}

// PredictRisk devuelve true si el modelo clasifica la entrada como riesgo agricola.
func (p *Predictor) PredictRisk(temp, humidity, wind, rain float32) (bool, error) {
	if p == nil || p.session == nil {
		return rulesRisk(temp, humidity), nil
	}
	in, err := ort.NewTensor(ort.NewShape(1, 4), []float32{temp, humidity, wind, rain})
	if err != nil {
		return rulesRisk(temp, humidity), err
	}
	defer in.Destroy()

	outputs := []ort.Value{nil}
	p.mu.Lock()
	err = p.session.Run([]ort.Value{in}, outputs)
	p.mu.Unlock()
	if err != nil {
		return rulesRisk(temp, humidity), err
	}
	defer outputs[0].Destroy()

	label, ok := outputs[0].(*ort.Tensor[int64])
	if !ok || len(label.GetData()) == 0 {
		log.Printf("salida ONNX inesperada: %T", outputs[0])
		return rulesRisk(temp, humidity), fmt.Errorf("salida ONNX inesperada")
	}
	return label.GetData()[0] == 1, nil
}

// rulesRisk: helada (<= 2C) o estres hidrico/incendio (>= 35C con humedad < 25%).
func rulesRisk(temp, humidity float32) bool {
	return temp <= 2.0 || (temp >= 35.0 && humidity < 25.0)
}
