/*
This is a sample Go program that works in conjuction
with http://byo-game-of-life.herokuapp.com/ as a back-end.
This example was used in the Indy Golang meetup of Jan 6, 2015.
Rewritten by Laszlo Szenes from Larry Price's example given during the meetup .
*/

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

//variable types used to represent the API and the internal representation of the
//game of life playing field

//coordinates that are used for parsing the input, encoding the output and indexing the
//table representation
type coord struct {
	Column int `json:"column"`
	Row    int `json:"row"`
}

//type for the input and output parameters
//which are an array(list) of coordinates of the alive cells
type inpOut []coord

//Type for one element of the table representation of the playing field
type tableElem struct {
	alive  bool //indicates whether there's an alive cell in the element
	neighb byte //the number of live cells around the element
}

//The table representation of the playing field. It's a map, indexed by coordinates
//a map type is used in order to allow for an arbitrarily large playing field.
//the table only contains entries for live cells, and their surrounding locations
//because those are the only ones that matter for the calculation of the next generation
type table map[coord]tableElem

//calculating the next generation
//input: list of coordinates where the alive cells are
//output: list of coordinates of the alive cells in the next generation
func nextGen(cells inpOut) inpOut {
//note: type inpOut is a slice, which is passed by reference so no need to use pointers 
	
	T := make(table) //creating an empty table representing the playing field

	//iterating over the list of alive cells and building a map (table) of cells and neighbor counts
	for _, cc := range cells { //we discard the index, only need the current element's coordinates

		//Marking the element as `alive`
		//T[cc] might or might not exists at this point but
		//by reading it from it we get an initialized instance of tableElem
		//if it didn't exist before
		actCell := T[cc]
		actCell.alive = true
		T[cc] = actCell

		//iterating over the neighbooring squares and adding 1 to the neighbor count
		for x := -1; x < 2; x++ {
			for y := -1; y < 2; y++ {
				if x == 0 && y == 0 {
					continue
				} //skipping the current element

				//generating the coordinate index for the neighboring element
				//and incrementing its neighbor count
				newCoord := coord{Column: cc.Column + x, Row: cc.Row + y}
				temp := T[newCoord]
				temp.neighb++
				T[newCoord] = temp
			}
		}
	}

	//creating the array of the list of coordinates of the alive cells in the next generation
	//The array (more exactly it's a slice) has an initial size of 0, and a capacity
	//equal to the number of cells in the previous generation.
	//This is done to minimize the unnecessary copy operations when an `append` operation
	//exceeds the capacity of the slice
	out := make(inpOut, 0, len(cells))

	//iterating over the cellmap and generating a list of alive cells of the next generation
	for cc, elem := range T {
		//The rule for new cells - if it has exactly 3 neighbors
		//and surviving cells - if it has 2 or 3 neighbors
		if elem.neighb == 3 || (elem.neighb == 2 && elem.alive) {
			//adding the cell to the list of the next generation cells
			out = append(out, cc)
		}
	}
	return out
}

//The HTTP API handler
func handler(w http.ResponseWriter, r *http.Request) {
	//setting a header to disable the cross site scripting block of the browser
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//parse the input parameters we get from the front end (GET)
	r.ParseForm()

	//converting the JSON input into our own data structure
	var cellsList inpOut
	json.Unmarshal([]byte(r.FormValue("cells")), &cellsList)

	//running through `steps`(one of the input parameters) number of generations
	for steps, _ := strconv.Atoi(r.FormValue("steps")); steps > 0; steps-- {
		cellsList = nextGen(cellsList)
	}

	//generating the JSON response
	sendBack, _ := json.Marshal(cellsList)

	//sending the JSON response
	//String conversion is necessery because json.Marshal generates a byte array
	fmt.Fprint(w, string(sendBack))
}

//The `main` entry point of the program
func main() {
	//define the only route that we implement
	http.HandleFunc("/", handler)
	//listen on port 8080
	http.ListenAndServe(":8080", nil)
}
