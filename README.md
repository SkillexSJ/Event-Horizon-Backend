<div align="center">
  <img
    src="https://capsule-render.vercel.app/api?type=soft&height=300&color=gradient&&customColorList=6&text=Event%20Horizon&descAlign=50&animation=twinkling&textBg=false&reversal=false&fontSize=100&fontAlign=51&fontAlignY=57&desc=Discover%20Your%20Next%20Next%20Event&descAlignY=69"
    alt="Event Horizon"
  />
  <h2>Backend Service</h2>
  <p><strong>Robust Scalable Event Management API</strong></p>
  <p>The core engine powering the <strong>Event Horizon</strong> platform. Built with <strong>Go</strong> and <strong>Echo</strong>, this service handles authentication, event scheduling, ticket inventory management, and secure booking transactions with MongoDB.</p>
</div>

<div align="center">
    
  [![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
  [![Echo](https://img.shields.io/badge/Echo-v4-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://echo.labstack.com/)
  [![MongoDB](https://img.shields.io/badge/MongoDB-6.0-47A248.svg?style=for-the-badge&logo=mongodb&logoColor=white)](https://www.mongodb.com/)
  [![JWT](https://img.shields.io/badge/JWT-Auth-000000.svg?style=for-the-badge&logo=jsonwebtokens&logoColor=white)](https://jwt.io/)
  [![Deployed on Heroku](https://img.shields.io/badge/Deployed%20on-Heroku-430098.svg?style=for-the-badge&logo=heroku&logoColor=white)](https://www.heroku.com/)
</div>

<br />

## ✨ Features

- **🔐 Secure Authentication**: JWT-based auth with secure HttpOnly cookies.
- **📅 Event Management**: Complete CRUD for events with start/end times & location data.
- **🎟️ Smart Ticketing**:
  - Multiple ticket types (VIP, Regular, Student).
  - Real-time inventory tracking (`TotalQuantity` vs `AvailableQuantity`).
- **💳 Transactional Bookings**: ACID transactions for booking integrity (pending/confirmed/cancelled statuses).
- **👥 Role-Based Access**: Distinct `User` and `Host` capabilities.


## 🛠️ Tech Stack

<div align="center">

|                                              Core                                               |                                                 Database                                                  |                                                 Infra                                                  |
| :---------------------------------------------------------------------------------------------: | :-------------------------------------------------------------------------------------------------------: | :----------------------------------------------------------------------------------------------------: |
| <img src="https://skillicons.dev/icons?i=go" width="48" height="48" alt="Go" /><br/>**Go 1.25** | <img src="https://skillicons.dev/icons?i=mongodb" width="48" height="48" alt="MongoDB" /><br/>**MongoDB** | <img src="https://skillicons.dev/icons?i=heroku" width="48" height="48" alt="Heroku" /><br/>**Heroku** |

</div>

## 🗺️ Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API (Echo)
    participant D as MongoDB

    C->>A: POST /auth/login
    A->>D: Find User
    D-->>A: User Data
    A->>A: Validate Password (Bcrypt)
    A->>A: Generate JWT
    A-->>C: Set HttpOnly Cookie (token)
```

## 🛠️ API Reference

| Resource     | Method | Endpoint               | Description       | Access           |
| :----------- | :----- | :--------------------- | :---------------- | :--------------- |
| **Auth**     | POST   | `/auth/register`       | Register new user | Public           |
|              | POST   | `/auth/login`          | Login user        | Public           |
| **Events**   | GET    | `/events/all`          | List all events   | Public           |
|              | POST   | `/events/create`       | Create new event  | Protected (Host) |
|              | DELETE | `/events/:id`          | Delete event      | Protected (Host) |
| **Bookings** | POST   | `/bookings/create`     | Book a ticket     | Protected        |
|              | GET    | `/bookings/user`       | User's bookings   | Protected        |
|              | PUT    | `/bookings/:id/cancel` | Cancel booking    | Protected        |

## 🚀 Getting Started

### Prerequisites

- **Go 1.25+**
- **MongoDB** (Local or Atlas)

### 1. Clone & Setup

```bash
git clone https://github.com/yourusername/event-horizon-backend.git
cd Backend
go mod download
```

### 2. Environment Configuration

Create a `.env` file in the root directory:

```env
MONGO_URI=mongodb+srv://<user>:<password>@cluster.mongodb.net/?retryWrites=true&w=majority
PORT=3000
DATABASE_NAME=EventHorizonDB
JWT_SECRET=your-super-secret-jwt-key
```

### 3. Run Locally

```bash
go run main.go
```

The server will start at `http://localhost:3000`.

## 🌐 Deployment (Heroku)

This backend is optimized for Heroku deployment.

1.  **Create App**: `heroku create event-horizon-api`
2.  **Set Env Vars**:
    ```bash
    heroku config:set MONGO_URI="your_mongo_uri"
    heroku config:set JWT_SECRET="your_secret"
    ```
3.  **Deploy**:
    ```bash
    git push heroku main
    ```
