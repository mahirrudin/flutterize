package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

type SMS struct {
	ID        int       `json:"id"`
	Phone     string    `json:"phone"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type SendRequest struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

var (
	messages []SMS
	mu       sync.Mutex
	nextID   = 1
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>MockSMS - Flutterize</title>
    <meta charset="UTF-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', system-ui, sans-serif; background: #1a1a2e; color: #e0e0e0; }
        .header { background: #16213e; padding: 20px 30px; border-bottom: 2px solid #0f3460; }
        .header h1 { color: #00d2ff; font-size: 24px; }
        .header p { color: #888; font-size: 14px; margin-top: 4px; }
        .container { max-width: 900px; margin: 30px auto; padding: 0 20px; }
        .stats { display: flex; gap: 15px; margin-bottom: 20px; }
        .stat { background: #16213e; padding: 15px 20px; border-radius: 8px; border: 1px solid #0f3460; }
        .stat .num { font-size: 28px; font-weight: bold; color: #00d2ff; }
        .stat .label { font-size: 12px; color: #888; text-transform: uppercase; }
        .sms { background: #16213e; border: 1px solid #0f3460; border-radius: 8px; margin-bottom: 12px; padding: 18px; }
        .sms:hover { border-color: #00d2ff; }
        .sms .meta { display: flex; justify-content: space-between; margin-bottom: 8px; }
        .sms .phone { color: #00d2ff; font-weight: 600; font-size: 16px; }
        .sms .time { color: #666; font-size: 13px; }
        .sms .message { color: #bbb; font-size: 14px; white-space: pre-wrap; background: #0a0a1a; padding: 12px; border-radius: 4px; }
        .empty { text-align: center; padding: 60px; color: #666; }
        .refresh { color: #00d2ff; text-decoration: none; cursor: pointer; }
        .toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📱 MockSMS</h1>
        <p>HTTP :7600 — Flutterize SMS Simulator</p>
    </div>
    <div class="container">
        <div class="toolbar">
            <div class="stats">
                <div class="stat"><div class="num">{{len .}}</div><div class="label">Messages</div></div>
            </div>
            <a class="refresh" href="/" onclick="location.reload()">↻ Refresh</a>
        </div>
        {{if .}}
            {{range .}}
            <div class="sms">
                <div class="meta">
                    <span class="phone">{{.Phone}}</span>
                    <span class="time">{{.Timestamp.Format "2006-01-02 15:04:05"}}</span>
                </div>
                <div class="message">{{.Message}}</div>
            </div>
            {{end}}
        {{else}}
            <div class="empty">
                <p>No SMS messages received yet.</p>
                <p style="margin-top:8px;font-size:13px;">Waiting for POST /send requests...</p>
            </div>
        {{end}}
    </div>
</body>
</html>`

func main() {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(messages)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		reversed := make([]SMS, len(messages))
		for i, m := range messages {
			reversed[len(messages)-1-i] = m
		}
		tmpl.Execute(w, reversed)
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req SendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		if req.Phone == "" || req.Message == "" {
			http.Error(w, `{"error":"phone and message are required"}`, http.StatusBadRequest)
			return
		}

		mu.Lock()
		sms := SMS{
			ID:        nextID,
			Phone:     req.Phone,
			Message:   req.Message,
			Timestamp: time.Now(),
		}
		messages = append(messages, sms)
		nextID++
		mu.Unlock()

		log.Printf("SMS received: phone=%s message=%s", req.Phone, req.Message)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sms)
	})

	log.Println("MockSMS running on :7600")
	log.Fatal(http.ListenAndServe(":7600", nil))
}
