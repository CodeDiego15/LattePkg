package main

import "fmt"

// Aarch64ElfGcc representa una fórmula en Go.
type Aarch64ElfGcc struct {
	Description  string
	Homepage     string
	URL          string
	Sha256       string
	Dependencies []string
}

func (pkg Aarch64ElfGcc) Print() {
	fmt.Printf("Name: Aarch64ElfGcc\\n", "Aarch64ElfGcc")
	fmt.Printf("Description: Aarch64ElfGcc\\n", pkg.Description)
	fmt.Printf("Homepage: Aarch64ElfGcc\\n", pkg.Homepage)
	fmt.Printf("URL: %!s(MISSING)\\n", pkg.URL)
	fmt.Printf("Sha256: %!s(MISSING)\\n", pkg.Sha256)
	fmt.Printf("Dependencies: %!v(MISSING)\\n", pkg.Dependencies)
}

func main() {
	// Crear una instancia de %!s(MISSING)
	pkg := %!s(MISSING){
		Description:  "Descripción de %!s(MISSING)",
		Homepage:     "https://example.com",
		URL:          "https://example.com/example-1.0.0.tar.gz",
		Sha256:       "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Dependencies: []string{"dep1", "dep2"},
	}

	// Imprimir la información de la fórmula
	pkg.Print()
}
