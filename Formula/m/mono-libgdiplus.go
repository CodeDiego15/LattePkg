
package main

import (
    "fmt"
    "log"
    "os/exec"
)

// mono-libgdiplusFormula represents a formula in Go.
type mono-libgdiplusFormula struct {
    Description  string
    Homepage     string
    URL          string
    Sha256       string
    Dependencies []string
}

func (pkg mono-libgdiplusFormula) Print() {
    fmt.Printf("Name: mono-libgdiplus\n")
    fmt.Printf("Description: %s\n", pkg.Description)
    fmt.Printf("Homepage: %s\n", pkg.Homepage)
    fmt.Printf("URL: %s\n", pkg.URL)
    fmt.Printf("Sha256: %s\n", pkg.Sha256)
    fmt.Printf("Dependencies: %v\n", pkg.Dependencies)
}

func main() {
    pkg := mono-libgdiplusFormula{
        Description:  "GDI+-compatible API on non-Windows operating systems",
        Homepage:     "https://www.mono-project.com/docs/gui/libgdiplus/",
        URL:          "https://raw.githubusercontent.com/Homebrew/formula-patches/03cf8088210822aa2c1ab544ed58ea04c897d9c4/libtool/configure-big_sur.diff",
        Sha256:       "35acd6aebc19843f1a2b3a63e880baceb0f5278ab1ace661e57a502d9d78c93c",
        Dependencies: []string{"pkg-config", "cairo", "fontconfig", "freetype", "gettext", "giflib", "glib", "harfbuzz", "jpeg-turbo", "libexif", "libpng", "libtiff", "pango"},
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

    if err := pkg.Installmono-libgdiplus(); err != nil {
        log.Fatalf("Error during installation: %v", err)
    }

    fmt.Println("Installation completed successfully.")
}

func (pkg mono-libgdiplusFormula) Installmono-libgdiplus() error {
    cmd := exec.Command("curl", "-O", pkg.URL)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to download: %v", err)
    }

    tarball := "configure-big_sur.diff"
    cmd = exec.Command("tar", "-xf", tarball)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to extract tarball: %v", err)
    }

    sourceDir := "configure-big_sur"
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
