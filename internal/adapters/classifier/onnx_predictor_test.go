package classifier

import (
	"os"
	"testing"
)

func TestPredictRiskONNX(t *testing.T) {
	if os.Getenv("ONNXRUNTIME_LIB") == "" {
		t.Skip("ONNXRUNTIME_LIB no definido")
	}
	p, err := NewPredictor("../../../models/climate_model.onnx")
	if err != nil || p.session == nil {
		t.Fatalf("modelo no cargado: %v", err)
	}
	cases := []struct {
		t, h, w, r float32
		want       bool
	}{{1, 50, 5, 0, true}, {20, 50, 5, 0, false}, {38, 10, 5, 0, true}, {38, 50, 5, 0, false}}
	for _, c := range cases {
		got, err := p.PredictRisk(c.t, c.h, c.w, c.r)
		if err != nil || got != c.want {
			t.Errorf("%v: got %v err %v want %v", c, got, err, c.want)
		}
	}
}

func TestFallbackWithoutRuntime(t *testing.T) {
	var p *Predictor
	if r, _ := p.PredictRisk(0, 50, 0, 0); !r {
		t.Error("frio debe ser riesgo")
	}
}
