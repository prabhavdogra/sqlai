package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// QueryRequest is the payload received by our Go server.
type QueryRequest struct {
	Query string `json:"query"`
}

// OllamaResponse models the assumed response from Ollama.
type OllamaResponse struct {
	GeneratedText string `json:"generated_text"`
}

// QueryResponse is the final response returned to the client.
type QueryResponse struct {
	SQL     string        `json:"sql"`
	Results []interface{} `json:"results"`
}

var db *sql.DB

func main() {
	// Build database connection string from environment variables.
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Verify connectivity.
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to the database.")

	// Set up HTTP handler.
	http.HandleFunc("/query", queryHandler)

	log.Println("Go server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// queryHandler handles incoming POST requests to /query.
func queryHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get the SQL query from the Ollama service (Llama).
	sqlQuery, err := getSQLQueryFromOllama(req.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling Ollama service: %v", err), http.StatusInternalServerError)
		return
	}

	// Run the generated SQL query against the Postgres DB.
	results, err := runSQLQuery(sqlQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error running SQL query: %v", err), http.StatusInternalServerError)
		return
	}

	resp := QueryResponse{
		SQL:     sqlQuery,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// getSQLQueryFromOllama calls the Ollama service to convert a natural language query into SQL.
func getSQLQueryFromOllama(prompt string) (string, error) {
	// Adjust URL according to your Docker Compose configuration.
	url := "http://localhost:11434/api/generate"

	// Create payload for the Ollama service.
	payload := map[string]interface{}{
		"model":  "sqlcoder", // Change this if your Ollama model name differs.
		"prompt": fmt.Sprintf("Translate the following natural language query into an SQL query: %s", prompt),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and check response.
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama service error: status %d, response %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}
	return ollamaResp.GeneratedText, nil
}

// runSQLQuery executes the generated SQL query and collects the results.
func runSQLQuery(query string) ([]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []interface{}

	for rows.Next() {
		// Prepare a slice for the columns.
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Map column names to their values.
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			var v interface{}
			if b, ok := columnValues[i].([]byte); ok {
				v = string(b)
			} else {
				v = columnValues[i]
			}
			rowMap[colName] = v
		}
		results = append(results, rowMap)
	}

	return results, nil
}
