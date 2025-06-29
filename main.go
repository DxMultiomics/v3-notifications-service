package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub" // Import Pub/Sub SDK
	"github.com/joho/godotenv"
)

const (
	// projectID is automatically detected by the client library when running on GCP.
	topicID        = "notifications-topic"
	subscriptionID = "notifications-topic-sub"
)

// NotificationPayload defines the structure of our notification data
type NotificationPayload struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// EventBroadcaster manages and broadcasts events to connected SSE clients
type EventBroadcaster struct {
	// Channel to receive new events that need to be broadcasted (now fed by Pub/Sub)
	Notifier chan NotificationPayload
	// Channel to indicate new client connections
	newClients chan chan NotificationPayload
	// Channel to indicate client disconnections
	closingClients chan chan NotificationPayload
	// Map of currently connected clients, where each client has its own channel
	clients map[chan NotificationPayload]bool
}

// NewEventBroadcaster creates and returns an initialized EventBroadcaster
func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		Notifier:       make(chan NotificationPayload, 100), // Buffered for bursts
		newClients:     make(chan chan NotificationPayload),
		closingClients: make(chan chan NotificationPayload),
		clients:        make(map[chan NotificationPayload]bool),
	}
}

// Start method listens for new events, client connections/disconnections
// This should be run as a goroutine
func (b *EventBroadcaster) Start() {
	for {
		select {
		case s := <-b.newClients: // New client connected
			b.clients[s] = true
			log.Println("Added new SSE client")
		case s := <-b.closingClients: // Client disconnected
			delete(b.clients, s)
			close(s) // Close the client's channel
			log.Println("Removed SSE client")
		case event := <-b.Notifier: // New notification event to broadcast
			// Send the event to all active clients (will block briefly if client's buffer is full)
			for clientChan := range b.clients {
				// Direct send. This ensures the message is put into the buffered channel.
				// If the channel is full, this send will block until Space is available,
				// or until the client eventually becomes active.
				// For extreme cases, you might want a timeout on this send or check len(clientChan)
				clientChan <- event
			}
		}

	}
}

// ServeHTTP implements http.Handler for the SSE endpoint
func (b *EventBroadcaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// CORS headers for Remix client (IMPORTANT: Adjust for production)
	w.Header().Set("Access-Control-Allow-Origin", "http://dxmultiomics.com")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Get a channel to notify when the client disconnects
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel() // Ensure cancel is called when handler exits

	// Make sure the ResponseWriter is a Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a dedicated channel for this client
	clientChan := make(chan NotificationPayload, 5) // Use a small buffer, e.g., 5 messages
	b.newClients <- clientChan                      // Register this client

	// Send an initial message to signify connection
	fmt.Fprintf(w, "data: %s\n\n", "Connected to Go SSE Service")
	flusher.Flush()

	// Listen for events or client disconnection
	for {
		select {
		case <-ctx.Done(): // Client disconnected
			b.closingClients <- clientChan // Unregister this client
			log.Println("SSE client disconnected from Go service.")
			return
		case event := <-clientChan: // Received an event to send to this client
			jsonEvent, err := json.Marshal(event)
			if err != nil {
				log.Printf("Error marshaling event for client: %v", err)
				continue
			}
			// --- ADD THIS LINE HERE ---
			log.Printf("GO_SERVICE_WRITE: Attempting to write event to SSE client: %s\n", string(jsonEvent))
			// --- END ADDED LINE ---
			fmt.Fprintf(w, "data: %s\n\n", jsonEvent)
			flusher.Flush() // Send the data immediately
		}
	}
}

// PubSubPublisher handles publishing messages to a Pub/Sub topic
func PubSubPublisher(w http.ResponseWriter, r *http.Request, pubSubClient *pubsub.Client) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	notification := NotificationPayload{
		ID:        fmt.Sprintf("notification-%d", time.Now().UnixNano()),
		Type:      reqBody.Type,
		Message:   reqBody.Message,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Marshal the notification payload to JSON bytes for Pub/Sub
	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshaling notification for Pub/Sub: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get a reference to the topic
	t := pubSubClient.Topic(topicID)
	// Publish the message
	result := t.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	// Block until the publish operation completes
	id, err := result.Get(ctx)
	if err != nil {
		log.Printf("Failed to publish message to Pub/Sub: %v", err)
		http.Error(w, "Failed to publish notification", http.StatusInternalServerError)
		return
	}

	log.Printf("Published notification message (ID: %s) to Pub/Sub topic %s", id, topicID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Notification published to Pub/Sub\n")
}

// PubSubReceiver listens for messages on a Pub/Sub subscription and pushes them to the broadcaster
func PubSubReceiver(ctx context.Context, pubSubClient *pubsub.Client, broadcaster *EventBroadcaster) {
	sub := pubSubClient.Subscription(subscriptionID)

	log.Printf("Listening for messages on Pub/Sub subscription %s...", subscriptionID)

	// Receive messages
	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Received Pub/Sub message ID: %s", msg.ID)
		var notification NotificationPayload
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			log.Printf("Error unmarshaling Pub/Sub message data: %v", err)
			msg.Ack() // Acknowledge even on error to prevent reprocessing bad messages
			return
		}

		broadcaster.Notifier <- notification // Push to the broadcaster's internal channel
		log.Printf("Forwarded Pub/Sub message to SSE broadcaster: %v", notification)
		msg.Ack() // Acknowledge the message to remove it from the subscription
	})

	if err != nil && err != context.Canceled {
		log.Fatalf("Pub/Sub Receive failed: %v", err)
	}
	log.Println("Stopped listening on Pub/Sub subscription.")
}

func main() {
	// Root context for application lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Initialize Pub/Sub client
	// When running on GCP (like Cloud Run), the project ID is automatically
	// detected from the environment. Pass an empty string for the projectID.
	// For local development, you might need to set the GOOGLE_CLOUD_PROJECT env var.
	projectID := os.Getenv("GCP_PROJECT_ID")

	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer pubSubClient.Close() // Close client when main exits

	broadcaster := NewEventBroadcaster()
	go broadcaster.Start() // Start the event broadcaster as a goroutine

	// Start Pub/Sub receiver in a goroutine
	go PubSubReceiver(ctx, pubSubClient, broadcaster)

	// SSE Endpoint for clients (e.g., your Remix server `notificationEmitter.server.ts`)
	http.Handle("/sse/notifications", broadcaster)

	// REST Endpoint to trigger a notification (now publishes to Pub/Sub)
	// Use a closure to pass the pubSubClient to the handler
	http.HandleFunc("/api/send-notification", func(w http.ResponseWriter, r *http.Request) {
		PubSubPublisher(w, r, pubSubClient)
	})

	// Get port from environment for Cloud Run
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090" // Default for local development
	}
	listenAddr := fmt.Sprintf(":%s", port)

	fmt.Printf("Go Notification Service running on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
