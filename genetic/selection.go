package genetic

import (
	"math/rand"
)

const (
	numTournaments = 4 // Number of mini-tournaments to run in TournamentSelect
)

func TournamentSelect(population []*Individual, tournamentSize int) *Individual {
	var best *Individual

	for i := 0; i < numTournaments; i++ {
		tournamentBest := population[rand.Intn(len(population))]

		for j := 1; j < tournamentSize; j++ {
			participant := population[rand.Intn(len(population))]
			if participant.Fitness < tournamentBest.Fitness {
				tournamentBest = participant
			}
		}

		if best == nil || tournamentBest.Fitness < best.Fitness {
			best = tournamentBest
		}
	}

	return best
}
