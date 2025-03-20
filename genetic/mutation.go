package genetic

import (
	"image"
	"math/rand"
	"sync"

	"github.com/bishal0602/chaotic-canvas/utils"
	"github.com/fogleman/gg"
)

type AdaptiveMutationStrategy struct {
	baseMutationRate  float64
	diversityFactor   float64
	minMutationRate   float64
	maxMutationRate   float64
	populationHistory []float64 // Stores just average fitness values, not entire population
	historySize       int
	historyIndex      int // Index for circular buffer implementation
	plateauCounter    int // Counts generations with no improvement
	lastBestFitness   float64
}

func NewAdaptiveMutationStrategy(baseMutationRate float64) *AdaptiveMutationStrategy {
	return &AdaptiveMutationStrategy{
		baseMutationRate:  baseMutationRate,
		diversityFactor:   0.5,
		minMutationRate:   0.02,
		maxMutationRate:   0.2,
		historySize:       10,
		populationHistory: make([]float64, 10), // Pre-allocate with full size
		historyIndex:      0,
		plateauCounter:    0,
		lastBestFitness:   -1,
	}
}

// Update records the current generation's fitness and calculates the appropriate mutation rate
func (ams *AdaptiveMutationStrategy) Update(population []*Individual, generation, maxGenerations int) float64 {
	// Calculate current average fitness
	avgFitness := 0.0
	for _, ind := range population {
		avgFitness += ind.Fitness
	}
	avgFitness /= float64(len(population))

	// Store in history using circular buffer (more efficient than slice operations)
	ams.populationHistory[ams.historyIndex] = avgFitness
	ams.historyIndex = (ams.historyIndex + 1) % ams.historySize

	// Check for plateau
	bestFitness := population[0].Fitness
	if ams.lastBestFitness > 0 {
		if utils.Abs(bestFitness-ams.lastBestFitness) < 0.01 {
			ams.plateauCounter++
		} else {
			ams.plateauCounter = 0
		}
	}
	ams.lastBestFitness = bestFitness

	// Calculate progress
	stagnation := 0.0
	if generation >= ams.historySize {
		improvements := 0.0
		// Calculate improvements using circular buffer
		for i := 0; i < ams.historySize-1; i++ {
			current := (ams.historyIndex - 1 - i + ams.historySize) % ams.historySize
			previous := (current - 1 + ams.historySize) % ams.historySize

			// Inverted calculation since lower fitness is better
			relativeImprovement := (ams.populationHistory[current] - ams.populationHistory[previous]) / ams.populationHistory[previous]
			improvements += relativeImprovement
		}

		avgImprovement := improvements / float64(ams.historySize-1)

		// If improvement is very small, increase stagnation factor
		if avgImprovement < 0.001 {
			stagnation = 1.0 - (avgImprovement * 1000)
		}
	}

	// Calculate diversity - difference between best and average fitness
	diversity := 0.0
	if len(population) > 0 && avgFitness > 0 {
		// Inverted calculation since lower fitness is better
		diversity = (bestFitness - avgFitness) / bestFitness
	}

	// Calculate global progress (0.0 to 1.0)
	globalProgress := float64(generation) / float64(maxGenerations)

	// Calculate mutation rate based on factors:
	// 1. Higher when stagnating
	// 2. Higher when diversity is low
	// 3. Lower as we approach final generations
	// 4. Higher when stuck on a plateau
	mutationRate := ams.baseMutationRate
	mutationRate *= (1.0 + stagnation*2.0)                      // Increase by up to 3x when stagnating
	mutationRate *= (1.0 + (1.0-diversity)*ams.diversityFactor) // Increase when diversity is low
	mutationRate *= (1.0 - globalProgress*0.7)                  // Gradually reduce to 30% of initial rate

	// Increase mutation rate significantly if stuck on a plateau
	if ams.plateauCounter > 5 {
		plateauFactor := float64(ams.plateauCounter) / 10.0
		if plateauFactor > 2.0 {
			plateauFactor = 2.0
		}
		mutationRate *= (1.0 + plateauFactor)
	}

	// Mutation rate should not be strictly deterministic. Adding ±5% randomness
	mutationRate *= 1.0 + (rand.Float64()*0.1 - 0.05)

	return utils.Clamp(mutationRate, ams.minMutationRate, ams.maxMutationRate)
}

// Mutate creates a modified copy of the individual by adding random polygons.
func (ga *GeneticAlgorithm) Mutate(ind *Individual) *Individual {
	if rand.Float64() > ga.MutationRate {
		return ind
	}

	child := ind.CreateCopy()
	iterations := rand.Intn(3) + 1

	// Check if we should do a more radical mutation based on stagnation
	if rand.Float64() < ga.MutationRate*2 && ga.MutationRate > 0.1 {
		iterations += rand.Intn(3)
	}

	// Create a channel to collect polygons from goroutines
	polygonChan := make(chan Polygon, iterations)

	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			region := (child.Image.Bounds().Dx() + child.Image.Bounds().Dy())
			divisor := ga.MutationRate * float64(randomBetween(100, utils.FloorPowerOfTen(region)))
			region = region / int(utils.Max(divisor, 1)) // prevent divide by zero
			region = utils.Min(region, region>>2)

			numPoints := rand.Intn(4) + 3
			if ga.MutationRate > 0.1 {
				numPoints = rand.Intn(6) + 3 // Allow more complex polygons when stagnating
			}

			regionX := rand.Intn(child.Image.Bounds().Dx())
			regionY := rand.Intn(child.Image.Bounds().Dy())
			polygon := Polygon{
				Points: make([]image.Point, numPoints),
				Color:  utils.RandomRGBA(),
			}
			for j := 0; j < numPoints; j++ {
				x := utils.Clamp(regionX+rand.Intn(2*region)-region, 0, child.Image.Bounds().Dx()-1)
				y := utils.Clamp(regionY+rand.Intn(2*region)-region, 0, child.Image.Bounds().Dy()-1)
				polygon.Points[j] = image.Point{X: x, Y: y}
			}
			// Send the polygon to the main thread instead of modifying the slice directly
			polygonChan <- polygon
		}(i)
	}

	go func() {
		wg.Wait()
		close(polygonChan)
	}()

	for polygon := range polygonChan {
		child.Polygons = append(child.Polygons, polygon)
		dc := gg.NewContextForRGBA(child.Image)
		dc.SetRGBA255(int(polygon.Color.R), int(polygon.Color.G), int(polygon.Color.B), int(polygon.Color.A))
		for j, point := range polygon.Points {
			if j == 0 {
				dc.MoveTo(float64(point.X), float64(point.Y))
			} else {
				dc.LineTo(float64(point.X), float64(point.Y))
			}
		}
		dc.ClosePath()
		dc.Fill()
	}

	return child
}

func randomBetween(a, b int) int {
	if a > b {
		a, b = b, a // Swap if a > b to avoid errors
	}
	return rand.Intn(b-a+1) + a
}
