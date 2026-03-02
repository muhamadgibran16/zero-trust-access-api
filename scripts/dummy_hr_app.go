package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		
		userID := r.Header.Get("X-ZTA-User-ID")
		authStatus := r.Header.Get("X-ZTA-Authenticated")
		
		html := fmt.Sprintf(`
		<html>
			<head>
				<title>HR App (Dummy)</title>
				<style>
					body { font-family: system-ui, sans-serif; padding: 40px; background: #f0fdf4; color: #166534; }
					.card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
					.secure { display: inline-block; background: #22c55e; color: white; padding: 4px 8px; border-radius: 4px; font-weight: bold; font-size: 12px; margin-bottom: 20px;}
				</style>
			</head>
			<body>
				<div class="card">
					<div class="secure">🔒 Protected by Zero Trust</div>
					<h1>HR Information System</h1>
					<p>Welcome to the internal HR placeholder application.</p>
					<hr>
					<h3>Headers injected by ZTA Proxy:</h3>
					<ul>
						<li><b>X-ZTA-Authenticated:</b> %s</li>
						<li><b>X-ZTA-User-ID:</b> %s</li>
					</ul>
					<p><small>If you can see this page, the Reverse Proxy tunnel is working successfully!</small></p>
				</div>
			</body>
		</html>
		`, authStatus, userID)
		
		fmt.Fprint(w, html)
	})

	log.Println("Dummy HR App listening on :9090...")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
