package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"math"
	"math/rand"
	"os"
	"sync"
)

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
	temp := ind1[:cxpoint]
	ind1 = append(ind1[cxpoint:], ind2[:cxpoint]...)
	ind2 = append(ind2[cxpoint:], temp...)
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

			copy(ind1, population[idx1].elements)
			copy(ind2, population[idx2].elements)
			ind1, ind2 = CxOnePoint(ind1, ind2)
			if len(ind1) > 0 {
				offspring = append(offspring, Individual{elements: ind1, fitness: 0})
			}

		} else if choice < cxpb+mutpb { // apply mutation
			polygons := population[rand.Intn(len(population))].elements
			var ind = make([]Polygon, len(polygons))
			copy(ind, polygons)
			ind = mutate(ind)
			offspring = append(offspring, Individual{elements: ind, fitness: 0})
		} else { // Apply reproduction
			polygons := population[rand.Intn(len(population))].elements
			var ind = make([]Polygon, len(polygons))
			copy(ind, polygons)
			offspring = append(offspring, Individual{elements: ind, fitness: 0})
		}

	}

	return offspring
}

func mutate(solution []Polygon) []Polygon {
	// Mutate color
	if rand.Float64() < 0.4 {
		choice := &solution[rand.Intn(len(solution))]
		choice.color.r += uint8(rand.Intn(30) - 15)
		choice.color.g += uint8(rand.Intn(30) - 15)
		choice.color.b += uint8(rand.Intn(30) - 15)

	}

	// Mutate transparency
	if rand.Float64() < 0.5 {
		choice := &solution[rand.Intn(len(solution))]
		choice.color.a += uint8(rand.Intn(30) - 15)
	}

	// Add polygon
	if rand.Float64() < 0.3 {
		if len(solution) < 100 {
			solution = append(solution, MakePolygon())
		}
	}

	return solution
}

func evaluate(solution []Polygon) float64 {
	target, _ := gg.LoadImage("images/darwin.png")
	image := DrawSolution(solution)
	diff := ImageDifference(image.Image(), target)
	MAX := float64(math.MaxUint16 * image.Height() * image.Width())
	return (MAX - diff) / MAX

}

func main() {
	var individuals []Individual
	for i := 0; i < 256; i++ {
		individuals = append(individuals, MakeIndividual(10))
	}

	DrawSolution(individuals[0].elements)

	run(individuals, 1000)

}
