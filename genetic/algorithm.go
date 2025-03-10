package genetic

import (
	"image"
	"image/draw"
	"math"
	"runtime"
	"sort"
)

// GeneticAlgorithm represents the genetic algorithm parameters and state
type GeneticAlgorithm struct {
	TargetRGBA     *image.RGBA
	PopulationSize int
	Generations    int
	MutationRate   float64
}

type ImageResult struct {
	Img          image.Image
	Generation   int
	Fitness      float64
	MutationRate float64
}

func NewGeneticAlgorithm(target image.Image, popSize, generations int, mutationRate float64) (*GeneticAlgorithm, error) {
	bounds := target.Bounds()
	targetRGBA := image.NewRGBA(bounds)
	draw.Draw(targetRGBA, bounds, target, bounds.Min, draw.Src)

	return &GeneticAlgorithm{
		TargetRGBA:     targetRGBA,
		PopulationSize: popSize,
		Generations:    generations,
		MutationRate:   mutationRate,
	}, nil
}

func (ga *GeneticAlgorithm) Run(recv chan<- ImageResult, recvEvery int) (*Individual, error) {
	// Initialize
	defer close(recv)
	adaptiveMutation := NewAdaptiveMutationStrategy(ga.MutationRate)
	population := make([]*Individual, ga.PopulationSize)
	for i := range population {
		population[i] = NewIndividual(ga.TargetRGBA.Bounds().Dx(), ga.TargetRGBA.Bounds().Dy())
		population[i].CalculateFitness(ga.TargetRGBA)
	}

	bestFitness := math.Inf(1)
	var bestIndividual *Individual

	for gen := 0; gen < ga.Generations; gen++ {
		ga.MutationRate = adaptiveMutation.Update(population, gen, ga.Generations)
		// Evolve the old population
		population = ga.evolvePopulation(population)
		currentBest := population[0]

		if currentBest.Fitness < bestFitness {
			bestFitness = currentBest.Fitness
			bestIndividual = currentBest.CreateCopy()
		}

		// Send progress periodically
		if gen%recvEvery == 0 || gen == ga.Generations-1 {
			recv <- ImageResult{
				Generation:   gen,
				Img:          bestIndividual.Image,
				Fitness:      bestFitness,
				MutationRate: ga.MutationRate,
			}
		}
	}

	return bestIndividual, nil
}

// evolvePopulation creates a new population by selecting parents and applying crossover and mutation
// The population is sorted by fitness, with fittest individuals appearing first.
func (ga *GeneticAlgorithm) evolvePopulation(population []*Individual) []*Individual {
	newPopulation := make([]*Individual, ga.PopulationSize)
	batchSize := (runtime.NumCPU() * 3) / 2 * 2 // Ensure even number
	if batchSize > ga.PopulationSize {
		batchSize = ga.PopulationSize - (ga.PopulationSize % 2) // Ensure even
	}

	for start := 0; start < ga.PopulationSize-1; start += batchSize {
		end := start + batchSize
		if end > ga.PopulationSize {
			end = ga.PopulationSize - (ga.PopulationSize % 2)
		}
		batchChan := make(chan struct {
			indices     [2]int
			individuals [2]*Individual
		}, (end-start)/2)

		for i := start; i < end; i += 2 {
			go func(idx int) {
				parent1 := TournamentSelect(population, 6)
				parent2 := TournamentSelect(population, 6)

				child1, child2 := ga.Crossover(parent1, parent2)
				child1 = ga.Mutate(child1)
				child2 = ga.Mutate(child2)
				child1.CalculateFitness(ga.TargetRGBA)
				child2.CalculateFitness(ga.TargetRGBA)

				// Select best two from children and parents
				var result [2]*Individual
				candidates := [4]*Individual{child1, child2, parent1, parent2}
				sort.Slice(candidates[:], func(i, j int) bool {
					return candidates[i].Fitness < candidates[j].Fitness
				})
				result[0] = candidates[0].CreateCopy()
				result[1] = candidates[1].CreateCopy()

				batchChan <- struct {
					indices     [2]int
					individuals [2]*Individual
				}{[2]int{idx, idx + 1}, result}
			}(i)
		}

		for i := start; i < end; i += 2 {
			result := <-batchChan
			newPopulation[result.indices[0]] = result.individuals[0]
			newPopulation[result.indices[1]] = result.individuals[1]
		}

		runtime.GC()
	}

	// Handle remaining odd population member if any
	if ga.PopulationSize%2 != 0 {
		newPopulation[ga.PopulationSize-1] = population[0].CreateCopy()
	}

	sort.Slice(newPopulation, func(i, j int) bool {
		return newPopulation[i].Fitness < newPopulation[j].Fitness
	})

	return newPopulation
}
