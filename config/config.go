package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	TargetImagePath string
	OutDir          string
	PopulationSize  int
	Generations     int
	MutationRate    float64
	TournamentSize  int
	NoCompress      bool
	EnablePprof     bool
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.TargetImagePath, "target", "examples/afghan_girl.png", "Path to target image")
	flag.StringVar(&cfg.OutDir, "out", "output", "Output Directory")
	flag.IntVar(&cfg.PopulationSize, "pop", 500, "Population size")
	flag.IntVar(&cfg.Generations, "gen", 10000, "Number of generations")
	flag.Float64Var(&cfg.MutationRate, "mut", 0.05, "Mutation rate")
	flag.IntVar(&cfg.TournamentSize, "tour", 6, "Tournament selection size")
	flag.BoolVar(&cfg.NoCompress, "nocompress", false, "Switch to disable compress")
	flag.BoolVar(&cfg.EnablePprof, "pprof", false, "Enable pprof profiling")

	flag.Parse()

	// Validation
	if cfg.TargetImagePath == "" {
		return nil, fmt.Errorf("target image path cannot be empty")
	}
	if _, err := os.Stat(cfg.TargetImagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("target image file not found: %s", cfg.TargetImagePath)
	}

	if cfg.OutDir == "" {
		return nil, fmt.Errorf("output directory cannot be empty")
	}

	if cfg.PopulationSize <= 0 {
		return nil, fmt.Errorf("population size must be positive, got %d", cfg.PopulationSize)
	}

	if cfg.Generations <= 0 {
		return nil, fmt.Errorf("number of generations must be positive, got %d", cfg.Generations)
	}

	if cfg.MutationRate < 0.0 || cfg.MutationRate > 1.0 {
		return nil, fmt.Errorf("mutation rate must be between 0.0 and 1.0, got %f", cfg.MutationRate)
	}

	if cfg.TournamentSize <= 0 {
		return nil, fmt.Errorf("tournament size must be positive, got %d", cfg.TournamentSize)
	}

	if cfg.TournamentSize > cfg.PopulationSize {
		return nil, fmt.Errorf("tournament size (%d) cannot be larger than population size (%d)", cfg.TournamentSize, cfg.PopulationSize)
	}

	return cfg, nil
}
