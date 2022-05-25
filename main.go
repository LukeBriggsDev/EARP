package main

import (
	"fmt"
	"github.com/ericpauley/go-quantize/quantize"
	"github.com/fogleman/gg"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
)

var TARGET image.Image
var BG rgba
var MAX_POLYGONS int
var GENERERATIONS int
var IMAGE_PATH string

var TARGET_WIDTH int
var TARGET_HEIGHT int

func getMostCommonColor(image image.Image, pallete color.Palette) color.Color {
	count := make([]int, len(pallete))
	for x := 0; x < image.Bounds().Max.X; x++ {
		for y := 0; y < image.Bounds().Max.Y; y++ {
			count[pallete.Index(image.At(x, y))] += 1
		}
	}
	bestIdx := 0
	bestVal := 0
	for idx, val := range count {
		if val > bestVal {
			bestVal = val
			bestIdx = idx
		}
	}
	fmt.Println(bestIdx)
	return pallete[bestIdx]
}

func printUsage() {
	fmt.Println("Usage: EARP target_image no_of_polygons no_of_gens")
	os.Exit(0)
}

func init() {
	if len(os.Args) < 4 {
		printUsage()
	}
	IMAGE_PATH = os.Args[1]
	polygons, polyErr := strconv.Atoi(os.Args[2])
	gens, gensErr := strconv.Atoi(os.Args[3])

	if polyErr != nil {
		fmt.Println("Incorrect argument for polygon count")
		printUsage()
	}

	if gensErr != nil {
		fmt.Println("Incorrect argument for generation count")
		printUsage()
	}

	MAX_POLYGONS = polygons
	GENERERATIONS = gens

	TARGET, _ = gg.LoadImage(IMAGE_PATH)
	TARGET_WIDTH = TARGET.Bounds().Max.X
	TARGET_HEIGHT = TARGET.Bounds().Max.Y
	q := quantize.MedianCutQuantizer{}
	pallette := q.Quantize(make([]color.Color, 0, 16), TARGET)
	r, g, b, a := getMostCommonColor(TARGET, pallette).RGBA()
	BG = rgba{
		uint8(r),
		uint8(g),
		uint8(b),
		uint8(a),
	}
}

type Individual struct {
	elements []Polygon
	fitness  float64
}

func makeIndividual(n int) Individual {
	var polygons []Polygon
	for i := 0; i < n; i++ {
		polygons = append(polygons, makePolygon())
	}
	ind := Individual{
		elements: polygons,
		fitness:  -1,
	}
	return ind
}

func run(population []Individual, generations int) {
	for i := 0; i < generations; i++ {
		offspring := varOr(population, 120, 0.25, 0.4)

		var waitGroup sync.WaitGroup
		for i := 0; i < len(offspring); i++ {
			if offspring[i].fitness == -1 {
				waitGroup.Add(1)
				i := i
				go func() {
					defer waitGroup.Done()
					offspring[i].fitness = evaluate(offspring[i].elements)
				}()
			}
		}
		waitGroup.Wait()

		population = TournSel(offspring, len(population), 16)
		best := BestSel(population)
		img := drawSolution(best.elements)
		err := img.SavePNG("solution.png")
		if err != nil {
			return
		}
		fmt.Printf("%d\t%f\t%d\n", i, best.fitness, len(best.elements))

	}
}

func uint8NoOverflow(i int) uint8 {
	return uint8(math.Max(0, math.Min(float64(i), math.MaxUint8)))
}

func mutate(solution []Polygon) []Polygon {
	// Mutate color
	if rand.Float64() < 0.4 {
		choice := &solution[rand.Intn(len(solution))]
		choice.color.r = uint8NoOverflow(int(choice.color.r) + int(30*rand.NormFloat64()))
		choice.color.g = uint8NoOverflow(int(choice.color.r) + int(30*rand.NormFloat64()))
		choice.color.b = uint8NoOverflow(int(choice.color.r) + int(30*rand.NormFloat64()))

	}

	// Mutate transparency
	if rand.Float64() < 0.5 {
		choice := &solution[rand.Intn(len(solution))]
		choice.color.a = uint8NoOverflow(int(choice.color.a) + int(30*rand.NormFloat64()))
	}

	// Add polygon
	if rand.Float64() < 0.3 {
		if len(solution) < MAX_POLYGONS {
			solution = append(solution, makePolygon())
		}
	}

	// Remove polygon
	if rand.Float64() < 0.2 {
		if len(solution) > MAX_POLYGONS/5 {
			idx := rand.Intn(len(solution) - 1)
			solution = append(solution[:idx], solution[idx+1:]...)
		}
	}

	// Add point to polygon
	if rand.Float64() < 0.2 {
		diff := 75
		choice := &solution[rand.Intn(len(solution))]
		center := choice.getCenter()
		x := center.x + float64(rand.Intn(diff*2)-diff)
		y := center.y + float64(rand.Intn(diff*2)-diff)

		choice.addPoint(Point{x, y})
	}

	// Re-order polygons
	if rand.Float64() < 0.3 {
		rand.Shuffle(len(solution), func(i, j int) { solution[i], solution[j] = solution[j], solution[i] })
	}

	// Mutate individual points
	if rand.Float64() < 0.7 {
		choice := &solution[rand.Intn(len(solution))]
		i := rand.Intn(len(choice.vertices))
		x := choice.vertices[i].x
		y := choice.vertices[i].y
		if rand.Float64() < 0.4 {
			x += 10 * rand.NormFloat64()
			x = math.Max(0, math.Min(x, float64(TARGET.Bounds().Max.X)))

		}
		if rand.Float64() < 0.4 {
			y += 10 * rand.NormFloat64()
			y = math.Max(0, math.Min(y, float64(TARGET.Bounds().Max.Y)))
		}
		choice.vertices[i] = Point{x, y}

	}

	// Mutate all points
	if rand.Float64() < 0.1 {
		choice := &solution[rand.Intn(len(solution))]
		for i := 0; i < len(choice.vertices); i++ {
			x := choice.vertices[i].x
			y := choice.vertices[i].y
			if rand.Float64() < 0.4 {
				x += 10 * rand.NormFloat64()
				x = math.Max(0, math.Min(x, float64(TARGET.Bounds().Max.X)))

			}
			if rand.Float64() < 0.4 {
				y += 10 * rand.NormFloat64()
				y = math.Max(0, math.Min(y, float64(TARGET.Bounds().Max.Y)))
			}
			choice.vertices[i] = Point{x, y}
		}
	}

	return solution
}

func evaluate(solution []Polygon) float64 {
	img := drawSolution(solution)
	diff := ImageDifference(img.Image(), TARGET)
	MAX := float64(math.MaxUint16 * img.Height() * img.Width())
	return (MAX - diff) / MAX

}

func main() {

	var individuals []Individual
	for i := 0; i < 256; i++ {
		individuals = append(individuals, makeIndividual(MAX_POLYGONS))
	}

	drawSolution(individuals[0].elements)

	run(individuals, GENERERATIONS)

}
