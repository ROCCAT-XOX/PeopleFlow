# PeopleFlow HR Management System

<div align="center">
  <img src="frontend/static/images/PeopleFlow-Logo-Symbol.svg" alt="PeopleFlow Logo" width="200"/>
  
  **A comprehensive HR management system built with Go and modern web technologies**
</div>

## 🚀 Overview

PeopleFlow is a full-featured HR management system that provides comprehensive employee management, document handling, time tracking, and reporting functionalities. The system features a robust backend built with Go and offers both traditional HTML/CSS frontend and a modern Astro-based frontend under development.

## ✨ Key Features

- **Employee Management**: Complete employee lifecycle management with detailed profiles
- **Time Tracking**: Integration with external time tracking services (Timebutler, 123erfasst)
- **Document Management**: Secure document storage and handling
- **Absence Management**: Holiday and absence tracking with calendar views
- **Reporting & Analytics**: Comprehensive statistics and reporting capabilities
- **Role-Based Access Control**: Four-tier permission system (Admin, Manager, HR, Employee)
- **JWT Authentication**: Secure authentication with token-based sessions

## 🏗️ Architecture

### Backend (Go)
- **Framework**: Gin web framework with middleware for authentication and logging
- **Database**: MongoDB with structured document storage
- **Authentication**: JWT-based with role-based access control
- **Logging**: Structured logging with `slog` package for comprehensive monitoring
- **Testing**: Comprehensive test suite with 90%+ coverage for critical components

### Frontend
- **Current**: Traditional HTML/CSS/JavaScript with Tailwind CSS
- **Future**: Astro-based modern frontend (in development)

### External Integrations
- **Timebutler**: Time tracking and absence management
- **123erfasst**: Project tracking and time entries

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- MongoDB 5.0 or higher
- Node.js 18+ (for Astro frontend)
- Docker (optional)

### Running with Go

```bash
# Clone the repository
git clone <repository-url>
cd PeopleFlow

# Run the backend
go run main.go

# Or with automatic reloading during development
air
```

### Running with Docker

```bash
# Start all services
docker-compose up -d

# Or use the deployment script
./deploy.sh

# With custom parameters
MONGODB_PORT=27017 APP_PORT=8080 IMAGE_TAG=latest PLATFORM=linux/amd64 ./deploy.sh
```

### Astro Frontend (Development)

```bash
# Navigate to Astro frontend
cd frontend-astro

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

## 🗄️ Database Setup

The application automatically creates a MongoDB database. For local development:

```bash
# Run MongoDB container
docker run -d --name mongodb \
  --network peopleflow-network \
  -p 27017:27017 \
  -v mongodb_data:/data/db \
  --restart unless-stopped \
  mongo:latest
```

## 🔐 Default Login

After starting the application, access it at `http://localhost:8080`

**Default Admin Credentials:**
- Email: `admin@peopleflow.com`
- Password: `admin`

## 🧪 Testing

The application includes a comprehensive test suite covering models, repositories, middleware, and core functionality.

### Running All Tests

```bash
# Run complete test suite
go test -v ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running Specific Test Categories

```bash
# Model tests (validation, business logic)
go test -v ./backend/model/...

# Repository tests (database operations)
go test -v ./backend/repository/...

# Middleware tests (authentication, authorization)
go test -v ./backend/middleware/...
```

### Running Individual Test Functions

```bash
# User model validation tests
go test -v ./backend/model -run TestUser

# Authentication middleware tests
go test -v ./backend/middleware -run TestAuth

# Password hashing and validation
go test -v ./backend/model -run TestUserPasswordHashing
```

### Performance Benchmarks

```bash
# Run benchmark tests
go test -bench=. ./backend/model/...

# Run specific benchmarks
go test -bench=BenchmarkUser ./backend/model/
```

### Test Coverage Analysis

```bash
# Generate detailed coverage report
go test -coverprofile=coverage.out ./backend/model/...
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Test Examples

**Model Validation Tests:**
```bash
$ go test -v ./backend/model -run TestUserValidation
=== RUN   TestUserValidation
=== RUN   TestUserValidation/valid_user
=== RUN   TestUserValidation/invalid_email
=== RUN   TestUserValidation/empty_first_name
--- PASS: TestUserValidation (0.00s)
    --- PASS: TestUserValidation/valid_user (0.00s)
    --- PASS: TestUserValidation/invalid_email (0.00s)
    --- PASS: TestUserValidation/empty_first_name (0.00s)
```

**Password Security Tests:**
```bash
$ go test -v ./backend/model -run TestUserPasswordHashing
=== RUN   TestUserPasswordHashing
--- PASS: TestUserPasswordHashing (0.15s)
    user_test.go:145: Password hashed successfully
    user_test.go:149: Original password cleared: true
    user_test.go:155: Password verification successful
```

## 🏛️ Code Architecture

### Recent Improvements

This version includes significant architectural improvements implemented with comprehensive testing and logging:

#### 🔍 Structured Logging
- **Implementation**: Complete structured logging with Go's `slog` package
- **Features**: Context-aware logging, performance monitoring, request tracking
- **Coverage**: All middleware, repositories, and critical business operations

#### 🗃️ Repository Pattern
- **Base Repository**: Abstract base with common database operations
- **Error Handling**: Comprehensive error wrapping and context preservation
- **Performance**: Built-in performance monitoring and query optimization
- **Testing**: Full test coverage with mocked database operations

#### 🔐 Enhanced Authentication & Authorization
- **JWT Security**: Improved token validation with expiration handling
- **Role-Based Access**: Four-tier permission system with granular controls
- **Backward Compatibility**: Seamless support for existing user passwords
- **Middleware Chain**: Comprehensive auth middleware with request logging

#### 📊 Model Validation
- **Input Validation**: Comprehensive field validation for all models
- **Business Rules**: Built-in business logic validation
- **Security**: Password strength requirements and secure hashing
- **Type Safety**: Strong typing with custom validation errors

### Project Structure

```
backend/
├── handler/          # HTTP request handlers
├── middleware/       # Authentication and logging middleware
├── model/           # Business models with validation
├── repository/      # Data access layer with base patterns
├── service/         # Business logic and external integrations
├── utils/           # Utilities (JWT, crypto, logging)
└── db/              # Database connection management

frontend/            # Traditional HTML/CSS/JS frontend
frontend-astro/      # Modern Astro-based frontend (development)
```

### Authentication Flow

1. **Login Request**: User submits credentials via web form or API
2. **Password Validation**: Backward-compatible password checking (legacy + new hashing)
3. **JWT Generation**: Secure token creation with user roles and expiration
4. **Session Management**: Cookie-based sessions for web, header-based for API
5. **Request Authorization**: Middleware validates tokens and enforces role permissions

### Role-Based Permissions

| Role | Permissions |
|------|-------------|
| **Admin** | Full system access, user management, system settings |
| **Manager** | Employee management, reports, document access |
| **HR** | Employee data management, absence tracking |
| **Employee** | Personal data access, time tracking, document viewing |

## 🔧 Development

### Adding New Features

1. **Models**: Add to `backend/model/` with comprehensive validation
2. **Repository**: Extend base repository pattern in `backend/repository/`
3. **Handlers**: Create HTTP handlers in `backend/handler/`
4. **Routes**: Register in `backend/router.go` with appropriate middleware
5. **Tests**: Add comprehensive tests for all components

### Testing Strategy

The application follows a comprehensive testing approach:

- **Unit Tests**: Individual function and method testing
- **Integration Tests**: Database and service integration testing
- **Validation Tests**: Input validation and business rule testing
- **Security Tests**: Authentication and authorization testing
- **Performance Tests**: Benchmark testing for critical operations

### Code Quality

- **Coverage Target**: 90%+ for critical business logic
- **Validation**: Comprehensive input validation on all models
- **Security**: Secure password hashing, JWT validation, role-based access
- **Logging**: Structured logging with performance monitoring
- **Error Handling**: Comprehensive error wrapping and context preservation

## 📚 API Documentation

### Authentication Endpoints

```
POST /login          # User authentication
POST /logout         # Session termination
```

### User Management

```
GET  /users          # List all users (Admin/Manager)
POST /users          # Create new user (Admin/HR)
PUT  /users/:id      # Update user (Admin/HR/Self)
DELETE /users/:id    # Delete user (Admin only)
```

### Employee Management

```
GET  /employees      # List employees
POST /employees      # Create employee
PUT  /employees/:id  # Update employee
GET  /employees/:id  # Employee details
```

## 🔒 Security Features

- **Password Security**: bcrypt hashing with backward compatibility
- **JWT Tokens**: Secure token-based authentication with expiration
- **Role-Based Access**: Granular permission system
- **Input Validation**: Comprehensive validation on all inputs
- **SQL Injection Protection**: MongoDB with proper query construction
- **Session Management**: Secure cookie handling

## 🚀 Deployment

### Production Deployment

```bash
# Build for production
go build -o peopleflow main.go

# Run with environment variables
export MONGODB_URI="mongodb://localhost:27017"
export JWT_SECRET="your-secret-key"
export APP_PORT="8080"
./peopleflow
```

### Docker Deployment

```bash
# Build and deploy
docker-compose up -d

# View logs
docker-compose logs -f peopleflow
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Add** comprehensive tests for your changes
4. **Ensure** all tests pass (`go test ./...`)
5. **Commit** your changes (`git commit -m 'Add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Code Standards

- **Testing**: All new code must include comprehensive tests
- **Logging**: Use structured logging for all operations
- **Validation**: Implement input validation for all models
- **Documentation**: Update documentation for API changes
- **Security**: Follow security best practices

## 📊 Test Coverage Report

Current test coverage by component:

| Component | Coverage | Status |
|-----------|----------|--------|
| Models | 92% | ✅ Excellent |
| Middleware | 85% | ✅ Good |
| Repositories | 75% | ⚠️ Improving |
| Handlers | 60% | 🔄 In Progress |

## 🐛 Troubleshooting

### Common Issues

**Login Issues:**
- Verify MongoDB connection
- Check admin user creation
- Ensure password hashing compatibility

**Test Failures:**
- Run `go mod tidy` to sync dependencies
- Verify MongoDB test instance is running
- Check template availability for middleware tests

**Database Connection:**
- Verify MongoDB URI in environment variables
- Check network connectivity
- Ensure proper database permissions

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Support

For support and questions:
- **Issues**: [GitHub Issues](https://github.com/your-org/peopleflow/issues)
- **Documentation**: See `/docs` directory
- **Email**: support@peopleflow.com

---

<div align="center">
  <img src="frontend/static/images/PeopleFlow-Logoschrift.svg" alt="PeopleFlow" width="150"/>
  
  **Built with ❤️ using Go, MongoDB, and modern web technologies**
</div>