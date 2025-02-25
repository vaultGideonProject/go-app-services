# VaultMessenger Go Services

VaultMessenger's backend services built with Go, integrating Firebase Firestore, Cloud Tasks, and Gin for high-performance REST APIs.

## 🚀 Features

- **Real-time Messaging**: Retrieve and send chat messages via Firestore.
- **Conversation Management**: Handle user conversations with proper ordering and pagination.
- **Push Notifications**: Cloud Tasks integration to schedule and send notifications.
- **Secure Firebase Authentication**: Ensuring secure access to Firestore data.
- **RESTful API**: Built using Gin for clean, efficient routing.

## 📦 Tech Stack

- **Go**: Main programming language.
- **Firebase**: Firestore for real-time data storage.
- **Cloud Tasks**: Schedule and execute background tasks.
- **Gin**: Lightweight web framework.
- **Google Cloud Platform**: Hosting and service orchestration.

## 📁 Project Structure

```
.
├── main.go                   # Entry point
├── firebaseConfig/           # Contains google-services.json
├── handlers/                 # API handlers (get messages, send notifications, etc.)
├── go.mod                    # Go module dependencies
├── go.sum                    # Dependency checksums
└── README.md                 # Project documentation
```

## 🔧 Installation

1. **Clone the repository:**

```bash
git clone https://github.com/yourusername/vaultmessenger-go.git
cd vaultmessenger-go
```

2. **Set up Firebase credentials:**

Ensure your Firebase service account JSON file is placed at:

```
firebaseConfig/google-services.json
```

3. **Install dependencies:**

```bash
go mod tidy
```

4. **Run the server:**

```bash
go run main.go
```

The API will run on `http://localhost:8080`

## 🛠️ API Endpoints

### Get Messages

```http
GET /get-messages/:userId/:receiverId
```

- **userId**: ID of the sender.
- **receiverId**: ID of the receiver.

**Response:**

```json
{
  "success": true,
  "message": [
    { "text": "Hello", "timestamp": "2024-02-24T12:00:00Z" }
  ]
}
```

### Get Conversations

```http
POST /get-conversations
```

**Request Body:**

```json
{
  "data": { "userId": "abc123" }
}
```

### Send Message

```http
POST /send-message
```

**Request Body:**

```json
{
  "senderUid": "abc123",
  "receiverUid": "xyz789",
  "message": { "text": "Hi there!" }
}
```

### Send Notification Task

```http
POST /send-notification-task
```

**Request Body:**

```json
{
  "token": "device_token",
  "title": "New Message",
  "body": "You have a new message",
  "imageUrl": "https://example.com/image.png",
  "userId": "abc123"
}
```

## 🌐 Deployment

To deploy, use Google Cloud Run or App Engine, ensuring Firebase and Cloud Tasks APIs are enabled.

```bash
gcloud run deploy vaultmessenger-go --source .
```

## 📚 Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## 🔒 License

This project is licensed under the MIT License.

---

Feel free to reach out if you have any questions about VaultMessenger's Go services! ✨

