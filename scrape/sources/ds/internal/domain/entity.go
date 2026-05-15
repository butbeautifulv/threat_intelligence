package domain

// Resource identifies a detections/signals artifact (Caldera, Sigma, etc.).
type Resource struct {
	Key    string
	Source string
	URL    string
	Kind   string
}
