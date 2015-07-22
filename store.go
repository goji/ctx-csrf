package csrf

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/zenazn/goji/web"
)

// Store represents the session storage used for CSRF tokens.
type Store interface {
	// Get returns the real CSRF token from the store.
	Get(c *web.C, r *http.Request) ([]byte, error)
	// Save stores the real CSRF token in the store and writes a
	// cookie to the http.ResponseWriter.
	// For non-cookie stores, the cookie should contain a unique (256 bit) ID
	// or key that references the token in the backend store.
	// csrf.GenerateRandomBytes is a helper function for generating secure IDs.
	Save(token []byte, w http.ResponseWriter) error
}

// CookieStore is a signed cookie session store for CSRF tokens.
type CookieStore struct {
	name   string
	maxAge int
	sc     *securecookie.SecureCookie
}

// Get retrieves a CSRF token from the session cookie. It returns an empty token
// if decoding fails (e.g. HMAC validation fails or the named cookie doesn't exist).
func (cs *CookieStore) Get(c *web.C, r *http.Request) ([]byte, error) {
	// Retrieve the cookie from the request
	cookie, err := r.Cookie(cs.name)
	if err != nil {
		return nil, err
	}

	token := make([]byte, tokenLength)
	// Decode the HMAC authenticated cookie.
	err = cs.sc.Decode(cs.name, cookie.Value, &token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Save stores the CSRF token in the session cookie.
func (cs *CookieStore) Save(token []byte, w http.ResponseWriter) error {
	// Generate an encoded cookie value with the CSRF token.
	encoded, err := cs.sc.Encode(cs.name, token)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:   cs.name,
		Value:  encoded,
		MaxAge: cs.maxAge,
	}

	// Set the Expires field on the cookie based on the MaxAge
	if cs.maxAge > 0 {
		cookie.Expires = time.Now().Add(
			time.Duration(cs.maxAge) * time.Second)
	} else {
		cookie.Expires = time.Unix(1, 0)
	}

	// Write the authenticated cookie to the response.
	http.SetCookie(w, cookie)

	return nil
}