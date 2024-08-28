
package main

import (
    "fmt"
    "log"
    "os/exec"
)

// terminalimageviewerFormula represents a formula in Go.
type terminalimageviewerFormula struct {
    Description  string
    Homepage     string
    URL          string
    Sha256       string
    Dependencies []string
}

func (pkg terminalimageviewerFormula) Print() {
    fmt.Printf("Name: terminalimageviewer\n")
    fmt.Printf("Description: %s\n", pkg.Description)
    fmt.Printf("Homepage: %s\n", pkg.Homepage)
    fmt.Printf("URL: %s\n", pkg.URL)
    fmt.Printf("Sha256: %s\n", pkg.Sha256)
    fmt.Printf("Dependencies: %v\n", pkg.Dependencies)
}

func main() {
    pkg := terminalimageviewerFormula{
        Description:  "Display images in a terminal using block graphic characters",
        Homepage:     "https://github.com/stefanhaustein/TerminalImageViewer",
        URL:          "https://github.com/stefanhaustein/TerminalImageViewer/archive/refs/tags/v1.2.1.tar.gz",
        Sha256:       "09bed7a395fb96ff84f0bf05ed864786956f90048e25c4c918cdf68055b88fbe",
        Dependencies: []string{"imagemagick"},
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

    if err := pkg.Installterminalimageviewer(); err != nil {
        log.Fatalf("Error during installation: %v", err)
    }

    fmt.Println("Installation completed successfully.")
}

func (pkg terminalimageviewerFormula) Installterminalimageviewer() error {
    cmd := exec.Command("curl", "-O", pkg.URL)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to download: %v", err)
    }

    tarball := "v1.2.1.tar.gz"
    cmd = exec.Command("tar", "-xf", tarball)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to extract tarball: %v", err)
    }

    sourceDir := "v1.2.1.tar"
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
