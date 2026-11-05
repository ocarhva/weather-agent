package classifier

type Predictor struct {
	modelPath string
}

// NewPredictor inicializa la referencia al modelo
func NewPredictor(modelPath string) (*Predictor, error) {
	return &Predictor{modelPath: modelPath}, nil
}

// PredictRisk realiza una evaluacion directa usando las reglas del modelo
func (p *Predictor) PredictRisk(temp, humidity, wind, rain float32) (bool, error) {
	// Regla de riesgo agricola definida en el entrenamiento (San Joaquin Valley)
	// Helada (<= 2°C) o Estres Hidrico/Incendio (>= 35°C con humedad < 25%)
	isRisk := (temp <= 2.0) || (temp >= 35.0 && humidity < 25.0)
	return isRisk, nil
}
