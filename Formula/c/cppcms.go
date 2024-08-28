
package main

import (
    "fmt"
    "log"
    "os/exec"
)

// cppcmsFormula represents a formula in Go.
type cppcmsFormula struct {
    Description  string
    Homepage     string
    URL          string
    Sha256       string
    Dependencies []string
}

func (pkg cppcmsFormula) Print() {
    fmt.Printf("Name: cppcms\n")
    fmt.Printf("Description: %s\n", pkg.Description)
    fmt.Printf("Homepage: %s\n", pkg.Homepage)
    fmt.Printf("URL: %s\n", pkg.URL)
    fmt.Printf("Sha256: %s\n", pkg.Sha256)
    fmt.Printf("Dependencies: %v\n", pkg.Dependencies)
}

func main() {
    pkg := cppcmsFormula{
        Description:  "Free High Performance Web Development Framework",
        Homepage:     "http://cppcms.com/wikipp/en/page/main",
        URL:          "https://downloads.sourceforge.net/project/cppcms/cppcms/1.2.1/cppcms-1.2.1.tar.bz2",
        Sha256:       "9b87fe3b52ead0794e23882355a33ffae08c1dd340d0b3710cb769e1dba7edb8",
        Dependencies: []string{"cmake", "openssl@3", "pcre", "python@3.12"},
    }

    pkg.Print()

    // Instalar dependencias si no están instaladas
    for _, dep := range pkg.Dependencies {
        if !isDependencyInstalled(dep) {
            fmt.Printf("🛠️ Dependency %s not found. Installing...
", dep)
            cmd := exec.Command("brew", "install", dep)
            if err := cmd.Run(); err != nil {
                log.Fatalf("Error installing dependency %s: %v", dep, err)
            }
        } else {
            fmt.Printf("✅ Dependency %s is already installed.
", dep)
        }
    }

    if err := pkg.Installcppcms(); err != nil {
        log.Fatalf("Error during installation: %v", err)
    }

    fmt.Println("Installation completed successfully.")
}

func (pkg cppcmsFormula) Installcppcms() error {
    cmd := exec.Command("curl", "-O", pkg.URL)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to download: %v", err)
    }

    tarball := "cppcms-1.2.1.tar.bz2"
    cmd = exec.Command("tar", "-xf", tarball)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to extract tarball: %v", err)
    }

    sourceDir := "cppcms-1.2.1.tar"
    cmd = exec.Command("sh", "-c", fmt.Sprintf("cd %s && PKG_CONFIG_PATH=/usr/local/lib/pkgconfig ./configure --sysconfdir=/etc --with-lispdir=/usr/share/emacs/site-lisp --with-packager=Homebrew --with-packager-version=4.15.6 --with-packager-bug-reports=https://github.com/Homebrew/homebrew-core/issues && make install", sourceDir))
    cmd.Stdout = log.Writer()
    cmd.Stderr = log.Writer()

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to configure and install: %v", err)
    }

    return nil
}

func isDependencyInstalled(dep string) bool {
    cmd := exec.Command("brew", "list", dep)
    output, err := cmd.CombinedOutput()
    return err == nil && strings.TrimSpace(string(output)) != ""
}
