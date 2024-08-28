
package main

import (
    "fmt"
    "log"
    "os/exec"
)

// open-scene-graphFormula represents a formula in Go.
type open-scene-graphFormula struct {
    Description  string
    Homepage     string
    URL          string
    Sha256       string
    Dependencies []string
}

func (pkg open-scene-graphFormula) Print() {
    fmt.Printf("Name: open-scene-graph\n")
    fmt.Printf("Description: %s\n", pkg.Description)
    fmt.Printf("Homepage: %s\n", pkg.Homepage)
    fmt.Printf("URL: %s\n", pkg.URL)
    fmt.Printf("Sha256: %s\n", pkg.Sha256)
    fmt.Printf("Dependencies: %v\n", pkg.Dependencies)
}

func main() {
    pkg := open-scene-graphFormula{
        Description:  "3D graphics toolkit",
        Homepage:     "https://github.com/openscenegraph/OpenSceneGraph",
        URL:          "https://github.com/openscenegraph/OpenSceneGraph/commit/21f5a0adfb57dc4c28b696e93beface45de28194.patch?full_index=1",
        Sha256:       "43c4367454e8de65443937a3509f96d4d273b50431b0a4fde16607c88183b247",
        Dependencies: []string{"cmake", "doxygen", "graphviz", "pkg-config", "fontconfig", "freetype", "jpeg-turbo", "sdl2", "cairo", "giflib", "glib", "libpng", "librsvg", "libx11", "libxinerama", "libxrandr", "mesa"},
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

    if err := pkg.Installopen-scene-graph(); err != nil {
        log.Fatalf("Error during installation: %v", err)
    }

    fmt.Println("Installation completed successfully.")
}

func (pkg open-scene-graphFormula) Installopen-scene-graph() error {
    cmd := exec.Command("curl", "-O", pkg.URL)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to download: %v", err)
    }

    tarball := "21f5a0adfb57dc4c28b696e93beface45de28194.patch?full_index=1"
    cmd = exec.Command("tar", "-xf", tarball)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to extract tarball: %v", err)
    }

    sourceDir := "21f5a0adfb57dc4c28b696e93beface45de28194"
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
