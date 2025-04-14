package genetic

import (
	"image"
	"image/color"
	"testing"
)

func TestEvolutionMaintainsPopulationSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping evolution test in short mode")
	}
	// Create small test image
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))

	popSize := 49
	gen := 400
	ga, err := NewGeneticAlgorithm(img, popSize, gen, 0.05, 3)
	if err != nil {
		t.Fatalf("Failed to create GA: %v", err)
	}

	for range gen {
		newPop := ga.evolvePopulation(ga.Population)

		// Check population size remains constant
		if len(newPop) != popSize {
			t.Errorf("Population size changed: expected %d, got %d", popSize, len(newPop))
		}
	}

}

func TestEvolutionInvariants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping evolution test in short mode")
	}

	targetImg := createCheckerPattern(30, 30, 2)

	pop := 40
	gen := 100
	ga, _ := NewGeneticAlgorithm(targetImg, pop, gen, 0.05, 3)
	recv := make(chan ImageResult)
	go func() {
		for range recv {
		} // Just consume
	}()

	result, _ := ga.Run(recv, 5)

	// Population is sorted by fitness
	for i := 1; i < len(ga.Population); i++ {
		if ga.Population[i].Fitness < ga.Population[i-1].Fitness {
			t.Errorf("Population not sorted by fitness at index %d", i)
		}
	}

	// Returned fitness is the best fitness
	if result.Fitness != ga.Population[0].Fitness {
		t.Errorf("Best individual fitness mismatch: expected %f, got %f", result.Fitness, ga.Population[0].Fitness)
	}

	// Image dimensions are preserved
	if ga.Population[0].Image.Bounds().Dx() != targetImg.Bounds().Dx() ||
		ga.Population[0].Image.Bounds().Dy() != targetImg.Bounds().Dy() {
		t.Errorf(
			"Image dimensions mismatch: expected %dx%d, got %dx%d",
			targetImg.Bounds().Dx(),
			targetImg.Bounds().Dy(),
			ga.Population[0].Image.Bounds().Dx(),
			ga.Population[0].Image.Bounds().Dy(),
		)
	}
}

func TestEvolutionImprovesFitness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping evolution test in short mode")
	}

	targetImg := createCheckerPattern(30, 30, 2)

	pop := 40
	gen := 100
	ga, _ := NewGeneticAlgorithm(targetImg, pop, gen, 0.05, 3)
	initialBestFitness := ga.Population[0].Fitness
	recv := make(chan ImageResult)
	go func() {
		for range recv {
		} // Just consume
	}()

	result, _ := ga.Run(recv, 5)

	// Final fitness should improve (decrease)
	if result.Fitness >= initialBestFitness {
		t.Errorf("Warning: Evolution didn't improve fitness in test run")
	}
}

func TestIdenticalImageZeroFitness(t *testing.T) {
	img1 := createCheckerPattern(50, 50, 2)
	img2 := createCheckerPattern(50, 50, 2)

	ind := &Individual{Image: img2}
	ind.CalculateFitness(img1)
	if ind.Fitness != 0 {
		t.Errorf("Expected fitness 0 for identical images, got %f", ind.Fitness)
	}
}

func createCheckerPattern(width, height, size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	colorA := color.RGBA{180, 50, 90, 255}  // muted magenta
	colorB := color.RGBA{60, 200, 180, 255} // teal-green

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			xCell := x / size
			yCell := y / size
			if (xCell+yCell)%2 == 0 {
				img.Set(x, y, colorA)
			} else {
				img.Set(x, y, colorB)
			}
		}
	}

	return img
}
