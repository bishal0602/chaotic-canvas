package main

import (
	"fmt"

	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bishal0602/chaotic-canvas/config"
	"github.com/bishal0602/chaotic-canvas/genetic"
	"github.com/bishal0602/chaotic-canvas/imageio"
)

const (
	compressedImageDimension       int = 540
	defaultProgressUpdateFrequency int = 100
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	log.Printf(`Starting image evolution with:
- Target image: %s
- Output Directory: %s
- Population size: %d
- Generations: %d
- Mutation rate: %.2f
- Tournament size: %d
- Compress: %t`,
		cfg.TargetImagePath, cfg.OutDir, cfg.PopulationSize, cfg.Generations, cfg.MutationRate, cfg.TournamentSize, !cfg.NoCompress,
	)

	img, err := imageio.Read(cfg.TargetImagePath)
	if err != nil {
		log.Fatalf("error reading target image: %v", err)
	}
	if !cfg.NoCompress {
		img = imageio.Resize(img, compressedImageDimension)
	}
	// Create output directory for images
	if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
		log.Fatalf("error creating output directory: %v", err)
	}

	recv := make(chan genetic.ImageResult)
	go func() {
		for result := range recv {
			outPath := filepath.Join(cfg.OutDir, fmt.Sprintf("best_gen_%d.png", result.Generation))
			if err := imageio.Save(outPath, result.Img); err != nil {
				log.Printf("Error saving image (gen %d): %v\n", result.Generation, err)
			} else {
				log.Printf("Generation %d - Best fitness: %.2f - Mutation Rate: %.2f", result.Generation, result.Fitness, result.MutationRate)

			}
		}
	}()

	algorithm, err := genetic.NewGeneticAlgorithm(img, cfg.PopulationSize, cfg.Generations, cfg.MutationRate, cfg.TournamentSize)
	if err != nil {
		log.Fatalf("Error initializing genetic algorithm: %v\n", err)
	}

	startTime := time.Now()
	bestIndividual, err := algorithm.Run(recv, defaultProgressUpdateFrequency)
	if err != nil {
		log.Fatalf("Error running genetic algorithm: %v\n", err)
	}
	elapsed := time.Since(startTime)

	// Save the final best individual
	outPath := filepath.Join(cfg.OutDir, "final_result.png")
	if err := imageio.Save(outPath, bestIndividual.Image); err != nil {
		log.Fatalf("Error saving final image: %v\n", err)
	}

	log.Printf("Evolution completed in %v\n", elapsed)
	log.Printf("Final fitness: %.2f\n", bestIndividual.Fitness)
	log.Printf("Final image saved to: %s\n", outPath)
}
