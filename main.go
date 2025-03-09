package main

import (
	"fmt"

	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bishal0602/chaotic-canvas/config"
	"github.com/bishal0602/chaotic-canvas/genetic"
	"github.com/bishal0602/chaotic-canvas/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	log.Printf("Starting image evolution with:\n")
	fmt.Printf("- Target image: %s\n", cfg.TargetImagePath)
	fmt.Printf("- Output Directory: %s\n", cfg.OutDir)
	fmt.Printf("- Population size: %d\n", cfg.PopulationSize)
	fmt.Printf("- Generations: %d\n", cfg.Generations)
	fmt.Printf("- Mutation rate: %.2f\n", cfg.MutationRate)

	algorithm, err := genetic.NewGeneticAlgorithm(cfg.TargetImagePath, cfg.OutDir, cfg.PopulationSize, cfg.Generations, cfg.MutationRate)
	if err != nil {
		log.Printf("Error initializing genetic algorithm: %v\n", err)
		os.Exit(1)
	}

	startTime := time.Now()
	bestIndividual, err := algorithm.Run()
	if err != nil {
		log.Printf("Error running genetic algorithm: %v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(startTime)

	// Save the final best individual
	outPath := filepath.Join(cfg.OutDir, "final_result.png")
	if err := utils.SaveImage(outPath, bestIndividual.Image); err != nil {
		log.Printf("Error saving final image: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Evolution completed in %v\n", elapsed)
	log.Printf("Final fitness: %.2f\n", bestIndividual.Fitness)
	log.Printf("Final image saved to: %s\n", outPath)
}
