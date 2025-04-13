package genetic

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/bishal0602/chaotic-canvas/mathutil"
	"github.com/fogleman/gg"
)

// Individual represents a candidate solution in the genetic algorithm
type Individual struct {
	Fitness float64
	Image   *image.RGBA
}

// Polygon represents a colored polygon
type Polygon struct {
	Points []image.Point
	Color  color.RGBA
}

// NewIndividual creates a new individual with random polygons
func NewIndividual(width, height int) *Individual {
	ind := &Individual{
		Fitness: math.Inf(1),
		Image:   image.NewRGBA(image.Rect(0, 0, width, height)),
	}

	// Create random background color
	bgColor := RandomRGBA()
	draw.Draw(ind.Image, ind.Image.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Add random polygons
	ind.createRandomPolygons()

	return ind
}

// CreateCopy creates a deep copy of the individual
func (ind *Individual) CreateCopy() *Individual {
	newInd := &Individual{
		Fitness: ind.Fitness,
		Image:   image.NewRGBA(image.Rect(0, 0, ind.Image.Bounds().Dx(), ind.Image.Bounds().Dy())),
	}

	// Copy image pixels
	draw.Draw(newInd.Image, newInd.Image.Bounds(), ind.Image, image.Point{}, draw.Src)

	return newInd
}

// createRandomPolygons creates random polygons for the individual
func (ind *Individual) createRandomPolygons() {
	numOfPoly := rand.Intn(5) + 3
	region := (ind.Image.Bounds().Dx() + ind.Image.Bounds().Dy()) / 8

	for i := 0; i < numOfPoly; i++ {
		numOfVertices := rand.Intn(4) + 3

		regionX := rand.Intn(ind.Image.Bounds().Dx())
		regionY := rand.Intn(ind.Image.Bounds().Dy())

		polygon := Polygon{
			Points: make([]image.Point, numOfVertices),
			Color:  RandomRGBA(),
		}

		// Generate random points for the polygon
		for j := 0; j < numOfVertices; j++ {
			x := mathutil.Clamp(regionX+rand.Intn(2*region)-region, 0, ind.Image.Bounds().Dx()-1)
			y := mathutil.Clamp(regionY+rand.Intn(2*region)-region, 0, ind.Image.Bounds().Dy()-1)
			polygon.Points[j] = image.Point{X: x, Y: y}
		}

		// Draw the polygon
		dc := gg.NewContextForRGBA(ind.Image)
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
}

// CalculateFitness calculates the fitness using parallel processing
func (ind *Individual) CalculateFitness(targetImage *image.RGBA) {
	bounds := targetImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	numGoroutines := runtime.GOMAXPROCS(0)

	// Divide work into chunks
	rowsPerGoroutine := height / numGoroutines
	differences := make([]float64, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		startY := i * rowsPerGoroutine
		endY := startY + rowsPerGoroutine
		if i == numGoroutines-1 {
			endY = height
		}

		go func(startY, endY, idx int) {
			defer wg.Done()
			differences[idx] = calculateRegionFitness(ind.Image, targetImage, startY, endY)
		}(startY, endY, i)
	}

	wg.Wait()

	// Sum up all differences
	var totalDifference float64
	for _, diff := range differences {
		totalDifference += diff
	}

	ind.Fitness = totalDifference / float64(width*height)
}

func calculateRegionFitness(img1, img2 *image.RGBA, startY, endY int) float64 {
	var difference float64
	width := img1.Bounds().Dx()

	for y := startY; y < endY; y++ {
		i := y * img1.Stride
		for x := 0; x < width; x++ {
			idx := i + x*4

			// Direct pixel access for better performance
			r1, g1, b1, a1 := img1.Pix[idx], img1.Pix[idx+1], img1.Pix[idx+2], img1.Pix[idx+3]
			r2, g2, b2, a2 := img2.Pix[idx], img2.Pix[idx+1], img2.Pix[idx+2], img2.Pix[idx+3]

			rDiff := float64(int(r1) - int(r2))
			gDiff := float64(int(g1) - int(g2))
			bDiff := float64(int(b1) - int(b2))
			aDiff := float64(int(a1) - int(a2))

			difference += math.Sqrt(rDiff*rDiff + gDiff*gDiff + bDiff*bDiff + aDiff*aDiff)
		}
	}

	return difference
}
