package genetic

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/bishal0602/chaotic-canvas/mathutil"
)

const (
	blendCrossoverThreshold    = 0.3
	pointCrossoverThreshold    = 0.7 // blend + point
	gaussianCrossoverThreshold = 0.9 // blend + point + gaussian

	patchCrossoverSwapProbability = 0.3
	patchSize                     = 8

	gaussianNoiseScale = 0.1
)

func (ga *GeneticAlgorithm) Crossover(parent1 *Individual, parent2 *Individual) (*Individual, *Individual) {
	var child1, child2 *Individual

	r := rand.Float64()
	if r < blendCrossoverThreshold {
		child1, child2 = blendCrossover(parent1, parent2)
	} else if r < pointCrossoverThreshold {
		child1, child2 = crossoverPoint(parent1, parent2)
	} else if r < gaussianCrossoverThreshold {
		child1, child2 = gaussianPerturbationCrossover(parent1, parent2)
	} else {
		child1, child2 = patchCrossover(parent1, parent2)
	}
	return child1, child2
}

// blendCrossover performs a blend crossover operation between two parent individuals.
// It creates two children by interpolating pixel values between parents using a random alpha value.
func blendCrossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	child1 := parent1.CreateBlankCopy()
	child2 := parent2.CreateBlankCopy()

	bounds := child1.Image.Bounds()
	height := bounds.Dy()
	numGoroutines := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup

	// Process image in parallel strips
	rowsPerGoroutine := height / numGoroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			blendAlpha := rand.Float64()
			bounds := child1.Image.Bounds()
			for y := startY; y < endY; y++ {
				i := y * child1.Image.Stride
				for x := 0; x < bounds.Dx(); x++ {
					idx := i + x*4
					// Process both children in the same loop
					for j := 0; j < 4; j++ {
						p1 := float64(parent1.Image.Pix[idx+j])
						p2 := float64(parent2.Image.Pix[idx+j])
						child1.Image.Pix[idx+j] = uint8(p1*(1-blendAlpha) + p2*blendAlpha)
						child2.Image.Pix[idx+j] = uint8(p1*blendAlpha + p2*(1-blendAlpha))
					}
				}
			}
		}(i*rowsPerGoroutine, (i+1)*rowsPerGoroutine)
	}

	wg.Wait()
	return child1, child2
}

// crossoverPoint performs a single-point crossover between two parent individuals.
// It randomly chooses either a horizontal or vertical split point and creates two children
// by combining sections from both parents. The split can be either:
//   - Horizontal: Takes upper portion from parent1 and lower portion from parent2 for child1
//     Takes upper portion from parent2 and lower portion from parent1 for child2
//   - Vertical: Takes left portion from parent1 and right portion from parent2 for child1
//     Takes left portion from parent2 and right portion from parent1 for child2
func crossoverPoint(parent1, parent2 *Individual) (*Individual, *Individual) {
	child1 := parent1.CreateBlankCopy()
	child2 := parent2.CreateBlankCopy()

	isHorizontal := rand.Float64() <= 0.5
	bounds := child1.Image.Bounds()
	stride := child1.Image.Stride

	if isHorizontal {
		splitPoint := rand.Intn(bounds.Dy()-1) + 1
		// Child 1: upper from parent1, lower from parent2
		copy(child1.Image.Pix[:splitPoint*stride], parent1.Image.Pix[:splitPoint*stride])
		copy(child1.Image.Pix[splitPoint*stride:], parent2.Image.Pix[splitPoint*stride:])
		// Child 2: upper from parent2, lower from parent1
		copy(child2.Image.Pix[:splitPoint*stride], parent2.Image.Pix[:splitPoint*stride])
		copy(child2.Image.Pix[splitPoint*stride:], parent1.Image.Pix[splitPoint*stride:])
	} else {
		splitPoint := rand.Intn(bounds.Dx()-1) + 1
		for y := 0; y < bounds.Dy(); y++ {
			i := y * stride
			// Child 1: left from parent1, right from parent2
			copy(child1.Image.Pix[i:i+splitPoint*4], parent1.Image.Pix[i:i+splitPoint*4])
			copy(child1.Image.Pix[i+splitPoint*4:i+stride], parent2.Image.Pix[i+splitPoint*4:i+stride])
			// Child 2: left from parent2, right from parent1
			copy(child2.Image.Pix[i:i+splitPoint*4], parent2.Image.Pix[i:i+splitPoint*4])
			copy(child2.Image.Pix[i+splitPoint*4:i+stride], parent1.Image.Pix[i+splitPoint*4:i+stride])
		}
	}

	return child1, child2
}

// gaussianPerturbationCrossover creates two children by averaging the pixel values of both parents
// and then adding/subtracting a small amount of Gaussian noise to create variation.
// This method is useful for making subtle changes while preserving the overall image structure:
// - Child1 receives the parents' average pixel values plus small Gaussian noise
// - Child2 receives the parents' average pixel values minus small Gaussian noise
// The results are clamped to ensure valid pixel values (0-255)
func gaussianPerturbationCrossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	child1 := parent1.CreateBlankCopy()
	child2 := parent2.CreateBlankCopy()

	bounds := child1.Image.Bounds()

	for y := 0; y < bounds.Dy(); y++ {
		i := y * child1.Image.Stride

		noise := rand.Float64() * gaussianNoiseScale // Small Gaussian noise

		for x := 0; x < bounds.Dx(); x++ {
			idx := i + x*4

			// using bit-wise operation instead of loops per channel is ~11% faster with same allocs
			// Access the entire 4-byte pixel for both parents at once
			p1 := uint32(parent1.Image.Pix[idx]) | uint32(parent1.Image.Pix[idx+1])<<8 | uint32(parent1.Image.Pix[idx+2])<<16 | uint32(parent1.Image.Pix[idx+3])<<24
			p2 := uint32(parent2.Image.Pix[idx]) | uint32(parent2.Image.Pix[idx+1])<<8 | uint32(parent2.Image.Pix[idx+2])<<16 | uint32(parent2.Image.Pix[idx+3])<<24

			mean := (p1 + p2) / 2

			// Extract individual channels from the mean
			meanR := uint8(mean & 0xFF)
			meanG := uint8((mean >> 8) & 0xFF)
			meanB := uint8((mean >> 16) & 0xFF)
			meanA := uint8((mean >> 24) & 0xFF)

			// Apply Gaussian noise
			child1.Image.Pix[idx] = uint8(mathutil.Clamp(float64(meanR)+noise, 0, 255))
			child1.Image.Pix[idx+1] = uint8(mathutil.Clamp(float64(meanG)+noise, 0, 255))
			child1.Image.Pix[idx+2] = uint8(mathutil.Clamp(float64(meanB)+noise, 0, 255))
			child1.Image.Pix[idx+3] = uint8(mathutil.Clamp(float64(meanA)+noise, 0, 255))

			child2.Image.Pix[idx] = uint8(mathutil.Clamp(float64(meanR)-noise, 0, 255))
			child2.Image.Pix[idx+1] = uint8(mathutil.Clamp(float64(meanG)-noise, 0, 255))
			child2.Image.Pix[idx+2] = uint8(mathutil.Clamp(float64(meanB)-noise, 0, 255))
			child2.Image.Pix[idx+3] = uint8(mathutil.Clamp(float64(meanA)-noise, 0, 255))
		}
	}

	return child1, child2
}

// patchCrossover creates two children by swapping rectangular patches between the parents.
// It works by:
// - Creating exact copies of both parents
// - Dividing the image into small patches of patchSize
// - For each patch, having a % chance to swap that patch between the children
// This method preserves local structure within patches while creating diversity
// by recombining different regions from both parents.
func patchCrossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	child1 := parent1.CreateCopy()
	child2 := parent2.CreateCopy()

	bounds := child1.Image.Bounds()

	for y := 0; y < bounds.Dy(); y += patchSize {
		for x := 0; x < bounds.Dx(); x += patchSize {
			if rand.Float64() < patchCrossoverSwapProbability {
				for dy := 0; dy < patchSize && (y+dy) < bounds.Dy(); dy++ {
					for dx := 0; dx < patchSize && (x+dx) < bounds.Dx(); dx++ {
						idx := ((y+dy)*bounds.Dx() + (x + dx)) * 4
						copy(child1.Image.Pix[idx:idx+4], parent2.Image.Pix[idx:idx+4])
						copy(child2.Image.Pix[idx:idx+4], parent1.Image.Pix[idx:idx+4])
					}
				}
			}
		}
	}

	return child1, child2
}
