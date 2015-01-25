package main

import (
    "log"
    "net/http"
    "net/url"
    "encoding/json"
    "strconv"
)

type cell struct {
    Row int `json:"row"`
    Column int `json:"column"`
}

type queryArguments struct {
    Cells []cell `json:"cells"`
    Generations int `json:"steps"`
}

func NewQueryArguments(query url.Values) queryArguments {
    generations, _ := strconv.Atoi(query["steps"][0])
    cells := []cell{}
    json.Unmarshal([]byte(query["cells"][0]), &cells)

    return queryArguments{cells, generations}
}

func buildGenerationMap(cells []cell) map[int]map[int]bool {
    generationMap := map[int]map[int]bool{}
    for _, c := range cells {
        if _, ok := generationMap[c.Row]; !ok {
            generationMap[c.Row] = map[int]bool{}
        }
        generationMap[c.Row][c.Column] = true

        for i := c.Row-1; i <= c.Row+1; i++ {
            if _, ok := generationMap[i]; !ok {   
                generationMap[i] = map[int]bool{}
            }
            for j := c.Column-1; j <= c.Column+1; j++ {
                generationMap[i][j] = generationMap[i][j]
            }
        }
    }

    return generationMap
}

func getNeighborCount(rowNumber int, colNumber int, gm map[int]map[int]bool) int {
    count := 0

    for i := rowNumber-1; i<=rowNumber+1; i++ {
        for j := colNumber-1; j<=colNumber+1; j++ {
            if gm[i][j] && !(i == rowNumber && j == colNumber) {
                count++
            }
        }
    }

    return count
}

func getNextGeneration(gm map[int]map[int]bool) []cell {
    ng := []cell{}

    for rowNumber, columns := range gm {
        for columnNumber, isAlive := range columns {
            nc := getNeighborCount(rowNumber, columnNumber, gm)
            if (isAlive && (nc == 2 || nc == 3)) ||
               (!isAlive && nc == 3) {
                ng = append(ng, cell{rowNumber, columnNumber})
            }
        }
    }

    return ng
}

func generationHandler(response http.ResponseWriter,
    request *http.Request) {
    response.Header().Set("Access-Control-Allow-Origin", "*")
    response.Header().Add("Access-Control-Allow-Headers", "Content-Type")

    request.ParseForm()

    qa := NewQueryArguments(request.Form)

    ng := qa.Cells
    for i := 0; i < qa.Generations; i++ {
        gm := buildGenerationMap(ng)
        ng = getNextGeneration(gm)
    }

    json.NewEncoder(response).Encode(ng)
}

func main() {
    log.Println("Listening on port 4444...")

    http.HandleFunc("/", generationHandler)
    http.ListenAndServe(":4444", nil)
}