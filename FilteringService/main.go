package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

type augmentedRectangle struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Time   string `json:"time"`
}

type rectangles struct {
	Main  rectangle
	Input []rectangle
}

func doIntersect(blX1 int, blY1 int, trX1 int, trY1 int,
	blX2 int, blY2 int, trX2 int, trY2 int) bool {

	// Find bottom-left point of intersection rectangle
	blX := math.Max(float64(blX1), float64(blX2))
	blY := math.Max(float64(blY1), float64(blY2))

	// Find top-right point of intersection rectangle
	trX := math.Min(float64(trX1), float64(trX2))
	trY := math.Min(float64(trY1), float64(trY2))

	// No intersection
	if blX > trX || blY > trY {
		return false
	}

	return true
}

func findIntersectingRectangles(mainRectangle rectangle, rectangles []rectangle, requestTime string) []augmentedRectangle {
	var intersectingRectangles []augmentedRectangle

	for _, input := range rectangles {
		if doIntersect(mainRectangle.X, mainRectangle.Y, mainRectangle.X+mainRectangle.Width, mainRectangle.Y+mainRectangle.Height,
			input.X, input.Y, input.X+input.Width, input.Y+input.Height) {
			log.Println("Squares intersect!")
			intersectingRectangles = append(intersectingRectangles,
				augmentedRectangle{input.X, input.Y, input.Width, input.Height, requestTime})
		} else {
			log.Println("Squares do not intersect!")
		}
	}
	return intersectingRectangles
}

func saveRectangles(rectangles []augmentedRectangle, fileName string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Fatalln("Failed to open file", err)
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, rect := range rectangles {
		row := []string{strconv.Itoa(rect.X), strconv.Itoa(rect.Y), strconv.Itoa(rect.Width), strconv.Itoa(rect.Height), rect.Time}
		if err := writer.Write(row); err != nil {
			log.Fatalln("Error writing record to file", err)
		}
	}
}

func postHandler(w http.ResponseWriter, req *http.Request) {
	requestTime := time.Now().Format("2006-01-02 15:04:05")

	decoder := json.NewDecoder(req.Body)
	var rectangles rectangles
	err := decoder.Decode(&rectangles)
	if err != nil {
		log.Println("Unaccepted request body format!")
		panic(err)
	}
	log.Println(rectangles)

	// Find rectangles that intersect with the main rectangle
	intersectingRectangles := findIntersectingRectangles(rectangles.Main, rectangles.Input, requestTime)

	// Save intersecting rectangles into data.csv file
	saveRectangles(intersectingRectangles, "data.csv")
}

func retrieveRectangles(file *os.File) []augmentedRectangle {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("Error reading records from file", err)
	}

	var rects []augmentedRectangle
	for _, record := range records {
		xInt, _ := strconv.Atoi(record[0])
		yInt, _ := strconv.Atoi(record[1])
		widthInt, _ := strconv.Atoi(record[2])
		heightInt, _ := strconv.Atoi(record[3])
		rect := augmentedRectangle{xInt, yInt, widthInt, heightInt, record[4]}
		rects = append(rects, rect)
	}
	return rects
}

func getHandler(w http.ResponseWriter, req *http.Request) {
	// Retrieve rectangles stored in data.csv
	file, err := os.Open("data.csv")
	if err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode("Data file not found.")
		return
	}
	rects := retrieveRectangles(file)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rects)
}

func multiplexer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getHandler(w, r)
	case "POST":
		postHandler(w, r)
	}
}

func main() {
	http.HandleFunc("/", multiplexer)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
