package internal

type SpeciesMetaData struct {
	Species string `json:"species"`
	Source string `json:"source"`
	Link string `json:"link"`
	Name string `json:"name"`
	ImagePath string `json:"image_path"`
	Summary string `json:"summary"`
}

type SpeciesFinder interface {
	FindMetaData(scientificName string) ([]SpeciesMetaData, error)
}