package constants

const (
	Carnivore  = 0
	Omnivore   = 1
	Herbivore  = 2
)

var SelectedDiet int // Global variable to store the selected diet

func SetSelectedDiet(diet int) {
	SelectedDiet = diet
}