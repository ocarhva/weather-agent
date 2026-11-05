import numpy as np
from sklearn.ensemble import RandomForestClassifier
from skl2onnx import convert_sklearn
from skl2onnx.common.data_types import FloatTensorType
import os

# 1. Datos simulados del clima en San Joaquín Valley, California
np.random.seed(42)
n_samples = 1000

# Variables: [temperatura (°C), humedad (%), viento (km/h), lluvia (mm)]
temp = np.random.uniform(-5, 38, n_samples)
humidity = np.random.uniform(10, 95, n_samples)
wind = np.random.uniform(0, 50, n_samples)
rain = np.random.uniform(0, 20, n_samples)

# Regla de riesgo agrícola: Helada (<= 2°C) o Estrés Hídrico/Incendio (>= 35°C con humedad < 25%)
is_risk = ((temp <= 2.0) | ((temp >= 35.0) & (humidity < 25.0))).astype(int)

X = np.column_stack((temp, humidity, wind, rain)).astype(np.float32)
y = is_risk

# 2. Entrenar RandomForest
clf = RandomForestClassifier(n_estimators=20, random_state=42)
clf.fit(X, y)

print(f"Modelo entrenado con éxito. Precisión: {clf.score(X, y) * 100:.2f}%")

# 3. Exportar a formato ONNX
initial_type = [('float_input', FloatTensorType([None, 4]))]
onnx_model = convert_sklearn(clf, initial_types=initial_type)

os.makedirs("../models", exist_ok=True)
output_path = "../models/climate_model.onnx"

with open(output_path, "wb") as f:
    f.write(onnx_model.SerializeToString())

print(f"¡Archivo generado con éxito en: {output_path}!")