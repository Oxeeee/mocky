# 🎭 mocky

<div align="center">
  <h3>Powerful HTTP Mock Server in Go</h3>
  <p>Easily replace HTTP responses from APIs with your own for testing and development</p>

  ![Go](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat-square&logo=go)
  ![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
  ![Build](https://img.shields.io/badge/Build-Passing-brightgreen?style=flat-square)
</div>

---

## 📖 Table of Contents

- [🚀 Quick Start](#-quick-start)
- [✨ Features](#-features)
- [🌐 WebUI & API](#-webui--api)
- [💡 Usage Examples](#-usage-examples)
- [📁 Project Structure](#-project-structure)
- [⚙️ Technical Requirements](#️-technical-requirements)
- [🚀 Deployment](#-deployment)

---

## 🚀 Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/Oxeeee/mocky
cd mocky

# 2. Start the server
go run cmd/main.go

# 3. Open the web interface
open http://localhost:8082/__mock/ui
```

🎉 **Done!** Your mock server is running and ready to use!

---

## ✨ Features

<table>
<tr>
<td>

### 🎯 Core Functions
- 🔧 **Create HTTP mocks** with support for headers, body, statusCode
- ✏️ **Edit** existing mocks
- 🗑️ **Delete** unnecessary mocks
- 📋 **View** all active mocks

</td>
<td>

### 🌟 Additional
- 🖥️ **Beautiful web interface**
- 🔄 **REST API** for automation
- 🚀 **Easy deployment**
- 📱 **Responsive design**

</td>
</tr>
</table>

---

## 🌐 WebUI & API

### 🖥️ Web UI

After starting the server, the web interface is available at:
```
🌍 http://localhost:8082/__mock/ui
```

**Interface capabilities:**

| Feature | Description |
|---------|-------------|
| ✅ **View** | All active mocks in a convenient format |
| ➕ **Add** | New HTTP mocks through an intuitive form |
| ✏️ **Edit** | Modify existing mocks |
| 🗑️ **Delete** | Remove unnecessary mocks |
| 🔄 **Refresh** | Update the mock list |
| 👀 **Details** | Toggle full content display |

### 🔌 API Endpoints

#### ![GET](https://img.shields.io/badge/GET-4CAF50?style=flat-square) `/__mock/list`
Get a list of all active mocks

```bash
curl -X GET http://localhost:8082/__mock/list
```

<details>
<summary>📋 Response example</summary>

```json
{
  "/api/users": {
    "GET": {
      "status_code": 200,
      "headers": {"Content-Type": "application/json"},
      "body": "{\"users\": []}"
    }
  }
}
```
</details>

#### ![POST](https://img.shields.io/badge/POST-2196F3?style=flat-square) `/__mock/add`
Add a new mock or update an existing one

<details>
<summary>📝 Request body</summary>

```json
{
  "method": "GET",
  "path": "/api/users",
  "response": {
    "status_code": 200,
    "headers": {
      "Content-Type": "application/json"
    },
    "body": "{\"users\": [{\"id\": 1, \"name\": \"John\"}]}"
  }
}
```
</details>

```bash
curl -X POST http://localhost:8082/__mock/add \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/users",
    "response": {
      "status_code": 200,
      "headers": {"Content-Type": "application/json"},
      "body": "{\"users\": []}"
    }
  }'
```

#### ![DELETE](https://img.shields.io/badge/DELETE-F44336?style=flat-square) `/__mock/delete`
Delete an existing mock

<details>
<summary>📝 Request body</summary>

```json
{
  "method": "GET",
  "path": "/api/users"
}
```
</details>

```bash
curl -X DELETE http://localhost:8082/__mock/delete \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/users"
  }'
```

### 🎯 Using Mocks
After creating a mock, all requests to the specified path will return the defined response:

```bash
curl -X GET http://localhost:8082/api/users
# → Will return the mocked response
```

---

## 💡 Usage Examples

### 🟢 Creating a Simple GET Mock

```bash
curl -X POST http://localhost:8082/__mock/add \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/health",
    "response": {
      "status_code": 200,
      "headers": {"Content-Type": "application/json"},
      "body": "{\"status\": \"ok\", \"timestamp\": \"2024-01-01T12:00:00Z\"}"
    }
  }'
```

### 🔴 Creating a POST Mock with Error

```bash
curl -X POST http://localhost:8082/__mock/add \
  -H "Content-Type: application/json" \
  -d '{
    "method": "POST",
    "path": "/api/users",
    "response": {
      "status_code": 400,
      "headers": {"Content-Type": "application/json"},
      "body": "{\"error\": \"Validation failed\", \"message\": \"Email is required\"}"
    }
  }'
```

### 🔧 Creating a Mock with CORS Headers

```bash
curl -X POST http://localhost:8082/__mock/add \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "path": "/api/data",
    "response": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE",
        "X-API-Version": "v1.0"
      },
      "body": "{\"data\": [1, 2, 3], \"total\": 3}"
    }
  }'
```

---

## 📁 Project Structure

```
🎭 asdf/
├── 📂 cmd/
│   └── 🏗️ main.go          # Main server file
├── 📂 templates/
│   └── 🎨 index.html       # Web interface
└── 📝 README.md            # Documentation
```

**Component description:**

- **`cmd/main.go`** - HTTP server with routing and API handlers
- **`templates/index.html`** - SPA interface with JavaScript for mock management
- **`README.md`** - Project documentation

---

## ⚙️ Technical Requirements

| Component | Requirement |
|-----------|-------------|
| **🔧 Go** | version 1.16 or higher |
| **🚪 Port** | 8082 (default) |
| **📦 Dependencies** | only Go standard library |
| **💾 Storage** | in-memory |
| **🌐 Browser** | any modern browser |

---

## 🚀 Deployment

### 🏠 Local Deployment

> **💡 Tip:** When working with corporate networks, it's recommended to use VK Tunnel instead of ngrok

When trying to access the mock server through ngrok from a corporate network, you might encounter a `DNS PROBE FINISHED NXDOMAIN` error. This is due to corporate network settings blocking addresses like `https://*.ngrok-free.app`.

**Alternative solution:** [VK Tunnel](https://dev.vk.com/ru/libraries/tunnel) by Russian developers.

### 📋 Step-by-step Instructions

1. **Clone and start the server:**
   ```bash
   git clone <repository-url>
   cd asdf
   go run cmd/main.go
   ```

2. **Set up tunnel (optional):**
   ```bash
   vk-tunnel --insecure=1 \
            --http-protocol=http \
            --ws-protocol=ws \
            --host=localhost \
            --port=8082 \
            --timeout=5000
   ```

3. **Done! 🎉**
   - **Local access:** `http://localhost:8082/__mock/ui`
   - **Through tunnel:** `https://{host}/__mock/ui`

---

<div align="center">
  <h3>🎭 Ready for mocking!</h3>
  <p>Create, test, and debug your APIs with ease</p>

  ⭐ **Star this project if you find it useful!**
</div>