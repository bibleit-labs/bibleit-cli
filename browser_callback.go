package main

import (
	"errors"
	"html"
	"io"
	"net/http"
)

type browserAuthorizationResult struct {
	code string
	err  error
}

func browserAuthorizationCallback(state string, result chan<- browserAuthorizationResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}
		query := r.URL.Query()
		if values := query["state"]; len(values) != 1 || values[0] != state {
			http.Error(w, "Invalid authentication state.", http.StatusBadRequest)
			return
		}
		if values, denied := query["error"]; denied {
			message := "Browser sign-in was denied. Return to your browser for details."
			if len(values) == 1 {
				switch values[0] {
				case "account_blocked":
					message = "Browser sign-in was denied because your account is blocked."
				case "access_denied":
					message = "Browser sign-in was denied."
				case "authorization_expired":
					message = "Browser sign-in expired; run bibleit auth login again."
				}
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			_, _ = io.WriteString(w, "<!doctype html><html lang=\"en\"><meta charset=\"utf-8\"><title>Sign-in denied · Bibleit</title><h1>Sign-in wasn't completed.</h1><p>"+html.EscapeString(message)+"</p><p>Return to your terminal. You can close this tab.</p></html>")
			select {
			case result <- browserAuthorizationResult{err: authenticationError{message: message}}:
			default:
			}
			return
		}
		codes := query["code"]
		if len(codes) != 1 || codes[0] == "" {
			http.Error(w, "Authorization code is missing.", http.StatusBadRequest)
			select {
			case result <- browserAuthorizationResult{err: errors.New("authorization code is missing")}:
			default:
			}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = io.WriteString(w, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>CLI connected · Bibleit</title><style>body{margin:0;min-height:100vh;display:grid;place-items:center;background:#090f1a;color:#f8f3e8;font:16px system-ui,sans-serif}.card{max-width:34rem;margin:2rem;padding:3rem;border:1px solid #344154;border-radius:20px;background:#111a29;box-shadow:0 24px 80px #0008}.mark{color:#f4b942;font-size:2rem}h1{font-size:2.2rem;margin:.8rem 0}p{color:#bac5d5;line-height:1.6}</style></head><body><main class="card"><div class="mark">✦</div><h1>Your CLI is connected now.</h1><p>Authentication completed successfully. Feel free to close this tab and return to your terminal.</p></main></body></html>`)
		select {
		case result <- browserAuthorizationResult{code: codes[0]}:
		default:
		}
	}
}
