package certificate

import (
	"bufio"
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/eviltomorrow/open-terminal/lib/fs"
)

type ApplicationInformation struct {
	CertificateConfig    *CertificateConfig
	NotBefor, NotAfter   time.Time
	CommonName           string
	CountryName          string
	ProvinceName         string
	LocalityName         string
	OrganizationName     string
	OrganizationUnitName string
}

type CertificateConfig struct {
	IsCA           bool
	IP             []net.IP
	DNS            []string
	ExpirationTime time.Duration
}

func generateCertificate(caKey *rsa.PrivateKey, caCert *x509.Certificate, bits int, info *ApplicationInformation) ([]byte, []byte, error) {
	if !info.CertificateConfig.IsCA {
		if caKey == nil || caCert == nil {
			return nil, nil, fmt.Errorf("miss ca key/cert")
		}
	}

	priv, err := rsa.GenerateKey(cryptorand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:         info.CommonName,
			Country:            []string{info.CountryName},
			Province:           []string{info.ProvinceName},
			Locality:           []string{info.LocalityName},
			Organization:       []string{info.OrganizationName},
			OrganizationalUnit: []string{info.OrganizationUnitName},
		},
		NotBefore: time.Now().Add(-24 * time.Hour),
		NotAfter:  time.Now().Add(info.CertificateConfig.ExpirationTime),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
	}

	if info.CertificateConfig.IsCA {
		template.IsCA = true
	} else {
		if i := net.ParseIP(info.CommonName); i != nil {
			template.IPAddresses = append(template.IPAddresses, i)
		} else {
			template.DNSNames = append(template.DNSNames, info.CommonName)
		}
		template.IPAddresses = append(template.IPAddresses, info.CertificateConfig.IP...)
		template.DNSNames = append(template.DNSNames, info.CertificateConfig.DNS...)
	}

	var key *rsa.PrivateKey

	if info.CertificateConfig.IsCA {
		caCert = &template
		key = priv
	} else {
		key = caKey
	}

	certBytes, err := x509.CreateCertificate(cryptorand.Reader, &template, caCert, &priv.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	return x509.MarshalPKCS1PrivateKey(priv), certBytes, nil
}

func readCertificate(path string) (*x509.Certificate, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buffer)
	if block == nil {
		return nil, fmt.Errorf("decode certificate failure, block is nil")
	}

	return x509.ParseCertificate(block.Bytes)
}

func writeCertificate(path string, cert []byte) error {
	_, err := x509.ParseCertificate(cert)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	return err
}

func readPKCS1PrivateKey(path string) (*rsa.PrivateKey, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buffer)
	if block == nil {
		return nil, fmt.Errorf("decode private key failure, block is nil")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

var (
	_ = readPKCS8PrivateKey
	_ = writePKCS1PrivateKey
)

func readPKCS8PrivateKey(path string) (*rsa.PrivateKey, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buffer)
	if block == nil {
		return nil, fmt.Errorf("decode private key failure, block is nil")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	if val, ok := key.(*rsa.PrivateKey); ok {
		return val, nil
	}
	return nil, fmt.Errorf("ParsePKCS8PrivateKey failure")
}

func writePKCS1PrivateKey(path string, privKey []byte) error {
	_, err := x509.ParsePKCS1PrivateKey(privKey)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKey}); err != nil {
		return err
	}

	if err := fs.MkdirAll(filepath.Dir(path)); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	return err
}

func writePKCS8PrivateKey(path string, privKey []byte) error {
	priv, err := x509.ParsePKCS1PrivateKey(privKey)
	if err != nil {
		return err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return err
	}

	if err := fs.MkdirAll(filepath.Dir(path)); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	return err
}

var regS = regexp.MustCompile(`\s+`)

func parseHostDNS(ip []string) []string {
	data := make([]string, 0, 8)
	data = append(data, "localhost")

	switch runtime.GOOS {
	case "windows":
		return data

	case "linux":
		fallthrough

	case "darwin":
		buf, err := os.ReadFile("/etc/hosts")
		if err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(buf))

			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				if text == "" || strings.HasPrefix(text, "#") {
					continue
				}
				text = regS.ReplaceAllString(text, " ")
				attrs := strings.Split(text, " ")
				if len(attrs) == 2 {
					for _, i := range ip {
						if i == attrs[0] {
							data = append(data, attrs[1])
						}
					}
				}
			}
			data = append(data, ip...)
		}
		return data

	default:
	}
	return data
}
