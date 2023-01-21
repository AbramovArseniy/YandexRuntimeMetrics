package server

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MemStorage struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]float64
}

type Server struct {
	handler     Handler
	FileHandler FileHandler
}

func NewServer() *Server {
	return &Server{
		handler: Handler{
			storage: MemStorage{
				CounterMetrics: make(map[string]int64),
				GaugeMetrics:   make(map[string]float64),
			},
		},
		FileHandler: FileHandler{},
	}
}
