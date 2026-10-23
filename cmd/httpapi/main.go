//
//  peterplanner
//  Copyright (c) 2026 Nicolas Gomollon <nicolas@gomollon.me>
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Request struct {
	Operation string `json:"operation"`
	Value     string `json:"value"`
	JSON      bool   `json:"json"`
}

var (
	uidPattern       = regexp.MustCompile(`(?i)^[a-f0-9]{40}$`)
	studentIDPattern = regexp.MustCompile(`^[0-9]{1,16}$`)
)

func main() {
	// Restrict permissions on newly created reports.
	syscall.Umask(0077)

	secret, err := os.ReadFile("/run/secrets/peterplanner_api_token")
	if err != nil {
		log.Fatal("Cannot read API authentication secret")
	}
	expectedToken := sha256.Sum256([]byte(strings.TrimSpace(string(secret))))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	mux.HandleFunc("POST /run", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		provided := sha256.Sum256([]byte(token))

		if !strings.HasPrefix(auth, "Bearer ") ||
			subtle.ConstantTimeCompare(provided[:], expectedToken[:]) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

		var req Request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		var binary string
		var args []string
		contentType := "application/json"

		switch req.Operation {
		case "uid":
			if !uidPattern.MatchString(req.Value) {
				http.Error(w, "Invalid UID", http.StatusBadRequest)
				return
			}
			binary = "/app/peterplanner"
			args = []string{"--uid", req.Value, "--json"}

		case "studentID":
			if !studentIDPattern.MatchString(req.Value) {
				http.Error(w, "Invalid student ID", http.StatusBadRequest)
				return
			}
			binary = "/app/peterplanner"
			args = []string{"--studentID", req.Value}
			if req.JSON {
				args = append(args, "--json")
			} else {
				contentType = "text/plain; charset=utf-8"
			}

		case "cookie", "download":
			if len(req.Value) == 0 || len(req.Value) > 8192 {
				http.Error(w, "Invalid cookie", http.StatusBadRequest)
				return
			}

			binary = "/app/peterplanner"
			if req.Operation == "download" {
				binary = "/app/peterplanner-download"
				contentType = "text/xml; charset=utf-8"
			}
			args = []string{"--cookie", req.Value}

		default:
			http.Error(w, "Unknown operation", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()

		// Arguments are passed directly without invoking a shell.
		cmd := exec.CommandContext(ctx, binary, args...)
		output, err := cmd.Output()

		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				http.Error(w, "Backend timeout", http.StatusGatewayTimeout)
			} else {
				log.Printf("Operation %s failed", req.Operation)
				http.Error(w, "Backend error", http.StatusBadGateway)
			}
			return
		}

		if len(output) > 64<<20 {
			http.Error(w, "Response too large", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Write(output)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      150 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Println("PeterPlanner API listening on :8080")
	log.Fatal(server.ListenAndServe())
}
