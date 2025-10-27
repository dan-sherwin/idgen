package idgen_test

import (
	"fmt"

	"github.com/dan-sherwin/idgen"
)

// Example showing basic generation and formatting.
func ExampleGenerator_Generate() {
	g, _ := idgen.New()
	raw := g.Generate()
	id := g.Format(raw)
	_ = id                    // use in logs/UI
	fmt.Println(len(id) == 8) // fixed width
	// (No Output block to avoid flakiness; this example is for documentation rendering.)
}

// Example showing Format/Parse round-trip and timestamp conversion.
func ExampleGenerator_Format() {
	g, _ := idgen.New()
	raw := g.Generate()
	id := g.Format(raw)
	back, _ := g.Parse(id)
	when := g.TimestampFromRaw(back)
	_ = when // use timestamp
	fmt.Println(back == raw)
	// (No Output block; sample is non-deterministic by design.)
}
