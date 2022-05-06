package main

import (
	"fmt"
	"math/rand"
	"os"
)

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

func clonePolygonSlice(ind *[]Polygon) []Polygon {
	newInd := make([]Polygon, len(*ind))
	for idx, elem := range *ind {
		newInd[idx].color = elem.color
		for _, point := range elem.vertices {
			newInd[idx].vertices = append(newInd[idx].vertices, Point{point.x, point.y})
		}
	}
	return newInd
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

			ind1 := clonePolygonSlice(&population[idx1].elements)
			ind2 := clonePolygonSlice(&population[idx2].elements)

			ind1, ind2 = CxOnePoint(ind1, ind2)
			if len(ind1) > 0 {
				offspring = append(offspring, Individual{elements: ind1, fitness: -1})
			}

		} else if choice < cxpb+mutpb { // apply mutation
			ind := clonePolygonSlice(&population[rand.Intn(len(population))].elements)
			ind = mutate(ind)
			offspring = append(offspring, Individual{elements: ind, fitness: -1})
		} else { // Apply reproduction
			idx := rand.Intn(len(population))
			ind := clonePolygonSlice(&population[idx].elements)
			offspring = append(offspring, Individual{elements: ind, fitness: population[idx].fitness})
		}

	}

	return offspring
}
