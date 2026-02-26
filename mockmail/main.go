package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Email struct {
	ID        int       `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	emails []Email
	mu     sync.Mutex
	nextID = 1
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>MockMail - Flutterize</title>
    <meta charset="UTF-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', system-ui, sans-serif; background: #1a1a2e; color: #e0e0e0; }
        .header { background: #16213e; padding: 20px 30px; border-bottom: 2px solid #0f3460; }
        .header h1 { color: #e94560; font-size: 24px; }
        .header p { color: #888; font-size: 14px; margin-top: 4px; }
        .container { max-width: 900px; margin: 30px auto; padding: 0 20px; }
        .stats { display: flex; gap: 15px; margin-bottom: 20px; }
        .stat { background: #16213e; padding: 15px 20px; border-radius: 8px; border: 1px solid #0f3460; }
        .stat .num { font-size: 28px; font-weight: bold; color: #e94560; }
        .stat .label { font-size: 12px; color: #888; text-transform: uppercase; }
        .email { background: #16213e; border: 1px solid #0f3460; border-radius: 8px; margin-bottom: 12px; padding: 18px; }
        .email:hover { border-color: #e94560; }
        .email .meta { display: flex; justify-content: space-between; margin-bottom: 8px; }
        .email .from { color: #e94560; font-weight: 600; }
        .email .to { color: #aaa; }
        .email .time { color: #666; font-size: 13px; }
        .email .subject { font-size: 16px; font-weight: 600; margin-bottom: 8px; color: #fff; }
        .email .body { color: #bbb; font-size: 14px; white-space: pre-wrap; background: #0a0a1a; padding: 12px; border-radius: 4px; }
        .empty { text-align: center; padding: 60px; color: #666; }
        .refresh { color: #e94560; text-decoration: none; cursor: pointer; }
        .toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📧 MockMail</h1>
        <p>SMTP :2525 — Flutterize Mail Simulator</p>
    </div>
    <div class="container">
        <div class="toolbar">
            <div class="stats">
                <div class="stat"><div class="num">{{len .}}</div><div class="label">Emails</div></div>
            </div>
            <a class="refresh" href="/" onclick="location.reload()">↻ Refresh</a>
        </div>
        {{if .}}
            {{range .}}
            <div class="email">
                <div class="meta">
                    <span><span class="from">{{.From}}</span> → <span class="to">{{.To}}</span></span>
                    <span class="time">{{.Timestamp.Format "2006-01-02 15:04:05"}}</span>
                </div>
                <div class="subject">{{.Subject}}</div>
                <div class="body">{{.Body}}</div>
            </div>
            {{end}}
        {{else}}
            <div class="empty">
                <p>No emails received yet.</p>
                <p style="margin-top:8px;font-size:13px;">Waiting for SMTP connections on port 2525...</p>
            </div>
        {{end}}
    </div>
</body>
</html>`

func main() {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	// Start SMTP server
	go startSMTP()

	// Web UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(emails)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Show newest first
		reversed := make([]Email, len(emails))
		for i, e := range emails {
			reversed[len(emails)-1-i] = e
		}
		tmpl.Execute(w, reversed)
	})

	log.Println("MockMail Web UI running on :8025")
	log.Fatal(http.ListenAndServe(":8025", nil))
}

func startSMTP() {
	listener, err := net.Listen("tcp", ":2525")
	if err != nil {
		log.Fatalf("Failed to start SMTP server: %v", err)
	}
	log.Println("MockMail SMTP server running on :2525")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("SMTP accept error: %v", err)
			continue
		}
		go handleSMTP(conn)
	}
}

func handleSMTP(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	writeLine := func(msg string) {
		fmt.Fprintf(writer, "%s\r\n", msg)
		writer.Flush()
	}

	writeLine("220 mockmail ESMTP MockMail")

	var from, to, subject, body string
	inData := false
	var dataLines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n")

		if inData {
			if line == "." {
				inData = false
				// Parse headers from data
				headersDone := false
				var bodyLines []string
				for _, dl := range dataLines {
					if !headersDone {
						if dl == "" {
							headersDone = true
							continue
						}
						upper := strings.ToUpper(dl)
						if strings.HasPrefix(upper, "SUBJECT:") {
							subject = strings.TrimSpace(dl[8:])
						}
					} else {
						bodyLines = append(bodyLines, dl)
					}
				}
				body = strings.Join(bodyLines, "\n")

				mu.Lock()
				emails = append(emails, Email{
					ID:        nextID,
					From:      from,
					To:        to,
					Subject:   subject,
					Body:      body,
					Timestamp: time.Now(),
				})
				nextID++
				mu.Unlock()

				log.Printf("Email received: from=%s to=%s subject=%s", from, to, subject)
				writeLine("250 OK")
				dataLines = nil
				continue
			}
			dataLines = append(dataLines, line)
			continue
		}

		upper := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(upper, "HELO"), strings.HasPrefix(upper, "EHLO"):
			writeLine("250 mockmail")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			from = extractEmail(line)
			writeLine("250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			to = extractEmail(line)
			writeLine("250 OK")
		case upper == "DATA":
			writeLine("354 Start mail input; end with <CRLF>.<CRLF>")
			inData = true
		case upper == "QUIT":
			writeLine("221 Bye")
			return
		case strings.HasPrefix(upper, "RSET"):
			from = ""
			to = ""
			writeLine("250 OK")
		case strings.HasPrefix(upper, "NOOP"):
			writeLine("250 OK")
		default:
			writeLine("500 Command not recognized")
		}
	}
}

func extractEmail(line string) string {
	start := strings.Index(line, "<")
	end := strings.Index(line, ">")
	if start >= 0 && end > start {
		return line[start+1 : end]
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return line
}
