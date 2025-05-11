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
	"strings"

	_ "github.com/lib/pq"
)

var ollamaHost = "https://8766-2a09-bac1-36a0-1b8-00-39-138.ngrok-free.app"

// QueryRequest is the payload received by our Go server.
type QueryRequest struct {
	Query string `json:"query"`
}

// OllamaResponse models the assumed response from Ollama.
type OllamaResponse struct {
	GeneratedText string `json:"response"`
	Query         string `json:"query"`
}

// QueryResponse is the final response returned to the client.
type QueryResponse struct {
	SQLQuery string        `json:"sql_query"`
	Results  []interface{} `json:"results"`
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

	log.Println("Go server listening on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// queryHandler handles incoming POST requests to /query.
func queryHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self' *:8081")

	// Only allow POST.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusOK)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get the SQL query from the Ollama service (Llama).
	sqlQuery, err := getSQLQueryFromGPT(req.Query)
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
		SQLQuery: sqlQuery,
		Results:  results,
	}

	json.NewEncoder(w).Encode(resp)
}

func getSQLQueryFromGPT(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	url := "https://api.openai.com/v1/chat/completions"

	userPrompt := fmt.Sprintf(`You are a helpful assistant that writes SQL queries for PostgreSQL.

Given the schema: %s

Write only the SQL query to answer the question: "%s"

Do not include any explanations, just the query.`, schema, prompt)

	// Build the request payload
	payload := struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}{
		Model: "gpt-3.5-turbo",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "system", Content: "You are a SQL assistant."},
			{Role: "user", Content: userPrompt},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	// Clean the result (remove any leading/trailing markdown or code formatting)
	raw := result.Choices[0].Message.Content
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```sql")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	return clean, nil
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
