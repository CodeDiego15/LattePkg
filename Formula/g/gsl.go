package main

import "fmt"

// GslFormulaFormula representa una fórmula en Go.
type GslFormulaFormula struct {
	Description  string
	Homepage     string
	URL          string
	Sha256       string
	Dependencies []string
}

func (pkg GslFormulaFormula) Print() {
	fmt.Printf("Name: Gsl\\n")
	fmt.Printf("Description: %s\\n", pkg.Description)
	fmt.Printf("Homepage: %s\\n", pkg.Homepage)
	fmt.Printf("URL: %s\\n", pkg.URL)
	fmt.Printf("Sha256: %s\\n", pkg.Sha256)
	fmt.Printf("Dependencies: %v\\n", pkg.Dependencies)
}

func main() {
	// Crear una instancia de GslFormulaFormula
	pkg := GslFormulaFormula{
		Description:  "Descripción de Gsl",
		Homepage:     "https://example.com",
		URL:          "https://example.com/example-1.0.0.tar.gz",
		Sha256:       "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Dependencies: []string{"dep1", "dep2"},
	}

	// Imprimir la información de la fórmula
	pkg.Print()
}
