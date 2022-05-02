package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"image"
	"math"
	"math/rand"
	"os"
	"sync"
)

var TARGET image.Image

func init() {
	TARGET, _ = gg.LoadImage("images/darwin.png")
}

type Individual struct {
	elements []Polygon
	fitness  float64
}

func CxOnePoint(ind1 []Polygon, ind2 []Polygon) ([]Polygon, []Polygon) {
	size := len(ind1)
	if len(ind2) < size {
		size = len(ind2)
	}

	cxpoint := rand.Intn(size-2) + 1
	temp := ind1[cxpoint:]
	ind1 = append(ind1[:cxpoint], ind2[cxpoint:]...)
	ind2 = append(ind2[:cxpoint], temp...)
	return ind1, ind2
}

func BestSel(individuals []Individual) Individual {
	best := Individual{fitness: 0}
	for _, ind := range individuals {
		if ind.fitness > best.fitness {
			best = ind
		}
	}
	return best
}

func TournSel(individuals []Individual, k int, tournsize int) []Individual {
	var chosen []Individual
	for i := 0; i < k; i++ {
		var aspirant []Individual
		for j := 0; j < tournsize; j++ {
			aspirant = append(aspirant, individuals[rand.Intn(len(individuals))])
		}
		chosen = append(chosen, BestSel(aspirant))
	}

	return chosen
}

func MakeIndividual(n int) Individual {
	var polygons []Polygon
	for i := 0; i < n; i++ {
		polygons = append(polygons, MakePolygon())
	}
	ind := Individual{
		elements: polygons,
		fitness:  0,
	}
	return ind
}

func run(population []Individual, generations int) {
	for i := 0; i < generations; i++ {
		offspring := varOr(population, 120, 0.25, 0.4)

		var waitGroup sync.WaitGroup
		for i := 0; i < len(offspring); i++ {
			waitGroup.Add(1)
			i := i
			go func() {
				defer waitGroup.Done()
				offspring[i].fitness = evaluate(offspring[i].elements)
			}()
		}
		waitGroup.Wait()

		population = TournSel(offspring, len(population), 16)
		best := BestSel(population)
		img := DrawSolution(best.elements)
		err := img.SavePNG("solution.png")
		if err != nil {
			return
		}
		fmt.Printf("%d\t%f\t%d\n", i, best.fitness, len(best.elements))

	}
}

func varOr(population []Individual, lambda int, cxpb float64, mutpb float64) []Individual {
	if cxpb+mutpb > 1 {
		fmt.Println("cx and mutation probabilities must sum < 1")
		os.Exit(1)
	}

	var offspring []Individual
	for i := 0; i < lambda; i++ {
		choice := rand.Float64()
		if choice < cxpb { // Apply crossover
			idx1 := rand.Intn(len(population))
			var idx2 = idx1
			for idx2 == idx1 {
				idx2 = rand.Intn(len(population))
			}

			ind1 := make([]Polygon, len(population[idx1].elements))
			ind2 := make([]Polygon, len(population[idx2].elements))
			// Copy ind1
			for idx, elem := range population[idx1].elements {
				ind1[idx].color = elem.color
				for _, point := range elem.vertices {
					ind1[idx].vertices = append(ind1[idx].vertices, Point{point.x, point.y})
				}
			}
			// Copy ind2
			for idx, elem := range population[idx2].elements {
				ind2[idx].color = elem.color
				for _, point := range elem.vertices {
					ind2[idx].vertices = append(ind2[idx].vertices, Point{point.x, point.y})
				}
			}

			ind1, ind2 = CxOnePoint(ind1, ind2)
			if len(ind1) > 0 {
				offspring = append(offspring, Individual{elements: ind1, fitness: 0})
			}

		} else if choice < cxpb+mutpb { // apply mutation
			idx := rand.Intn(len(population))
			polygons := population[idx].elements
			var ind = make([]Polygon, len(polygons))
			// Copy ind1
			for idx2, elem := range population[idx].elements {
				ind[idx2].color = elem.color
				for _, point := range elem.vertices {
					ind[idx2].vertices = append(ind[idx2].vertices, Point{point.x, point.y})
				}
			}
			ind = mutate(ind)
			offspring = append(offspring, Individual{elements: ind, fitness: 0})
		} else { // Apply reproduction
			idx := rand.Intn(len(population))
			polygons := population[idx].elements
			var ind = make([]Polygon, len(polygons))
			for idx2, elem := range population[idx].elements {
				ind[idx2].color = elem.color
				for _, point := range elem.vertices {
					ind[idx2].vertices = append(ind[idx2].vertices, Point{point.x, point.y})
				}
			}
			offspring = append(offspring, Individual{elements: ind, fitness: 0})
		}

	}

	return offspring
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
		if len(solution) < 100 {
			solution = append(solution, MakePolygon())
		}
	}

	// Remove polygon
	if rand.Float64() < 0.2 {
		if len(solution) > 100/5 {
			idx := rand.Intn(len(solution) - 1)
			solution = append(solution[:idx], solution[idx+1:]...)
		}
	}

	// Add point to polygon
	if rand.Float64() < 0.2 {
		diff := 75
		choice := &solution[rand.Intn(len(solution))]
		center := choice.GetCenter()
		x := center.x + float64(rand.Intn(diff*2)-diff)
		y := center.y + float64(rand.Intn(diff*2)-diff)

		choice.AddPoint(Point{x, y})
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
	img := DrawSolution(solution)
	diff := ImageDifference(img.Image(), TARGET)
	MAX := float64(math.MaxUint16 * img.Height() * img.Width())
	return (MAX - diff) / MAX

}

func main() {
	var individuals []Individual
	for i := 0; i < 256; i++ {
		individuals = append(individuals, MakeIndividual(100))
	}

	DrawSolution(individuals[0].elements)

	run(individuals, 2000)

}
