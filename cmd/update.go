package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

const (
	repoOwner = "alvaroeng98"
	repoName  = "HarnessInit"
	binaryName = "harness-init"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza harness-init a la última versión",
	RunE:  runUpdate,
}

func runUpdate(_ *cobra.Command, _ []string) error {
	if version == "dev" {
		fmt.Fprintln(os.Stderr, "Build local (dev): actualización automática no disponible.")
		fmt.Fprintln(os.Stderr, "Compila desde el código fuente: go install github.com/alvaroeng98/HarnessInit@latest")
		return nil
	}

	fmt.Print("Comprobando última versión... ")
	latestTag, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("no se pudo consultar la última versión: %w", err)
	}
	fmt.Println(latestTag)

	if latestTag == "v"+version {
		fmt.Printf("Ya en la versión más reciente (%s).\n", latestTag)
		return nil
	}

	archMap := map[string]string{"amd64": "amd64", "arm64": "arm64"}
	goarch, ok := archMap[runtime.GOARCH]
	if !ok {
		return fmt.Errorf("arquitectura no soportada: %s", runtime.GOARCH)
	}

	goos := runtime.GOOS
	archiveName := fmt.Sprintf("%s-%s-%s.tar.gz", binaryName, goos, goarch)
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", repoOwner, repoName, latestTag, archiveName)

	fmt.Printf("Descargando %s...\n", latestTag)
	tmpBin, err := downloadAndExtract(url)
	if err != nil {
		return fmt.Errorf("descarga fallida: %w", err)
	}
	defer os.Remove(tmpBin)

	currentBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("no se pudo determinar ruta del binario actual: %w", err)
	}

	if runtime.GOOS == "windows" {
		dest := filepath.Join(filepath.Dir(currentBin), binaryName+".new")
		if err := os.Rename(tmpBin, dest); err != nil {
			return fmt.Errorf("no se pudo guardar nuevo binario: %w", err)
		}
		fmt.Printf("Descargado en: %s\n", dest)
		fmt.Printf("Reemplaza manualmente: move /Y %s %s\n", dest, currentBin)
		return nil
	}

	if err := os.Rename(tmpBin, currentBin); err != nil {
		return fmt.Errorf("no se pudo reemplazar el binario: %w", err)
	}

	fmt.Printf("Actualizado a %s.\n", latestTag)
	return nil
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("respuesta inesperada de GitHub API")
	}
	return payload.TagName, nil
}

func downloadAndExtract(url string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d al descargar %s", resp.StatusCode, url)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Name != binaryName && hdr.Name != binaryName+".exe" {
			continue
		}

		tmpFile, err := os.CreateTemp("", binaryName+"-*")
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(tmpFile, tr); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", err
		}
		tmpFile.Close()
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			os.Remove(tmpFile.Name())
			return "", err
		}
		return tmpFile.Name(), nil
	}

	return "", fmt.Errorf("binario '%s' no encontrado en el archivo", binaryName)
}
