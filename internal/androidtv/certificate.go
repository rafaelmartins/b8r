package androidtv

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func CreateCertificate(path string) (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("androidtv: failed to generate private key: %w", err)
	}
	rv := &tls.Certificate{
		PrivateKey: priv,
	}

	sn, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, fmt.Errorf("androidtv: failed to generate serial number: %w", err)
	}
	certBase := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Country:            []string{"XX"},
			Organization:       []string{"CompanyName"},
			OrganizationalUnit: []string{"CompanySectionName"},
			Locality:           []string{"CityName"},
			Province:           []string{"StateName"},
			CommonName:         "b8r",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(100, 0, 0),
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certBase, &certBase, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("androidtv: failed to create certificate: %w", err)
	}
	rv.Certificate = append(rv.Certificate, cert)

	fp, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	privv, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	if err := pem.Encode(fp, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privv,
	}); err != nil {
		return nil, err
	}

	if err := pem.Encode(fp, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}); err != nil {
		return nil, err
	}
	return rv, nil
}

func OpenCertificate(path string) (*tls.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rv := &tls.Certificate{}
	for {
		block, remaining := pem.Decode(data)
		if block == nil {
			break
		}
		data = remaining

		switch block.Type {
		case "PRIVATE KEY":
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			rv.PrivateKey = key

		case "CERTIFICATE":
			rv.Certificate = append(rv.Certificate, block.Bytes)
		}
	}

	return rv, nil
}
