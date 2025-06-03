package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/vitistack/datacenter-operator/internal/clients"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AuthMiddleware is a middleware that validates Kubernetes tokens.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, "Unauthorized: Unable to check the token", http.StatusUnauthorized)
			return
		}

		if !validateKubernetesToken(token) {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractToken extracts the token from the Authorization header.
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if strings.HasPrefix(bearerToken, "Bearer ") {
		return strings.TrimPrefix(bearerToken, "Bearer ")
	}
	return ""
}

// validateKubernetesToken validates the token using the Kubernetes API.
func validateKubernetesToken(token string) bool {
	clientset := clients.Kubernetes

	// Create a TokenReview request
	tokenReview := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token: token,
		},
	}

	// Submit TokenReview request to Kubernetes API
	result, err := clientset.AuthenticationV1().TokenReviews().Create(context.TODO(), tokenReview, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating TokenReview: %v\n", err)
		return false
	}

	// Check if the token is valid
	if result.Status.Authenticated {
		return true
	}
	return false
}
