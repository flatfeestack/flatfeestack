package main

// Data wraps the "data" JSON
type Data struct {
	Datasets []Dataset `json:"datasets"`
	Labels   []string  `json:"labels"`
}

type Dataset struct {
	Label string    `json:"label,omitempty"`
	Data  []float64 `json:"data"`
	Fill  bool      `json:"fill,omitempty"`
	//https://www.chartjs.org/docs/latest/configuration/elements.html
	BackgroundColor  string `json:"backgroundColor"`
	BorderColor      string `json:"borderColor"`
	PointBorderWidth int    `json:"pointBorderWidth"`
}
