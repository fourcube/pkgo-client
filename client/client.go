package client

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fourcube/pkgo-client/xlog"
)

const DefaultPackageDirectory = "/tmp/pkgo"

var (
	ErrNoUpdates = errors.New("no updates available")
)

type Package struct {
	Name           string
	PublicKey      string
	CurrentVersion string
	WorkingDir     string
	// Will be replaced during update
	ExecutablePath string
	LicenseKey     string
	RepositoryUrl  string
}

func Update(pkg Package) error {
	_, err := os.Executable()
	if err != nil {
		return err
	}

	pkgMeta, err := getMeta(pkg.RepositoryUrl, pkg.Name, "current", pkg.LicenseKey)
	if err != nil || pkgMeta == nil {
		return err
	}
	xlog.Print("updating to %v", pkgMeta.Version)

	if pkgMeta.Version == pkg.CurrentVersion {
		xlog.Print("Version already up to date.")
		return ErrNoUpdates
	}

	fileBytes, err := download(pkg.RepositoryUrl, pkg.Name, "current", pkg.LicenseKey)
	if err != nil {
		return err
	}

	block, _ := pem.Decode([]byte(pkg.PublicKey))

	key, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		xlog.Print("%v", err)
		return err
	}

	ecdsaPublicKey := key.(*ecdsa.PublicKey)
	signatureBytes, err := hex.DecodeString(pkgMeta.Signature)

	if err != nil {
		xlog.Print("%v", err)
		return err
	}

	var ecdsaSig struct {
		R, S *big.Int
	}

	_, err = asn1.Unmarshal(signatureBytes, &ecdsaSig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ECDSA signature: %v", err)
	}

	// Verify the SHA256 hash of the file contents
	h := sha256.New()
	h.Write(fileBytes)
	validBinary := ecdsa.Verify(ecdsaPublicKey, h.Sum(nil), ecdsaSig.R, ecdsaSig.S)

	if !validBinary {
		return fmt.Errorf("signature does not match file contents")
	}

	if pkgMeta.Unpack {
		return extractPackage(pkg, pkgMeta, fileBytes)
	}

	// Save Executable
	newExecutablePath := fmt.Sprintf("%s-tmp", pkg.ExecutablePath)
	oldExecutablePath := pkg.ExecutablePath
	bakExecutablePath := fmt.Sprintf("%s-bak", pkg.ExecutablePath)

	stat, err := os.Stat(oldExecutablePath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(newExecutablePath, fileBytes, stat.Mode())
	if err != nil {
		return err
	}

	err = os.Rename(oldExecutablePath, bakExecutablePath)
	if err != nil {
		return fmt.Errorf("error during backup %v", err)
	}
	err = os.Rename(newExecutablePath, oldExecutablePath)
	if err != nil {
		return fmt.Errorf("error during replacement of binary %v", err)
	}

	return nil
}

func extractPackage(pkg Package, pkgMeta *PackageMeta, fileBytes []byte) error {
	os.MkdirAll(DefaultPackageDirectory, 0777)
	fileName := fmt.Sprintf("%s-%s_", pkg.Name, pkgMeta.Version)
	tmpFile, err := ioutil.TempFile(DefaultPackageDirectory, fileName)
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	xlog.Print("downloaded archive to %s", tmpFile.Name())

	if _, err := tmpFile.Write(fileBytes); err != nil {
		return err
	}

	unpackCmd := exec.Command("tar", "-xvzf", tmpFile.Name(), "-C", pkg.WorkingDir)
	unpackCmd.Dir = pkg.WorkingDir

	out, err := unpackCmd.CombinedOutput()
	if err != nil {
		xlog.Print(string(out))
		return err
	}

	return nil
}

func getMeta(baseUrl string, pkgName string, version string, licenseKey string) (*PackageMeta, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/packages/%s", baseUrl, pkgName)
	req, err := http.NewRequest("GET", url, nil)
	req.Header = map[string][]string{
		"X-License": {licenseKey},
	}

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return nil, fmt.Errorf("package meta request failed with status %v (Invalid license?)", resp.StatusCode)
	}

	pkgMeta := &PackageMeta{}
	err = json.NewDecoder(resp.Body).Decode(pkgMeta)
	if err != nil {
		return nil, err
	}

	return pkgMeta, nil
}

func download(baseUrl string, pkgName string, version string, licenseKey string) ([]byte, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/packages/%s/download", baseUrl, pkgName)
	req, err := http.NewRequest("GET", url, nil)
	req.Header = map[string][]string{
		"X-License": {licenseKey},
	}

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return nil, fmt.Errorf("package download failed with status %v (Invalid license?)", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// executablePath returns a (dir, filename) tuple for the
// current executable
func executablePath() (string, string, os.FileMode) {
	executable, err := os.Executable()
	if err != nil {
		xlog.Fatal("failed to get current executable name, %v", err)
	}

	info, err := os.Stat(executable)
	if err != nil {
		xlog.Fatal("failed to stat current executable name, %v", err)
	}

	return filepath.Dir(executable), filepath.Base(executable), info.Mode()
}
