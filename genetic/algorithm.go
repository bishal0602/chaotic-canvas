package genetic

import (
	"errors"
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
	TournamentSize int
	Population     []*Individual
}

type ImageResult struct {
	Img          image.Image
	Generation   int
	Fitness      float64
	MutationRate float64
}

func NewGeneticAlgorithm(target image.Image, popSize, generations int, mutationRate float64, tournamentSize int) (*GeneticAlgorithm, error) {
	if target == nil || popSize <= 0 || generations <= 0 || mutationRate < 0 || mutationRate > 1 || tournamentSize <= 0 {
		return nil, errors.New("invalid parameters for genetic algorithm")
	}

	bounds := target.Bounds()
	targetRGBA := image.NewRGBA(bounds)
	draw.Draw(targetRGBA, bounds, target, bounds.Min, draw.Src)

	population := make([]*Individual, popSize)
	for i := range population {
		population[i] = NewIndividual(targetRGBA.Bounds().Dx(), targetRGBA.Bounds().Dy())
		population[i].CalculateFitness(targetRGBA)
	}
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness < population[j].Fitness
	})

	return &GeneticAlgorithm{
		TargetRGBA:     targetRGBA,
		PopulationSize: popSize,
		Generations:    generations,
		MutationRate:   mutationRate,
		TournamentSize: tournamentSize,
		Population:     population,
	}, nil
}

func (ga *GeneticAlgorithm) Run(recv chan<- ImageResult, recvEvery int) (*Individual, error) {
	// Initialize
	defer close(recv)
	mutationStrategy := NewAdaptiveMutationStrategy(ga.MutationRate)

	bestFitness := math.Inf(1)
	var bestIndividual *Individual

	for gen := 0; gen < ga.Generations; gen++ {
		ga.MutationRate = mutationStrategy.Update(ga.Population, gen, ga.Generations)
		// Evolve the old population
		newPopulation := ga.evolvePopulation(ga.Population)
		currentBest := newPopulation[0]
		ga.Population = newPopulation

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
				parent1 := TournamentSelect(population, ga.TournamentSize)
				parent2 := TournamentSelect(population, ga.TournamentSize)

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
