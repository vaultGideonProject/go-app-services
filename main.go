package main

import (
	_ "cloud.google.com/go/cloudtasks/apiv2"
	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	_ "firebase.google.com/go"
	_ "firebase.google.com/go/auth"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	_ "google.golang.org/api/tasks/v1"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"io"
	"log"
	"net/http"
)

// Firebase Configs
const (
	FirestoreCollection    = "conversations"
	FirestoreMessageSubCol = "messages"
	ProjectID              = "vaultmessengerdev"
	LocationID             = "europe-west3"
	QueueID                = "notificationsDefault"
	QueueURL               = "https://services.vaultmessenger.com/send-notification-task"
)

var (
	firestoreClient *firestore.Client
	tasksClient     *cloudtasks.Client
)

func initFirebase() error {
	ctx := context.Background()

	opt := option.WithCredentialsFile("/Users/supernova/IdeaProjects/goservices/firebaseConfig/google-services.json")

	// Initialize Firestore client
	var err error
	firestoreClient, err = firestore.NewClient(ctx, ProjectID, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firestore client: %v", err)
	}

	err = firestoreClient.Close()
	if err != nil {
		return err
	}

	// Initialize Cloud Tasks client
	tasksClient, err = cloudtasks.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloud Tasks client: %v", err)
	}

	return nil
}

func main() {
	if err := initFirebase(); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	defer func(firestoreClient *firestore.Client) {
		err := firestoreClient.Close()
		if err != nil {

		}
	}(firestoreClient)

	router := gin.Default()

	// API routes
	router.GET("/get-messages/:userId/:receiverId", getMessagesHandler)
	router.POST("/get-conversations", getConversationsHandler)
	router.POST("/send-message", sendMessageHandler)
	router.POST("/send-notification-task", sendNotificationTaskHandler)

	// Start the server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Get Conversations Handler
func getMessagesHandler(c *gin.Context) {
	userId := c.Param("userId")
	receiverId := c.Param("receiverId")
	ctx := context.Background()

	// Query Firestore
	collection := firestoreClient.Collection(FirestoreCollection).Doc(userId).Collection(receiverId)
	docs, err := collection.OrderBy("timestamp", firestore.Asc).Limit(100).Documents(ctx).GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve messages: %v", err)})
		return
	}

	// Parse conversations and map to the expected format
	messages := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		data := doc.Data()
		messages[i] = data
	}

	// Send the conversations data inside a wrapper object
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": messages, // Send conversations directly in the response body
	})
}

// Get Conversations Handler
func getConversationsHandler(c *gin.Context) {
	// Log the raw request body to ensure userId is being sent
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
	// Parse the userId from the request body (now expecting the userId inside a data object)
	var requestBody struct {
		Data struct {
			UserId string `json:"userId"` // Define the UserId inside a data object
		} `json:"data"` // Indicate that the root field is "data"
	}

	// Unmarshal the body into the struct
	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		// Log the error details and return a 400 Bad Request
		fmt.Printf("HTTP 400 Bad Request: Invalid request body, error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if userId is empty
	if requestBody.Data.UserId == "" {
		// Log the empty userId error and return a 400 Bad Request
		fmt.Printf("HTTP 400 Bad Request: Missing UserId in request body\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserId is required"})
		return
	}

	// Extract the userId
	userId := requestBody.Data.UserId
	ctx := context.Background()

	// Ensure Firestore path is correctly constructed
	collection := firestoreClient.Collection(FirestoreCollection).Doc(userId).Collection(FirestoreMessageSubCol)

	// Retrieve documents, order by timestamp and limit to 100
	docs, err := collection.OrderBy("timestamp", firestore.Asc).Limit(100).Documents(ctx).GetAll()
	if err != nil {
		// Log the error with more details
		fmt.Printf("HTTP 500 Internal Server Error: Failed to retrieve conversations for userId: %s, error: %v\n", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve conversations: %v", err)})
		return
	}

	// Parse conversations and map to the expected format
	conversations := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		data := doc.Data()
		conversations[i] = data
	}

	// Send the conversations data inside a wrapper object
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": conversations, // Send conversations directly in the response body
	})
}

// Send Message Handler
func sendMessageHandler(c *gin.Context) {
	var payload struct {
		SenderUid   string                 `json:"senderUid"`
		ReceiverUid string                 `json:"receiverUid"`
		Message     map[string]interface{} `json:"message"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	ctx := context.Background()

	// Add message to Firestore
	senderMessages := firestoreClient.Collection(FirestoreCollection).Doc(payload.SenderUid).Collection(payload.ReceiverUid)
	receiverMessages := firestoreClient.Collection(FirestoreCollection).Doc(payload.ReceiverUid).Collection(payload.SenderUid)

	_, err := senderMessages.NewDoc().Set(ctx, payload.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save message for sender: %v", err)})
		return
	}

	_, err = receiverMessages.NewDoc().Set(ctx, payload.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save message for receiver: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Message sent successfully"})
}

// Send Notification Task Handler
func sendNotificationTaskHandler(c *gin.Context) {
	var payload struct {
		Token    string `json:"token"`
		Title    string `json:"title"`
		Body     string `json:"body"`
		ImageUrl string `json:"imageUrl"`
		UserId   string `json:"userId"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	ctx := context.Background()

	// Create a Cloud Task
	task := &taskspb.Task{
		MessageType: &taskspb.Task_HttpRequest{
			HttpRequest: &taskspb.HttpRequest{
				HttpMethod: taskspb.HttpMethod_POST,
				Url:        QueueURL,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte(fmt.Sprintf(`{"token":"%s","title":"%s","body":"%s","imageUrl":"%s","userId":"%s"}`,
					payload.Token, payload.Title, payload.Body, payload.ImageUrl, payload.UserId)),
			},
		},
	}

	// Schedule task execution
	parent := fmt.Sprintf("projects/%s/locations/%s/queues/%s", ProjectID, LocationID, QueueID)
	if _, err := tasksClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
		Parent: parent,
		Task:   task,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create notification task: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Notification task created"})
}
