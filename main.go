package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	// Import codes package
	// Import status package
	notificationpb "dxmultiomics.com/service-notification/notificationpb"
	// "github.com/improbable-eng/grpc-web/go/grpcweb" // REMOVE THIS IMPORT
)

// --- Constants ---
const (
	defaultPort = "8090" // Default for local development (Go backend)
)

// --- NotificationPayload (internal struct for Pub/Sub unmarshaling)---
type NotificationPayload struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// --- EventBroadcaster ---
type EventBroadcaster struct {
	Notifier       chan *notificationpb.Notification
	newClients     chan chan *notificationpb.Notification
	closingClients chan chan *notificationpb.Notification
	clients        map[chan *notificationpb.Notification]bool
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		Notifier:       make(chan *notificationpb.Notification, 100),
		newClients:     make(chan chan *notificationpb.Notification),
		closingClients: make(chan chan *notificationpb.Notification),
		clients:        make(map[chan *notificationpb.Notification]bool),
	}
}

func (b *EventBroadcaster) Start() {
	for {
		select {
		case s := <-b.newClients:
			b.clients[s] = true
			slog.Info("Added new gRPC streaming client")
		case s := <-b.closingClients:
			delete(b.clients, s)
			close(s)
			slog.Info("Removed gRPC streaming client")
		case event := <-b.Notifier:
			for clientChan := range b.clients {
				// Non-blocking send or drop if client's buffer is full
				select {
				case clientChan <- event:
				default:
					slog.Warn("Dropped event for a client: channel full", "reason", "slow_consumer")
				}
			}
		}
	}
}

// --- PubSubReceiver ---
func PubSubReceiver(ctx context.Context, pubSubClient *pubsub.Client, broadcaster *EventBroadcaster, subscriptionID string) {
	sub := pubSubClient.Subscription(subscriptionID)
	// Add a check to ensure the subscription exists at startup. This provides a clear error message.
	exists, err := sub.Exists(ctx)
	if err != nil {
		slog.Error("Failed to check for Pub/Sub subscription existence", "error", err)
		os.Exit(1)
	}
	if !exists {
		slog.Error("Pub/Sub subscription does not exist. Please create it.", "subscription_id", subscriptionID)
		os.Exit(1)
	}

	slog.Info("Listening for messages on Pub/Sub subscription", "subscription_id", subscriptionID)

	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		slog.Info("Received Pub/Sub message", "message_id", msg.ID)
		var notificationPayload NotificationPayload
		if err := json.Unmarshal(msg.Data, &notificationPayload); err != nil {
			slog.Error("Error unmarshaling Pub/Sub message data", "error", err, "message_id", msg.ID)
			msg.Ack()
			return
		}

		protoNotification := &notificationpb.Notification{
			Id:        notificationPayload.ID,
			Type:      notificationPayload.Type,
			Message:   notificationPayload.Message,
			Timestamp: notificationPayload.Timestamp,
		}

		broadcaster.Notifier <- protoNotification
		slog.Info("Forwarded Pub/Sub message to gRPC broadcaster", "message_id", protoNotification.Id)
		msg.Ack()
	})

	if err != nil && err != context.Canceled {
		slog.Error("Pub/Sub Receive failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Stopped listening on Pub/Sub subscription.")
}

// --- gRPC Notification Service Implementation ---
type notificationService struct {
	notificationpb.UnimplementedNotificationServiceServer
	broadcaster *EventBroadcaster
}

func (s *notificationService) StreamNotifications(req *notificationpb.StreamNotificationsRequest, stream notificationpb.NotificationService_StreamNotificationsServer) error {
	slog.Info("Client connected for StreamNotifications", "user_id", req.GetUserId())

	if req.GetUserId() == "" {
		// During development, you might allow anonymous connections.
		// In production, this should be a hard error after implementing real authentication.
		slog.Warn("Connecting client did not provide a user_id.")
	}

	clientChan := make(chan *notificationpb.Notification, 5)
	s.broadcaster.newClients <- clientChan

	defer func() {
		slog.Info("Client disconnected from StreamNotifications. Unregistering.", "user_id", req.GetUserId())
		s.broadcaster.closingClients <- clientChan
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case notification := <-clientChan:
			if err := stream.Send(notification); err != nil {
				slog.Error("Failed to send notification to client. Closing stream.", "user_id", req.GetUserId(), "error", err)
				return err
			}
			slog.Info("Sent notification to client", "user_id", req.GetUserId(), "message_id", notification.GetId())
		}
	}
}

// --- Main function: Sets up pure gRPC server ---
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		slog.Info("Could not load .env file, using environment variables", "error", err)
	}

	// --- Configuration from Environment Variables (Best Practice) ---
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		slog.Error("GCP_PROJECT_ID environment variable not set. This is required.")
		os.Exit(1)
	}

	subscriptionID := os.Getenv("PUBSUB_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		slog.Error("PUBSUB_SUBSCRIPTION_ID environment variable not set. This is required.")
		os.Exit(1)
	}

	// Get port from environment for Cloud Run or local dev
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	listenAddr := fmt.Sprintf(":%s", port)

	// --- Initialize Clients ---
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		slog.Error("Failed to create Pub/Sub client", "error", err)
		os.Exit(1)
	}
	defer pubSubClient.Close()

	// --- Start Application Components ---
	broadcaster := NewEventBroadcaster()
	go broadcaster.Start()

	go PubSubReceiver(ctx, pubSubClient, broadcaster, subscriptionID)

	// --- Setup pure gRPC server (no HTTP wrapper here) ---
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		slog.Error("Failed to listen on address", "address", listenAddr, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(grpcServer, &notificationService{broadcaster: broadcaster})

	// --- Start server and handle graceful shutdown ---
	go func() {
		slog.Info("Go gRPC Notification Service starting", "address", listenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	// This is a standard pattern for robust services.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received

	slog.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	slog.Info("gRPC server stopped.")
	// The main context cancel() is already deferred, which will signal PubSubReceiver to stop.
}
