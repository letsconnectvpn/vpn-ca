package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type caInfo struct {
	caDir  string
	caKey  crypto.Signer
	caCert *x509.Certificate
}

func caInit(caDir string, certName string, notAfter time.Time) {
	caKey := generateKey(filepath.Join(caDir, "ca.key"))
	makeRootCert(caKey, filepath.Join(caDir, "ca.crt"), certName, notAfter)
}

func getCa(caDir string) *caInfo {
	caKey := readKey(filepath.Join(caDir, "ca.key"))
	caCert := readCert(filepath.Join(caDir, "ca.crt"))

	return &caInfo{caDir, caKey, caCert}
}

func readPem(pemFile, pemType string) []byte {
	pemData, err := ioutil.ReadFile(pemFile)
	fatalIfErr(err, "unable to open PEM")
	block, _ := pem.Decode(pemData)
	if block == nil {
		log.Fatalf("unable to decode PEM")
	}
	if block.Type != pemType {
		log.Fatalf("incorrect PEM type, expected '%s'", pemType)
	}

	return block.Bytes
}

func writePem(pemFile string, derBytes []byte, pemType string) {
	file, err := os.OpenFile(pemFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	fatalIfErr(err, "unable to open file to write key")
	defer file.Close()
	err = pem.Encode(file, &pem.Block{
		Type:  pemType,
		Bytes: derBytes,
	})
	fatalIfErr(err, "unable to convert to PEM")
}

func readKey(pemFile string) crypto.Signer {
	derBytes := readPem(pemFile, "PRIVATE KEY")
	signer, err := x509.ParsePKCS8PrivateKey(derBytes)
	fatalIfErr(err, "unable to parse private key")

	return signer.(crypto.Signer)
}

func readCert(pemFile string) *x509.Certificate {
	derBytes := readPem(pemFile, "CERTIFICATE")
	cert, err := x509.ParseCertificate(derBytes)
	fatalIfErr(err, "unable to parse cert")

	return cert
}

func generateKey(filename string) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 3072)
	fatalIfErr(err, "unable to generate key")
	der, err := x509.MarshalPKCS8PrivateKey(key)
	fatalIfErr(err, "unable to convert key to DER")
	writePem(filename, der, "PRIVATE KEY")

	return key
}

func makeRootCert(key crypto.Signer, filename string, certName string, notAfter time.Time) (*x509.Certificate, error) {
	tpl := getCaTemplate(certName, notAfter)
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, key.Public(), key)
	fatalIfErr(err, "unable to generate DER")
	writePem(filename, der, "CERTIFICATE")

	return x509.ParseCertificate(der)
}

func validateCertName(certName string) {
	validCertName := regexp.MustCompile(`^[a-zA-Z0-9-.]+$`)
	if !validCertName.MatchString(certName) {
		log.Fatalf("invalid 'certName' specified")
	}
}

func getCaTemplate(certName string, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: certName,
		},
		SerialNumber:          generateSerial(),
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}
}

func getClientTemplate(certName string, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: certName,
		},
		SerialNumber:          generateSerial(),
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
}

func getServerTemplate(certName string, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: certName,
		},
		SerialNumber:          generateSerial(),
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{certName},
	}
}

func generateSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	fatalIfErr(err, "unable to generate serial number")

	return serialNumber
}

func getTemplate(certName string, notAfter time.Time, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: certName,
		},
		SerialNumber:          generateSerial(),
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
}

func sign(caInfo *caInfo, certName string, tpl *x509.Certificate) *x509.Certificate {
	certKey := generateKey(filepath.Join(caInfo.caDir, fmt.Sprintf("%s.key", certName)))
	der, err := x509.CreateCertificate(rand.Reader, tpl, caInfo.caCert, certKey.Public(), caInfo.caKey)
	fatalIfErr(err, "unable to generate DER")

	certFile := filepath.Join(caInfo.caDir, fmt.Sprintf("%s.crt", certName))
	writePem(certFile, der, "CERTIFICATE")

	cert, err := x509.ParseCertificate(der)
	fatalIfErr(err, "unable to parse cert")

	return cert
}

func parseNotAfter(notAfter *string, defaultNotAfter time.Time, caNotAfter time.Time) time.Time {
	var notAfterTime time.Time
	if *notAfter == "" {
		notAfterTime = defaultNotAfter
	} else if *notAfter == "CA" {
		notAfterTime = caNotAfter
	} else {
		parsedTime, err := time.Parse(time.RFC3339, *notAfter)
		fatalIfErr(err, "unable to parse -not-after")
		if !parsedTime.After(time.Now()) {
			log.Fatalf("-not-after must be in the future")
		}
		notAfterTime = parsedTime
	}

	// make sure the certificate won't outlive the CA
	if notAfterTime.After(caNotAfter) {
		log.Fatalf("-not-after can't outlive the CA")
	}

	return notAfterTime
}

func main() {
	var caDir = os.Getenv("CA_DIR")
	if "" == caDir {
		caDir = "."
	}
	var initCa = flag.Bool("init-ca", false, "Generate CA Certificate/Key")
	var serverCert = flag.Bool("server", false, "Generate a Server Certificate/Key")
	var clientCert = flag.Bool("client", false, "Generate a Client Certificate/Key")
	var certName = flag.String("name", "", "Name on the CA/Server/Client Certificate")
	var notAfter = flag.String("not-after", "", "Limit Certificate Validity for Server/Client Certificate")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	if *initCa {
		if *certName == "" {
			*certName = "Root CA"
		}
		// CA gets written to ca.key/ca.crt, so no need to validate the
		// name to prevent issues writing the certificate...
		caInit(caDir, *certName, time.Now().AddDate(5, 0, 0))
		return
	}

	if *certName == "" {
		flag.Usage()
		os.Exit(1)
	}

	caInfo := getCa(caDir)

	if *serverCert {
		validateCertName(*certName)
		notAfterTime := parseNotAfter(notAfter, time.Now().AddDate(1, 0, 0), caInfo.caCert.NotAfter)
		sign(caInfo, *certName, getServerTemplate(*certName, notAfterTime))
		return
	}

	if *clientCert {
		validateCertName(*certName)
		notAfterTime := parseNotAfter(notAfter, time.Now().AddDate(1, 0, 0), caInfo.caCert.NotAfter)
		sign(caInfo, *certName, getClientTemplate(*certName, notAfterTime))
		return
	}

	flag.Usage()
	os.Exit(1)
}

func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err)
	}
}
