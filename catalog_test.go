package catalog

import (
    "os"
    "testing"
)

func TestLoadFixture(t *testing.T) {
    file, err := os.Open("fixtures/catalog.json")
    if err != nil { t.Fatal(err) }
    defer file.Close()
    entries, err := Load(file)
    if err != nil { t.Fatal(err) }
    if len(entries) < 2 { t.Fatalf("目录条目不足: %d", len(entries)) }
}
