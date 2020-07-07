package main

import (
	"context"
	"encoding/json"
	"go.opencensus.io/trace"
	"log"
)

func main() {
	_, span := trace.StartSpan(context.Background(), "test: upload")
	defer span.End()
	spanCJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		panic(err)
	}

	var spanCtx trace.SpanContext
	err = json.Unmarshal(spanCJSON, &spanCtx)
	if err != nil {
		panic(err)
	}

	log.Printf("%+v", spanCtx)
}
