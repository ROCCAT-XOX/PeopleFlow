# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PeopleFlow is an HR management system built with Go (backend) and HTML/Tailwind CSS with a newer Astro-based frontend under development. The system enables comprehensive employee management, document handling, time tracking, and reporting functionalities.

## Development Commands

### Running the Application

#### Backend (Go)
```bash
# Run directly with Go
go run main.go

# Run with automatic reloading during development
air
```

#### Frontend (Astro)
```bash
# Navigate to Astro frontend directory
cd frontend-astro

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

### Docker Deployment

```bash
# Run with Docker Compose
docker-compose up -d

# Deploy using the deployment script
./deploy.sh

# With custom parameters
MONGODB_PORT=27017 APP_PORT=8080 IMAGE_TAG=latest PLATFORM=linux/amd64 ./deploy.sh
```

### Database

The application uses MongoDB. When developing locally, you can run it in a container:

```bash
# Run MongoDB container
docker run -d --name mongodb \
  --network peopleflow-network \
  -p 27017:27017 \
  -v mongodb_data:/data/db \
  --restart unless-stopped \
  mongo:latest
```

## Architecture Overview

### Backend Structure

- **Router (`backend/router.go`)**: Central routing configuration for all endpoints
- **Handlers (`backend/handler/`)**: HTTP handlers that process requests and responses
- **Models (`backend/model/`)**: Data structures and business logic
- **Repositories (`backend/repository/`)**: Database access and queries
- **Services (`backend/service/`)**: Business logic and external integrations
- **Middleware (`backend/middleware/`)**: Auth and access control middleware
- **Background (`backend/background/`)**: Background worker for periodic tasks

### Authentication and Authorization

- JWT-based authentication system
- Role-based access control with four levels:
  - Admin: Full system access
  - Manager: Employee, document, and report management
  - HR: Employee and document management
  - User: Access to own data only

### External Integrations

The system includes integrations with:
- Timebutler: For time tracking and absence management
- 123erfasst: For project tracking and time entries

### Frontend Architecture

- Traditional HTML/CSS/JS frontend with Tailwind CSS in `frontend/`
- New Astro-based frontend in development in `frontend-astro/`

## Common Development Workflows

### Adding a New API Endpoint

1. Create or update a handler in `backend/handler/`
2. Register the route in `backend/router.go`
3. Apply appropriate middleware for authentication and access control

### Adding a New Feature to the Frontend

1. Create or update the template in `frontend/templates/`
2. Add any required JavaScript in `frontend/static/js/`
3. Add any static assets in `frontend/static/`

### Accessing the Application

After starting the application, it's available at http://localhost:8080

Default admin login:
- Email: admin@PeopleFlow.com
- Password: admin