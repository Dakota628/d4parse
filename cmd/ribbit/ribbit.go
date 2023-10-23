package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bnet/cdn"
	"github.com/Dakota628/d4parse/pkg/bnet/ribbit2"
	"log"
	"os"
)

func main() {
	c, err := ribbit2.NewClient(os.Args[1])
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}

	//// === START RIBBIT ===
	//req := ribbit2.Request{
	//	//Command: []byte("v2/products/fenris/cdns"),
	//	Command: []byte("v2/products/fenris/versions"),
	//}
	//
	//rresp, err := c.Do(req)
	//if err != nil {
	//	log.Fatalf("error sending request: %s", err)
	//}
	//fmt.Printf("====\ndata\n====\n%s\n====\n", string(rresp.Data))
	//
	//doc, err := rresp.BPSV()
	//if err != nil {
	//	log.Fatalf("error parsing BPSV: %s", err)
	//}
	//fmt.Printf("====\nparsed\n====\n%#v\n====\n", doc)
	//// === END RIBBIT ===

	_cdn, err := cdn.NewCDN(c, "fenris", "us")
	if err != nil {
		log.Fatalf("error creating CDN: %s", err)
	}

	data, err := _cdn.GetEncodingTable()
	if err != nil {
		log.Fatalf("error getting encoding table: %s", err)
	}
	fmt.Printf("%#v", data)
}
