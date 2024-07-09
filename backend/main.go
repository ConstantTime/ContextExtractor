package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rakshittiwari/smart-context-extractor/backend/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sashabaranov/go-openai"
)

var db *sql.DB

func main() {

    connStr := "postgres://rakshit:rakshit@localhost/smart_context_extractor?sslmode=disable"
    if !strings.Contains(connStr, "sslmode=") {
        connStr += "?sslmode=disable"
    }
	
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error opening database connection: ", err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Fatal("Error pinging database: ", err)
    }

    fmt.Println("Successfully connected to the database!")

    // Router setup
    r := mux.NewRouter()

    // Routes
    r.HandleFunc("/api/tickets", createTicketHandler).Methods("POST")
    r.HandleFunc("/api/tickets", getTicketsHandler).Methods("GET")
    r.HandleFunc("/api/tickets/{id}", getTicketHandler).Methods("GET")
    r.HandleFunc("/api/context-rules", getContextRulesHandler).Methods("GET")
    r.HandleFunc("/api/context-rules", createContextRuleHandler).Methods("POST")

	err = godotenv.Load("conf.env")
	if err != nil {
		fmt.Println("Unable to load the env file")
	}

	fmt.Println("Server starting!")
    // Start server
    log.Fatal(http.ListenAndServe(":8080", r))
}

func createTicketHandler(w http.ResponseWriter, r *http.Request) {
    var ticket models.Ticket
    err := json.NewDecoder(r.Body).Decode(&ticket)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Extract context using LLM
    context := extractContext(ticket.Content)

    // Set the initial status based on context extraction result
    if context == "No context in content" {
        ticket.Status = "context_demanded"
        // Draft an email for missing context
        emailDraft := draftEmailForMissingContext(ticket)
        json.NewEncoder(w).Encode(map[string]string{
            "status": ticket.Status,
            "email_draft": emailDraft,
        })
        // Save ticket with status and empty context
        _, err = db.Exec("INSERT INTO tickets (customer_id, subject, content, context, status) VALUES ($1, $2, $3, $4, $5)",
            ticket.CustomerID, ticket.Subject, ticket.Content, "", ticket.Status)
        if err != nil {
            log.Printf("Error saving ticket: %v", err)
            http.Error(w, "Error saving ticket", http.StatusInternalServerError)
        }
        return
    }

    // Context was successfully extracted
    ticket.Status = "escalated"
    ticket.Context = context

    // Save ticket with extracted context and status
    err = db.QueryRow("INSERT INTO tickets (customer_id, subject, content, context, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
        ticket.CustomerID, ticket.Subject, ticket.Content, ticket.Context, ticket.Status).Scan(&ticket.ID)
    if err != nil {
        log.Printf("Error saving ticket: %v", err)
        http.Error(w, "Error saving ticket", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(ticket)
}

func getTicketsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, customer_id, subject, content, context, status FROM tickets")
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    tickets := []models.Ticket{}
    for rows.Next() {
        var t models.Ticket
        if err := rows.Scan(&t.ID, &t.CustomerID, &t.Subject, &t.Content, &t.Context, &t.Status); err != nil {
            log.Printf("Error scanning ticket: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        tickets = append(tickets, t)
    }

    // Check for errors from iterating over rows
    if err := rows.Err(); err != nil {
        log.Printf("Error iterating over rows: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(tickets)
}

func getTicketHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var t models.Ticket
    err := db.QueryRow("SELECT id, customer_id, subject, content, context FROM tickets WHERE id = $1", id).
        Scan(&t.ID, &t.CustomerID, &t.Subject, &t.Content, &t.Context)
    
    if err != nil {
        if err == sql.ErrNoRows {
            // No ticket found with the given ID
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{"error": "Ticket not found"})
            return
        }
        // For any other database error
        log.Printf("Database error: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(t)
}

func getContextRulesHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, keywords FROM context_rules")
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var rules []models.ContextRule
    for rows.Next() {
        var rule models.ContextRule
        if err := rows.Scan(&rule.ID, &rule.Name, &rule.Keywords); err != nil {
            log.Printf("Error scanning context rule: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        rules = append(rules, rule)
    }

    // Check for errors from iterating over rows
    if err := rows.Err(); err != nil {
        log.Printf("Error iterating over rows: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")

    if len(rules) == 0 {
        // No rules found
        w.WriteHeader(http.StatusOK) 
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "No context rules found",
            "rules":   []models.ContextRule{}, 
        })
        return
    }


    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Context rules retrieved successfully",
        "count":   len(rules),
        "rules":   rules,
    })
}

func createContextRuleHandler(w http.ResponseWriter, r *http.Request) {
    var rule models.ContextRule
    err := json.NewDecoder(r.Body).Decode(&rule)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = db.QueryRow("INSERT INTO context_rules (name, keywords) VALUES ($1, $2) RETURNING id",
        rule.Name, rule.Keywords).Scan(&rule.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(rule)
}

func extractContext(content string) string {
    rows, err := db.Query("SELECT name, keywords FROM context_rules")
    if err != nil {
        log.Printf("Error fetching context rules: %v", err)
        return "No context in content"
    }
    defer rows.Close()

    var rules []models.ContextRule
    for rows.Next() {
        var rule models.ContextRule
        if err := rows.Scan(&rule.Name, &rule.Keywords); err != nil {
            log.Printf("Error scanning context rule: %v", err)
            return "No context in content"
        }
        rules = append(rules, rule)
    }

    prompt := `You are an AI assistant designed to extract specific context metadata from customer support tickets. Your task is to identify and extract unique identifiers or relevant metadata based on predefined keywords.

Here are the rules for context extraction:
`
    for _, rule := range rules {
        prompt += fmt.Sprintf("- If the content contains keywords like '%s', extract any relevant unique identifiers or metadata related to these keywords.\n", rule.Keywords)
    }

    prompt += `
Your response should ONLY include the extracted metadata. If no relevant metadata is found, respond with "No context in content".

Examples:
1. Input: "My transaction #1234 failed yesterday. Can you help?"
   Output: Transaction ID: 1234

2. Input: "I can't find my insurance details. My customer ID is ABC123."
   Output: Customer ID: ABC123

3. Input: "I'm having trouble logging in. Can you reset my password?"
   Output: No context in content

Now, extract the relevant context metadata from this support ticket:
` + content

    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT3Dot5Turbo,
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleSystem,
                    Content: prompt,
                },
            },
        },
    )

    if err != nil {
        log.Printf("ChatCompletion error: %v\n", err)
        return "No context in content"
    }

    extractedContext := resp.Choices[0].Message.Content
	fmt.Println(extractedContext)
    if extractedContext == "No context in content" {
        return "No context in content"
    }

    return extractedContext
}

func draftEmailForMissingContext(ticket models.Ticket) string {
    return fmt.Sprintf(`
Subject: More Information Needed for Your Support Request

Dear Customer,

We received your support request regarding "%s". To assist you better, we need some additional information.

Please reply to this email with the following details:
1. [Specific information request based on ticket category]
2. Any relevant documents or screenshots

Thank you for your cooperation.

Best regards,
Customer Support Team
`, ticket.Subject)
}