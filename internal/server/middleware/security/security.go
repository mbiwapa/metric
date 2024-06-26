// Package security provides middleware for checking the IP addresses of requests.
package security

import (
	"net"
	"net/http"
)

// New creates a new HTTP middleware that checks if the IP address from the X-Real-IP header is within a trusted subnet.
// It returns a function that takes an http.Handler and returns an http.Handler.
//
// If the IP address from the X-Real-IP header is not within the trusted subnet, it responds with a 403 Forbidden status.
// If the trustedSubnet parameter is empty, the request is processed without additional restrictions.
//
// Parameters:
//   - trustedSubnet: A string representing the trusted subnet in CIDR format.
//
// Returns:
//   - A function that takes an http.Handler and returns an http.Handler.
func New(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		fn := func(w http.ResponseWriter, r *http.Request) {
			// Check if the request's IP address is within the trusted subnet
			if trustedSubnet != "" {
				realIP := r.Header.Get("X-Real-IP")
				if realIP == "" {
					http.Error(w, "missing X-Real-IP header", http.StatusForbidden)
					return
				}

				ip := net.ParseIP(realIP)
				_, subnet, err := net.ParseCIDR(trustedSubnet)
				if err != nil {
					http.Error(w, "invalid trusted subnet", http.StatusInternalServerError)
					return
				}

				if !subnet.Contains(ip) {
					http.Error(w, "forbidden: IP not in trusted subnet", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
