# MapleApps Token Refresh Prototype

A React + Vite prototype demonstrating the background token refresh system extracted from the `maplefile-login` project. This standalone application showcases automatic token refresh using Web Workers and encrypted token storage.

## 🚀 Features

- **Background Token Refresh**: Automatic token refresh using Web Workers
- **Encrypted Token Storage**: Secure token storage in localStorage
- **Cross-Tab Communication**: BroadcastChannel for multi-tab synchronization
- **Real-time Monitoring**: Live status updates and activity logging
- **Demo Mode**: Test tokens and scenarios for development

## 🏗️ Architecture

### Core Components

1. **Token Refresh Worker** (`public/token-refresh-worker.js`)
   - Runs in background Web Worker
   - Checks tokens every 30 seconds
   - Refreshes tokens 5 minutes before expiry
   - Handles API communication

2. **Worker Manager** (`src/services/tokenRefreshWorkerManager.js`)
   - Manages worker lifecycle
   - Bridges main thread and worker
   - Handles message routing

3. **Storage Service** (`src/services/tokenStorageService.js`)
   - Encrypted token storage
   - Token expiry tracking
   - Authentication state management

4. **React App** (`src/App.jsx`)
   - User interface for monitoring
   - Controls for testing
   - Real-time activity logging

## 📋 Prerequisites

- Node.js (v18 or higher)
- npm or yarn
- Backend API server (for token refresh endpoint)

## 🛠️ Installation

1. **Clone or create the project directory:**
   ```bash
   mkdir maplefile-refreshtoken
   cd maplefile-refreshtoken
   ```

2. **Create the project files:**
   - Copy all the provided files into their respective directories
   - Ensure the directory structure matches the file paths

3. **Install dependencies:**
   ```bash
   npm install
   ```

## 🚀 Usage

### Development Server

```bash
npm run dev
```

This starts the development server on `http://localhost:3001`

### Build for Production

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## 🔧 Configuration

### Backend API

The app expects a token refresh endpoint at:
```
POST /iam/api/v1/token/refresh
```

Update the proxy configuration in `vite.config.js`:
```javascript
proxy: {
  "/iam/api": {
    target: "http://your-backend-url", // Update this
    changeOrigin: true,
    secure: false,
  },
}
```

### Token Refresh API

**Request Format:**
```json
{
  "value": "encrypted_refresh_token"
}
```

**Response Format (Success - 201):**
```json
{
  "encrypted_access_token": "base64_encrypted_token",
  "encrypted_refresh_token": "base64_encrypted_token",
  "token_nonce": "base64_nonce",
  "access_token_expiry_date": "2024-01-15T11:00:00Z",
  "refresh_token_expiry_date": "2024-01-29T10:30:00Z",
  "username": "user@example.com"
}
```

## 🎮 Testing the System

### 1. Create Demo Tokens
- Click "🎭 Create Demo Tokens" to generate test tokens
- Creates tokens with 30-minute access token expiry
- Starts background monitoring automatically

### 2. Test Automatic Refresh
- Click "⏰ Create Expiring Soon Tokens"
- Creates tokens that expire in 3 minutes
- Watch the system automatically refresh them

### 3. Manual Operations
- **Manual Refresh**: Force an immediate token refresh
- **Force Token Check**: Trigger an immediate status check
- **Clear Tokens**: Remove all tokens and stop monitoring

### 4. Monitor Activity
- Real-time activity log shows all system events
- Worker status panel displays current state
- Token information panel shows expiry times

## 📊 Monitoring

### Worker Status
- **Initialized**: Whether the Web Worker is running
- **Refreshing**: If a refresh operation is in progress
- **Last Check**: Timestamp of the last token check
- **Check Interval**: How often tokens are checked (30s)

### Token Information
- **Encrypted Tokens**: Whether tokens are stored
- **Authentication**: Overall auth status
- **Token Format**: Separate vs legacy format
- **Expiry Status**: Time remaining for each token type

## 🔐 Security Features

- **Encrypted Storage**: All tokens stored encrypted in localStorage
- **Cross-Tab Sync**: Changes propagated across browser tabs
- **Automatic Cleanup**: Tokens cleared on refresh failure
- **Force Logout**: User logged out when refresh token expires

## 🛠️ Development

### Project Structure
```
maplefile-refreshtoken/
├── public/
│   ├── token-refresh-worker.js     # Web Worker for background refresh
│   └── vite.svg                    # Vite icon
├── src/
│   ├── services/
│   │   ├── tokenRefreshWorkerManager.js  # Worker management
│   │   └── tokenStorageService.js        # Token storage
│   ├── App.jsx                     # Main React component
│   ├── App.css                     # App styles
│   ├── main.jsx                    # React entry point
│   └── index.css                   # Global styles
├── package.json
├── vite.config.js
├── eslint.config.js
└── README.md
```

### Key Differences from Original

This prototype extracts only the token refresh functionality:

- **Removed**: Login flow, crypto operations, full auth system
- **Simplified**: Storage service, worker manager
- **Added**: Demo token creation, real-time monitoring UI
- **Focused**: Pure token refresh testing and demonstration

### Customization

1. **Refresh Timing**: Modify `CHECK_INTERVAL` and `REFRESH_THRESHOLD` in the worker
2. **Storage Keys**: Update `STORAGE_KEYS` if using different key names
3. **API Endpoint**: Change the endpoint URL in the worker and vite config
4. **Token Format**: Supports both separate and legacy token formats

## 🐛 Troubleshooting

### Worker Not Initializing
- Check browser console for errors
- Ensure `token-refresh-worker.js` is in the `public` folder
- Verify Web Worker support in your browser

### API Connection Issues
- Check the backend server is running
- Verify the proxy configuration in `vite.config.js`
- Check network tab for failed requests

### Token Refresh Failures
- Ensure the backend endpoint returns the correct format
- Check that refresh tokens are valid
- Verify the API accepts the expected request format

## 📝 Logging

The system provides comprehensive logging:

- **Console Logs**: Detailed technical information
- **Activity Log**: User-friendly event tracking
- **Worker Messages**: Inter-thread communication
- **Storage Changes**: Token storage operations

## 🔄 Integration

To integrate this into your own project:

1. Copy the worker file to your public directory
2. Include the worker manager and storage service
3. Initialize the worker in your app
4. Set up event listeners for token events
5. Configure your API endpoints

## 📄 License

This prototype is part of the MapleApps project and follows the same licensing terms.

## 🤝 Contributing

This is a prototype for demonstration purposes. For production use, consider:

- Adding proper error boundaries
- Implementing retry logic
- Adding telemetry and monitoring
- Enhancing security measures
- Adding comprehensive tests
