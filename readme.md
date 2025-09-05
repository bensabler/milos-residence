# Milo's Residence

A production-ready, cat-themed bed & breakfast booking system built with Go. Features room reservations, availability checking, and an admin dashboard.

## Architecture

- **Pattern**: Repository pattern with dependency injection
- **Structure**: Clean architecture with separated concerns (handlers, models, repository, rendering)
- **Database**: PostgreSQL with connection pooling and prepared statements
- **Sessions**: Server-side session management with secure cookie handling
- **Security**: CSRF protection, input validation, and SQL injection prevention

## Tech Stack

**Backend**
- Go 1.24.6 with modules
- [Chi](https://github.com/go-chi/chi) - Lightweight HTTP router with middleware support
- [SCS](https://github.com/alexedwards/scs) - Session management
- [Goose](https://github.com/pressly/goose) - Database migrations
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver with connection pooling

**Frontend**
- Bootstrap 5 with responsive design
- Vanilla JavaScript with modern ES6+ features
- Template-driven server-side rendering

**Security & Validation**
- [nosurf](https://github.com/justinas/nosurf) - CSRF protection
- [govalidator](https://github.com/asaskevich/govalidator) - Input validation
- bcrypt password hashing

## Project Structure

```
├── cmd/web/                 # Application entry point
├── internal/
│   ├── config/             # Application configuration
│   ├── driver/             # Database connection management
│   ├── forms/              # Form validation and error handling
│   ├── handlers/           # HTTP request handlers
│   ├── helpers/            # Utility functions
│   ├── models/             # Data models and structures
│   ├── render/             # Template rendering engine
│   └── repository/         # Data access layer
├── migrations/             # Database schema migrations
├── templates/              # HTML templates
└── static/                # Static assets (CSS, JS, images)
```

## Key Features

**Core Functionality**
- Real-time room availability checking
- Reservation management with conflict detection
- Email notifications with template system
- Administrative dashboard with calendar view
- User authentication and session management

**Technical Highlights**
- Template caching with development/production toggle
- Database connection pooling with health checks
- Comprehensive input validation and sanitization
- Middleware-based request processing pipeline
- Graceful error handling and user feedback

## API Endpoints

**Public Routes**
```
GET  /                           # Homepage
GET  /about                      # About page  
GET  /search-availability        # Availability search form
POST /search-availability        # Process availability search
POST /search-availability-json   # JSON API for availability
GET  /make-reservation           # Reservation form
POST /make-reservation           # Process reservation
```

**Admin Routes** (Authentication Required)
```
GET  /admin/dashboard                    # Admin overview
GET  /admin/reservations-all            # All reservations
GET  /admin/reservations-new            # New reservations  
GET  /admin/reservations-calendar       # Calendar view
POST /admin/reservations-calendar       # Update room blocks
```

## Getting Started

### Prerequisites

- Go 1.24.6+
- PostgreSQL 12+
- Make (recommended)

### Setup

1. **Environment Configuration**:
   ```bash
   cp .env.example .env
   # Configure database credentials and settings
   ```

2. **Database Setup**:
   ```bash
   # Apply migrations
   make up
   
   # Seed initial data (optional)
   make seed
   ```

3. **Development**:
   ```bash
   # Install dependencies
   go mod tidy
   
   # Run with hot reload
   make dev
   
   # Or build and run
   make run
   ```

## Testing

- **Unit Tests**: Comprehensive handler and form validation testing
- **Integration Tests**: Database repository testing with test doubles
- **Coverage**: `go test -cover ./...`

```bash
make test     # Run all tests
make vet      # Static analysis
make fmt      # Code formatting
```

## Security Features

- **CSRF Protection**: Token-based request validation
- **Session Security**: HttpOnly cookies with SameSite protection
- **Input Validation**: Server-side validation with user-friendly error messages
- **SQL Injection Prevention**: Parameterized queries and prepared statements
- **Authentication**: Secure password hashing with bcrypt

## Performance Optimizations

- **Database**: Connection pooling with configurable limits
- **Templates**: Production template caching with development bypass
- **Static Assets**: Efficient serving with proper cache headers
- **Middleware**: Optimized request processing pipeline

## Deployment

**Production Build**:
```bash
make build-linux    # Linux binary
make build-windows  # Windows binary  
make build-macos    # macOS binary
```

**Environment Variables**:
- `APP_ENV=prod` - Enables production optimizations
- `USE_TEMPLATE_CACHE=true` - Template caching
- `DB_*` - Database configuration

## Development Tools

```bash
make clean    # Remove build artifacts
make tidy     # Clean up dependencies
make help     # Show all available commands
```

## License

All rights reserved.