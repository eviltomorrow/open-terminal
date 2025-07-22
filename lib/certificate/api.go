package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc/credentials"
)

var (
	CertsDir   = "../var/certs"
	CaCertsDir = "../usr/certs"
)

type Config struct {
	CaCertFile string
	CaKeyFile  string

	ClientCertFile string
	ClientKeyFile  string
	ServerCertFile string
	ServerKeyFile  string
}

func BuildDefaultAppInfo(ip []string) *ApplicationInformation {
	var app = &ApplicationInformation{
		CertificateConfig: &CertificateConfig{
			IsCA:           false,
			ExpirationTime: 24 * time.Hour * 365 * 10,
		},
		CommonName:           "www.ultrapower.com.cn",
		CountryName:          "CN",
		ProvinceName:         "BeiJing",
		LocalityName:         "BeiJing",
		OrganizationName:     "Ultr@Power",
		OrganizationUnitName: "CBG",
		NotBefor:             time.Now().AddDate(0, 0, -7),
		NotAfter:             time.Now().AddDate(10, 0, -7),
	}

	for _, v := range ip {
		if v != "" {
			parsedIP := net.ParseIP(v)
			if parsedIP != nil {
				app.CertificateConfig.IP = append(app.CertificateConfig.IP, parsedIP)
			}
		}
	}

	dns := parseHostDNS(ip)
	app.CertificateConfig.DNS = append(app.CertificateConfig.DNS, dns...)

	return app
}

func CreateOrOverrideFile(info *ApplicationInformation, c *Config) error {
	caKey, err := readPKCS1PrivateKey(c.CaKeyFile)
	if err != nil {
		return err
	}

	caCert, err := readCertificate(c.CaCertFile)
	if err != nil {
		return err
	}

	clientPrivBytes, clientCertBytes, err := generateCertificate(caKey, caCert, 2048, info)
	if err != nil {
		return err
	}

	if c.ClientKeyFile != "" {
		if err := writePKCS8PrivateKey(c.ClientKeyFile, clientPrivBytes); err != nil {
			return err
		}
	}
	if c.ClientCertFile != "" {
		if err := writeCertificate(c.ClientCertFile, clientCertBytes); err != nil {
			return err
		}
	}

	serverPrivBytes, serverCertBytes, err := generateCertificate(caKey, caCert, 2048, info)
	if err != nil {
		return err
	}

	if c.ServerKeyFile != "" {
		if err := writePKCS8PrivateKey(c.ServerKeyFile, serverPrivBytes); err != nil {
			return err
		}
	}

	if c.ServerCertFile != "" {
		if err := writeCertificate(c.ServerCertFile, serverCertBytes); err != nil {
			return err
		}
	}

	return nil
}

func LoadClientCredentials(doamin string, c *Config) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(c.ClientCertFile, c.ClientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("LoadX509KeyPair failure, nest error: %v", err)
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(c.CaCertFile)
	if err != nil {
		return nil, fmt.Errorf("ReadFile ca cert file failure, nest error: %v", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("AppendCertsFromPEM failure")
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   doamin,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})

	return creds, nil
}

func LoadServerCredentials(c *Config) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(c.ServerCertFile, c.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("LoadX509KeyPair failure, nest error: %v", err)
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(c.CaCertFile)
	if err != nil {
		return nil, fmt.Errorf("ReadFile ca cert file failure, nest error: %v", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("AppendCertsFromPEM failure")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	})
	return creds, nil
}
