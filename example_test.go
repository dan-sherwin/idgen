package idgen_test

import (
	"fmt"
	"time"

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

// Example showing TimestampFromID for direct ID → time.
func ExampleGenerator_TimestampFromID() {
	g, _ := idgen.New()
	raw := g.Generate()
	id := g.Format(raw)
	when, _ := g.TimestampFromID(id)
	_ = when // use timestamp
	// Intentionally no Output block to avoid flakiness.
}

// Example showing custom options (wider width, slower pace).
func ExampleNew_customOptions() {
	g, _ := idgen.New(
		idgen.WithWidth(9),
		idgen.WithPace(2*time.Millisecond),
	)
	raw := g.Generate()
	id := g.Format(raw)
	_ = id
	// No Output: non-deterministic.
}
