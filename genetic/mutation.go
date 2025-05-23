package genetic

import (
	"image"
	"math"
	"math/rand"
	"sync"

	"github.com/bishal0602/chaotic-canvas/mathutil"
	"github.com/fogleman/gg"
)

const (
	// Mutation History
	mutationHistorySize int = 10

	// Adaptive Strategy Parameters
	minMutationRateFloor     float64 = 0.01
	minMutationRateScale     float64 = 0.2
	maxMutationRateScale     float64 = 5.0
	maxMutationRateCeiling   float64 = 0.4
	defaultDiversityFactor   float64 = 0.5
	plateauFitnessThreshold  float64 = 0.01
	plateauDurationThreshold int     = 5

	// Mutation Application Parameters
	minMutationIterations              = 1
	maxMutationIterationsBase          = 3
	radicalMutationExtraIterations     = 3
	minPolygonPoints               int = 3
	maxPolygonPoints               int = 6
	highMutationExtraPoints        int = 2
)

// MutationHistory tracks fitness progress over time.
type MutationHistory struct {
	history      []float64
	size         int
	index        int
	lastBest     float64
	plateauCount int
}

// NewMutationHistory creates a history tracker.
func NewMutationHistory(size int) *MutationHistory {
	return &MutationHistory{
		history: make([]float64, size),
		size:    size,
	}
}

// Record stores fitness and detects plateaus.
func (mh *MutationHistory) Record(avgFitness, bestFitness float64) {
	mh.history[mh.index] = avgFitness
	mh.index = (mh.index + 1) % mh.size

	if mh.lastBest > 0 && mathutil.Abs(bestFitness-mh.lastBest) < plateauFitnessThreshold {
		mh.plateauCount++
	} else {
		mh.plateauCount = 0
	}
	mh.lastBest = bestFitness
}

// GetImprovementScore measures relative fitness progress.
func (mh *MutationHistory) GetImprovementScore() float64 {
	improvements := 0.0
	for i := 0; i < mh.size-1; i++ {
		cur := (mh.index - 1 - i + mh.size) % mh.size
		prev := (cur - 1 + mh.size) % mh.size
		improvements += (mh.history[cur] - mh.history[prev]) / mh.history[prev]
	}
	return improvements / float64(mh.size-1)
}

type AdaptiveMutationStrategy struct {
	baseRate float64
	minRate  float64
	maxRate  float64
	history  *MutationHistory
}

func NewAdaptiveMutationStrategy(baseMutationRate float64) *AdaptiveMutationStrategy {
	return &AdaptiveMutationStrategy{
		baseRate: baseMutationRate,
		minRate:  mathutil.Max(minMutationRateFloor, minMutationRateScale*baseMutationRate),
		maxRate:  mathutil.Min(maxMutationRateCeiling, maxMutationRateScale*baseMutationRate),
		history:  NewMutationHistory(mutationHistorySize),
	}
}

// Update records the current generation's fitness and calculates the appropriate mutation rate
func (ams *AdaptiveMutationStrategy) Update(pop []*Individual, gen, maxGen int) float64 {
	// Calculate current average fitness
	avgFitness := 0.0
	for _, ind := range pop {
		avgFitness += ind.Fitness
	}
	avgFitness /= float64(len(pop))
	bestFitness := pop[0].Fitness
	ams.history.Record(avgFitness, bestFitness)

	// Calculate stagnation (lack of progress)
	stagnation := 0.0
	if gen >= ams.history.size {
		improvement := ams.history.GetImprovementScore()
		// If improvement is very small, increase stagnation factor
		if improvement < 0.001 {
			stagnation = 1.0 - (improvement * 1000)
		}
	}

	// Calculate diversity as difference between best and average fitness
	diversity := 0.0
	if len(pop) > 0 && avgFitness > 0 {
		diversity = math.Abs(bestFitness-avgFitness) / bestFitness
	}

	progress := float64(gen) / float64(maxGen)

	return ams.computeMutationRate(stagnation, diversity, progress)
}

// MutationCache struct holds precomputed values
type MutationCache struct {
	LogSize    float64
	FloorPower int
	MaxLimit   int
}
type CacheManager struct {
	cache sync.Map // Thread-safe cache
}

var cacheManager = &CacheManager{}

func (cm *CacheManager) getMutationCache(region int) *MutationCache {
	if cached, ok := cm.cache.Load(region); ok {
		return cached.(*MutationCache)
	}

	logSize := 0.0
	if region > 1 {
		logSize = math.Log10(float64(region))
	} else {
		logSize = 1 // Prevent log(0) and logSize = 0
	}
	floorPower := mathutil.FloorPowerOfTen(region)
	maxLimit := region >> 4 // At most 1/16 of the image area
	cache := &MutationCache{
		LogSize:    logSize,
		FloorPower: floorPower,
		MaxLimit:   maxLimit,
	}
	cm.cache.Store(region, cache)

	return cache
}

// Mutate creates a modified copy of the individual by adding random polygons.
func (ga *GeneticAlgorithm) Mutate(ind *Individual) *Individual {
	if rand.Float64() > ga.MutationRate {
		return ind
	}

	child := ind.CreateCopy()
	iterations := func() int {
		it := mathutil.RandomBetween(minMutationIterations, maxMutationIterationsBase)
		// Check if we should do a more radical mutation based on stagnation
		if ga.MutationRate > 0.1 && rand.Float64() < ga.MutationRate*2 {
			it += rand.Intn(radicalMutationExtraIterations)
		}
		return it
	}()

	region := child.Image.Bounds().Dx() * child.Image.Bounds().Dy()
	// Retrieve precomputed mutation values from global cache
	cache := cacheManager.getMutationCache(region)
	maxLimit := cache.MaxLimit
	logSize := cache.LogSize
	floorPower := cache.FloorPower

	dc := gg.NewContextForRGBA(child.Image)

	for i := 0; i < iterations; i++ {
		// Randomly scale mutation size within a reasonable range
		scaleFactor := mathutil.RandomBetween(1, int(logSize*5))
		divisor := ga.MutationRate * float64(mathutil.RandomBetween(50, floorPower))
		regionLimit := (region / int(mathutil.Max(divisor, 1))) / scaleFactor
		regionLimit = mathutil.Clamp(regionLimit, 1, maxLimit)
		// fmt.Printf("scale: %v, divisor: %v, limit: %v, mut: %v\n", scaleFactor, divisor, regionLimit, ga.MutationRate)

		numPoints := func() int {
			n := mathutil.RandomBetween(minPolygonPoints, maxPolygonPoints)
			if ga.MutationRate > 0.1 {
				n += highMutationExtraPoints
			}
			return n
		}()

		regionX := rand.Intn(child.Image.Bounds().Dx())
		regionY := rand.Intn(child.Image.Bounds().Dy())

		polygon := Polygon{
			Points: make([]image.Point, numPoints),
			Color:  RandomRGBA(),
		}

		for j := 0; j < numPoints; j++ {
			x := mathutil.Clamp(regionX+rand.Intn(2*regionLimit)-regionLimit, 0, child.Image.Bounds().Dx()-1)
			y := mathutil.Clamp(regionY+rand.Intn(2*regionLimit)-regionLimit, 0, child.Image.Bounds().Dy()-1)
			polygon.Points[j] = image.Point{X: x, Y: y}
		}

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

// computeMutationRate adjusts mutation based on factors:
// 1. Higher when stagnating
// 2. Higher when diversity is low
// 3. Lower as we approach final generations
// 4. Higher when stuck on a plateau
func (ams *AdaptiveMutationStrategy) computeMutationRate(stagnation, diversity, progress float64) float64 {
	stagnationFactor := 1.0 + stagnation*2.0
	diversityFactor := 1.0 + (1.0-diversity)*0.5
	progressFactor := 1.0 - progress*0.7
	plateauFactor := 1.0

	// Increase mutation rate significantly if stuck on a plateau
	if ams.history.plateauCount > plateauDurationThreshold {
		factor := float64(ams.history.plateauCount) / 10.0
		if factor > 2.0 {
			factor = 2.0
		}
		plateauFactor = 1.0 + factor
	}

	// Mutation rate should not be strictly deterministic. Adding ±5% randomness
	randomFactor := 1.0 + (rand.Float64()*0.1 - 0.05)

	rate := ams.baseRate * stagnationFactor * diversityFactor * progressFactor * plateauFactor * randomFactor

	return mathutil.Clamp(rate, ams.minRate, ams.maxRate)
}
