package webpushy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// the following lifted mostly wholesale from https://github.com/SherClockHolmes/webpush-go/

func GenerateSenderKeys(randrd io.Reader) (SenderKeys, error) {
	// Get the private key from the P256 curve
	curve := elliptic.P256()

	private, x, y, err := elliptic.GenerateKey(curve, randrd)
	if err != nil {
		return SenderKeys{}, err
	}

	public := elliptic.Marshal(curve, x, y)

	pub := base64.RawURLEncoding.EncodeToString(public)
	priv := base64.RawURLEncoding.EncodeToString(private)

	return SenderKeys{Public: pub, Private: priv}, nil
}

func makeRequest(opts *SenderOptions, endpoint string, ttl time.Duration, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", strconv.Itoa(int(ttl.Seconds())))

	// Get VAPID Authorization header
	vapidAuthHeader, err := getVAPIDAuthorizationHeader(
		endpoint,
		opts.Identifier,
		opts.Keys.Public,
		opts.Keys.Private,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", vapidAuthHeader)
	return req, nil
}

func getVAPIDAuthorizationHeader(endpoint, subscriber, vapidPublicKey, vapidPrivateKey string) (string, error) {
	// Create the JWT token
	subURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"aud": fmt.Sprintf("%s://%s", subURL.Scheme, subURL.Host),
		"exp": time.Now().Add(time.Hour * 12).Unix(),
		"sub": fmt.Sprintf("mailto:%s", subscriber),
	})

	// ECDSA
	decodedVapidPrivateKey, err := base64.RawURLEncoding.DecodeString(vapidPrivateKey)
	if err != nil {
		return "", err
	}

	privKey := generateVAPIDHeaderKeys(decodedVapidPrivateKey)

	// Sign token with private key
	jwtString, err := token.SignedString(privKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"vapid t=%s, k=%s",
		jwtString,
		vapidPublicKey,
	), nil
}

// Generates the ECDSA public and private keys for the JWT encryption
func generateVAPIDHeaderKeys(privateKey []byte) *ecdsa.PrivateKey {
	// Public key
	curve := elliptic.P256()
	px, py := curve.ScalarMult(
		curve.Params().Gx,
		curve.Params().Gy,
		privateKey,
	)

	pubKey := ecdsa.PublicKey{
		Curve: curve,
		X:     px,
		Y:     py,
	}

	// Private key
	d := &big.Int{}
	d.SetBytes(privateKey)

	return &ecdsa.PrivateKey{
		PublicKey: pubKey,
		D:         d,
	}
}
