package main

import "fmt"

// Cmigemo representa una fórmula en Go.
type Cmigemo struct {
	Description  string
	Homepage     string
	URL          string
	Sha256       string
	Dependencies []string
}

func (pkg Cmigemo) Print() {
	fmt.Printf("Name: Cmigemo\\n", "Cmigemo")
	fmt.Printf("Description: Cmigemo\\n", pkg.Description)
	fmt.Printf("Homepage: Cmigemo\\n", pkg.Homepage)
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
