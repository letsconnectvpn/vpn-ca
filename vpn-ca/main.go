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

func initCa(caDir string) {
	caKey := generateKey(filepath.Join(caDir, "ca.key"))
	makeRootCert(caKey, filepath.Join(caDir, "ca.crt"))
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

func makeRootCert(key crypto.Signer, filename string) (*x509.Certificate, error) {
	tpl := getCaTemplate()
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, key.Public(), key)
	fatalIfErr(err, "unable to generate DER")
	writePem(filename, der, "CERTIFICATE")

	return x509.ParseCertificate(der)
}

func validateCommonName(commonName string) {
	validCommonName := regexp.MustCompile(`^[a-zA-Z0-9-.]+$`)
	if !validCommonName.MatchString(commonName) {
		log.Fatalf("invalid 'commonName' specified")
	}
}

func getCaTemplate() *x509.Certificate {
	// 5 years
	tpl := getTemplate("VPN CA", time.Now().AddDate(5, 0, 0), x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth})
	tpl.IsCA = true
	tpl.MaxPathLenZero = true

	return tpl
}

func getClientTemplate(commonName string, notAfter *time.Time) *x509.Certificate {
	return getTemplate(commonName, *notAfter, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
}

func getServerTemplate(commonName string, notAfter *time.Time) *x509.Certificate {
	return getTemplate(commonName, *notAfter, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
}

func generateSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	fatalIfErr(err, "unable to generate serial number")

	return serialNumber
}

func getTemplate(commonName string, notAfter time.Time, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: commonName,
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

func sign(caInfo *caInfo, commonName string, tpl *x509.Certificate) *x509.Certificate {
	certKey := generateKey(filepath.Join(caInfo.caDir, fmt.Sprintf("%s.key", commonName)))
	der, err := x509.CreateCertificate(rand.Reader, tpl, caInfo.caCert, certKey.Public(), caInfo.caKey)
	fatalIfErr(err, "unable to generate DER")

	certFile := filepath.Join(caInfo.caDir, fmt.Sprintf("%s.crt", commonName))
	writePem(certFile, der, "CERTIFICATE")

	cert, err := x509.ParseCertificate(der)
	fatalIfErr(err, "unable to parse cert")

	return cert
}

func parseNotAfter(notAfter *string, caNotAfter time.Time) time.Time {
	var notAfterTime time.Time
	if *notAfter == "" {
		// by default a generated certificate will expire after 1 year
		notAfterTime = time.Now().AddDate(1, 0, 0)
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
	var caDir = flag.String("ca-dir", ".", "the CA dir")
	var caInit = flag.Bool("init", false, "initialize the CA")
	var serverCommonName = flag.String("server", "", "generate a server certificate with provided CN")
	var clientCommonName = flag.String("client", "", "generate a client certificate with provided CN")
	var notAfter = flag.String("not-after", "", "limit certificate validity. Format: \"2019-08-16T14:00:00+00:00\", or \"CA\" to expire at CA expiry")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	if *caInit {
		initCa(*caDir)
		return
	}

	if *serverCommonName == "" && *clientCommonName == "" {
		flag.Usage()
		os.Exit(1)
	}

	caInfo := getCa(*caDir)

	if *serverCommonName != "" {
		validateCommonName(*serverCommonName)
		notAfterTime := parseNotAfter(notAfter, caInfo.caCert.NotAfter)
		sign(caInfo, *serverCommonName, getServerTemplate(*serverCommonName, &notAfterTime))
		return
	}

	if *clientCommonName != "" {
		validateCommonName(*clientCommonName)
		notAfterTime := parseNotAfter(notAfter, caInfo.caCert.NotAfter)
		sign(caInfo, *clientCommonName, getClientTemplate(*clientCommonName, &notAfterTime))
		return
	}
}

func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err)
	}
}
