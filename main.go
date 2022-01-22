// https://github.com/PacktPublishing/Go-Machine-Learning-Projects/tree/master/Chapter02
package main

import (
	"fmt"
	"gonum.org/v1/plot/vg"
	"gorgonia.org/tensor"
	"log"
	"os"
)

func main() {
	f, err := os.Open("data/train.csv")
	if err != nil {
		log.Fatalln("Open unknown error", err)
	}

	headers, data, indices, err := ingest(f)
	if err != nil {
		log.Fatalln("ingest unknown error", err)
	}

	c := cardinality(indices)
	for i, h := range headers {
		fmt.Printf("%s: %v\n", h, c[i])
	}

	rowLen, colLen, xsBack, ysBack, _, _ := clean(headers, data, indices, datahints, ignored)
	xs := tensor.New(tensor.WithShape(rowLen, colLen), tensor.WithBacking(xsBack))
	ys := tensor.New(tensor.WithShape(rowLen, 1), tensor.WithBacking(ysBack))

	fmt.Printf("rowLen: %d, colLen: %d\n", rowLen, colLen)
	fmt.Printf("xs:\n%+1.1sys:\n%1.1s", xs, ys)

	idxOfInterest := 19
	cef := cef(ysBack, indices[idxOfInterest])
	plt, err := plotCEF(cef)
	if err != nil {
		log.Fatalln("plotCEF unknown error", err)
	}
	plt.Title.Text = fmt.Sprintf("CEF for %s", headers[idxOfInterest])
	plt.X.Label.Text = headers[idxOfInterest]
	plt.Y.Label.Text = "EV of House Price"
	err = plt.Save(25*vg.Centimeter, 25*vg.Centimeter, "CEF.png")
	if err != nil {
		log.Fatalln("CEF Save unknown error", err)
	}

	hist, err := plotHist(ysBack)
	if err != nil {
		log.Fatalln("plotHist unknown error", err)
	}
	hist.Title.Text = "Histogram of House Price"
	err = hist.Save(25*vg.Centimeter, 25*vg.Centimeter, "hist.png")
	if err != nil {
		log.Fatalln("hist Save unknown error", err)
	}
}
