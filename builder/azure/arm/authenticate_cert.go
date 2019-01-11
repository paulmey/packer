package arm

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
)

func NewCertOAuthTokenProvider(env azure.Environment, clientID, clientCertPath, tenantID string) (oAuthTokenProvider, error) {
	cert, key, err := readCert(clientCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading certificate: %v", err)
	}

	audience := fmt.Sprintf("%s%s/oauth2/token", env.ActiveDirectoryEndpoint, tenantID)
	jwt, err := makeJWT(clientID, audience, cert, key, time.Hour, true)
	if err != nil {
		return nil, fmt.Errorf("Error generating JWT: %v", err)
	}

	return NewJWTOAuthTokenProvider(env, clientID, jwt, tenantID), nil
}

func makeJWT(clientID string, audience string,
	cert *x509.Certificate, privatekey interface{},
	validFor time.Duration, includeX5c bool) (string, error) {

	// The jti (JWT ID) claim provides a unique identifier for the JWT.
	jti := make([]byte, 20)
	_, err := rand.Read(jti)
	if err != nil {
		return "", err
	}

	var token *jwt.Token
	if cert.PublicKeyAlgorithm == x509.RSA {
		token = jwt.New(jwt.SigningMethodRS256)
	} else if cert.PublicKeyAlgorithm == x509.ECDSA {
		token = jwt.New(jwt.SigningMethodES256)
	} else {
		return "", fmt.Errorf("Don't know how to handle this type of key algorithm: %v", cert.PublicKeyAlgorithm)
	}

	hasher := sha1.New()
	if _, err := hasher.Write(cert.Raw); err != nil {
		return "", err
	}
	token.Header["x5t"] = base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	if includeX5c {
		token.Header["x5c"] = []string{base64.StdEncoding.EncodeToString(cert.Raw)}
	}

	token.Claims = jwt.MapClaims{
		"aud": audience,
		"iss": clientID,
		"sub": clientID,
		"jti": base64.URLEncoding.EncodeToString(jti),
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(validFor).Unix(),
	}

	return token.SignedString(privatekey)
}

func readCert(file string) (cert *x509.Certificate, key interface{}, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	d, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}

	blocks := []*pem.Block{}
	for len(d) > 0 {
		var b *pem.Block
		b, d = pem.Decode(d)
		if b == nil {
			break
		}
		blocks = append(blocks, b)
	}

	certs := []*x509.Certificate{}
	for _, block := range blocks {
		if block.Type == "CERTIFICATE" {
			c, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf(
					"Failed to read certificate block: %v", err)
			}
			certs = append(certs, c)
		} else if block.Type == "PRIVATE KEY" {
			key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf(
					"Failed to read private key block: %v", err)
			}
		}
	}

	if key == nil {
		return nil, nil, fmt.Errorf("Did not find private key in pem file")
	}

	switch key := key.(type) {
	case *rsa.PrivateKey:
		for _, c := range certs {
			if cp, ok := c.PublicKey.(*rsa.PublicKey); ok &&
				(cp.N.Cmp(key.PublicKey.N) == 0) {
				cert = c
			}
		}

	case *ecdsa.PrivateKey:
		for _, c := range certs {
			if cp, ok := c.PublicKey.(*ecdsa.PublicKey); ok &&
				(cp.X.Cmp(key.PublicKey.X) == 0) &&
				(cp.Y.Cmp(key.PublicKey.Y) == 0) {
				cert = c
			}
		}
	}

	if cert == nil {
		return nil, nil, fmt.Errorf("Did not find certificate belonging to private key in pem file")
	}

	return cert, key, nil
}
