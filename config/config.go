package config

import "flag"

type Config struct {
	TargetImagePath string
	OutDir          string
	PopulationSize  int
	Generations     int
	MutationRate    float64
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.TargetImagePath, "target", "examples/starry_night.png", "Path to target image")
	flag.StringVar(&cfg.OutDir, "out", "output", "Output Directory")
	flag.IntVar(&cfg.PopulationSize, "pop", 500, "Population size")
	flag.IntVar(&cfg.Generations, "gen", 7000, "Number of generations")
	flag.Float64Var(&cfg.MutationRate, "mut", 0.02, "Mutation rate")

	flag.Parse()
	return cfg, nil
}
