package main

import (
	"github.com/Dakota628/d4parse/pkg/bnet/cdn"
	"github.com/Dakota628/d4parse/pkg/bnet/ribbit2"
	"io"
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

	resp, err := _cdn.GetBuildConfig()
	if err != nil {
		log.Fatalf("error getting build config: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading build config body: %s", err)
	}

	println(string(body))
}
