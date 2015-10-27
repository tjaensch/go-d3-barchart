package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type stocksArray []singleStock

type singleStock struct {
	Name   string `json:"t"`
	Amount string `json:"l"`
}

var (
	err      error
	stocks   stocksArray
	response *http.Response
	body     []byte
)

func stockvalues(w http.ResponseWriter, r *http.Request) {
	// Use http://finance.google.com/finance/info?client=ig&q=NASDAQ:GOOG to get a JSON response
	response, err = http.Get("http://finance.google.com/finance/info?client=ig&q=NASDAQ:GOOG,NASDAQ:AAPL,NASDAQ:MSFT,NASDAQ:EBAY,NASDAQ:NFLX,NASDAQ:CSCO,NASDAQ:INTC")
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	// Read the data into a byte slice
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	// Remove whitespace from response
	data := bytes.TrimSpace(body)

	// Remove leading slashes and blank space to get byte slice that can be unmarshaled from JSON
	data = bytes.TrimPrefix(data, []byte("// "))

	//Unmarshal the JSON byte slice to a predefined struct
	err = json.Unmarshal(data, &stocks)
	if err != nil {
		fmt.Println(err)
	}

	//Marshal selected data back to JSON
	jsonData, err := json.Marshal(stocks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Write JSON to command line
	fmt.Println(string(jsonData))

	//Write JSON to HTTP response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

// handler for D3 bar chart
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, stats)
}

const stats = `
<!DOCTYPE html>
<meta charset="utf-8">
<style>
    .axis path,
    .axis line {
        fill: none;
        stroke: black;
        shape-rendering: crispEdges;
    }
    .axis text {
        font-family: sans-serif;
        font-size: 11px;
    }
</style>

<script src="https://cdnjs.cloudflare.com/ajax/libs/d3/3.5.6/d3.min.js" charset="utf-8"></script>
<body>
  <p>Simple D3 demo that fetches stock data from the Google Finance API with Go on the backend and displays a bar chart with D3 on the frontend. Code on <a href="https://github.com/tjaensch/go-d3-barchart" target="_blank">GitHub</a>. Thomas Jaensch 2015.</p> 
</body>
<script>

//Width and height
      var w = 800;
      var h = 300;

//Data
  d3.json("https://go-d3-barchart.appspot.com/stockvalues", function(error, data) {
              if (error) return console.warn(error);
              console.log(data);
      
      //Scales
      var xScale = d3.scale.ordinal()
              .domain(d3.range(data.length))
              .rangeBands([0, w], 0.05);

      var yScale = d3.scale.linear()
              .domain([0, d3.max(data, function(d) { return d.l; })])
              .range([10, h]);
      
      //Define key function, to be used when binding data
      var key = function(d) {
        return d.t;
      };
      
      //Create SVG element
      var svg = d3.select("body")
            .append("svg")
            .attr("width", w)
            .attr("height", h);

      //Create bars
      svg.selectAll("rect")
         .data(data, key)
         .enter()
         .append("rect")
         .attr("x", function(d, i) {
            return xScale(i);
         })
         .attr("y", function(d) {
            return h - yScale(d.l);
         })
         .attr("width", xScale.rangeBand())
         .attr("height", function(d) {
            return yScale(d.l);
         })
         .attr("fill", function(d) {
          return "rgb(0, 0, " + (d.l / 10) + ")";
         });

      //Create labels
      svg.selectAll("text")
         .data(data, key)
         .enter()
         .append("text")
         .text(function(d) {
            return d.t + " $" + d.l;
         })
         .attr("text-anchor", "middle")
         .attr("x", function(d, i) {
            return xScale(i) + xScale.rangeBand() / 2;
         })
         .attr("y", function(d) {
            return h - yScale(d.l) + 14;
         })
         .attr("font-family", "sans-serif")
         .attr("font-size", "11px")
         .attr("fill", "white");
});
</script>
`

func main() {

	http.HandleFunc("/stockvalues", stockvalues)
	http.HandleFunc("/", handler)
	log.Println("Listening on 8080...")
	http.ListenAndServe(":8080", nil)
}
