package http

import "net/http"

func MiddleAuth(next http.Handler, secrets []string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        url := r.URL
        if url.Path == "/health/" || url.Path == "/health" {
            next.ServeHTTP(w, r)
            return
        }
        secret := r.Header.Get("secret")
        for _, s := range secrets {
            if s == secret {
                next.ServeHTTP(w, r)
                return
            }
        }
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
    })
}
