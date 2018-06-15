package domain

type GenerationResult struct {
	Data Data `json:"data"`
}

type GenerationSeriesResult struct {
	Data []Data `json:"data"`
}

type Data struct {
	From string          `json:"from"`
	To   string          `json:"to"`
	Mix  Generationmixes `json:"generationmix"`
}

type Generationmix struct {
	Fuel       string  `json:"fuel"`
	Percentage float32 `json:"perc"`
}

type Generationmixes []Generationmix

// AggregateGreenEnergy calculates green energy percentage
func (g Generationmixes) AggregateGreenEnergy() (res float32) {
	for _, element := range g {
		switch element.Fuel {
		case "solar", "hydro", "wind":
			res += element.Percentage
		}
	}
	return res
}
