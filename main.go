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

	img, err := utils.ReadImage(cfg.TargetImagePath)
	if err != nil {
		log.Fatalf("error reading target image: %v", err)
	}
	// Create output directory for images
	if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
		log.Fatalf("error creating output directory: %v", err)
	}

	recv := make(chan genetic.ImageResult)
	recvEvery := 100
	go func() {
		for result := range recv {
			outPath := filepath.Join(cfg.OutDir, fmt.Sprintf("best_gen_%d.png", result.Generation))
			if err := utils.SaveImage(outPath, result.Img); err != nil {
				log.Printf("Error saving image (gen %d): %v\n", result.Generation, err)
			} else {
				log.Printf("Generation %d - Best fitness: %.2f - Mutation Rate: %.2f", result.Generation, result.Fitness, result.MutationRate)

			}
		}
	}()

	algorithm, err := genetic.NewGeneticAlgorithm(img, cfg.PopulationSize, cfg.Generations, cfg.MutationRate)
	if err != nil {
		log.Fatalf("Error initializing genetic algorithm: %v\n", err)
	}

	startTime := time.Now()
	bestIndividual, err := algorithm.Run(recv, recvEvery)
	if err != nil {
		log.Fatalf("Error running genetic algorithm: %v\n", err)
	}
	elapsed := time.Since(startTime)

	// Save the final best individual
	outPath := filepath.Join(cfg.OutDir, "final_result.png")
	if err := utils.SaveImage(outPath, bestIndividual.Image); err != nil {
		log.Fatalf("Error saving final image: %v\n", err)
	}

	log.Printf("Evolution completed in %v\n", elapsed)
	log.Printf("Final fitness: %.2f\n", bestIndividual.Fitness)
	log.Printf("Final image saved to: %s\n", outPath)
}
