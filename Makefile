# Makefile implements the Build Automation pattern for streamlined development and deployment workflows.
# This file demonstrates how professional Go projects organize build tasks, database management,
# and development utilities into convenient, repeatable commands that ensure consistent behavior
# across different development environments and team members. It serves as the central automation
# hub that simplifies complex development tasks into simple, memorable commands.
#
# Build automation provides several critical benefits for development teams:
# 1. **Consistency**: All developers use identical build and deployment procedures
# 2. **Efficiency**: Complex multi-step processes reduced to single commands
# 3. **Error Prevention**: Automated processes eliminate manual mistakes and configuration drift
# 4. **Documentation**: Makefile serves as executable documentation of project procedures
# 5. **Onboarding**: New team members can quickly learn and use standard development workflows
#
# The Makefile integrates multiple tools and processes including Go compilation, database migrations,
# environment configuration, and cross-platform builds into a cohesive development experience
# that supports both local development and production deployment scenarios.

# ---- Load local environment configuration ----
# This section implements the Environment Configuration pattern for flexible build behavior.
# The "-include .env" directive attempts to load local environment variables from a .env file
# if it exists, providing a way for developers to customize build behavior without modifying
# the shared Makefile. The leading dash (-) makes the include optional, preventing errors
# when the .env file doesn't exist, which is common in fresh project setups or CI environments.
#
# Environment loading enables:
# - Local database credentials without committing secrets to version control
# - Custom build flags and optimization settings per developer
# - Different configuration for development versus production builds
# - Override of default values with project-specific or environment-specific settings
-include .env
export

# ---- Application build configuration ----
# These variables implement the Configuration Management pattern for build customization.
# Using Make variables with default values (the ?= operator) allows these settings to be
# overridden through environment variables or command-line parameters, providing flexibility
# while maintaining sensible defaults for common development scenarios.

# APP defines the output binary name, defaulting to the project directory name
# This creates a clear relationship between the project name and the generated executable,
# making it easy to identify binaries and supporting consistent naming conventions
APP       ?= milos-residence

# MAIN specifies the main package location for Go build commands
# Go applications typically organize main packages under cmd/ subdirectories,
# with cmd/web being a common pattern for web applications that distinguishes
# web servers from other possible executables like CLI tools or background services
MAIN      ?= ./cmd/web

# LDFLAGS configures linker flags for optimized production builds
# The "-s -w" flags strip debugging information and symbol tables from binaries,
# significantly reducing executable size for production deployments while maintaining
# full functionality. This optimization is particularly important for containerized
# applications where image size affects deployment speed and storage costs
LDFLAGS   ?= -s -w 

# GOFLAGS provides additional Go build flags for advanced build customization
# This variable allows developers to pass additional flags to Go build commands,
# such as build tags for conditional compilation, race detection flags for testing,
# or module proxy configuration for dependency management in restricted environments
GOFLAGS   ?=

# ---- Database configuration ----
# This section implements the Database Configuration pattern for flexible data persistence.
# Database settings use environment variables with sensible defaults, enabling the same
# Makefile to work across development, testing, and production environments with different
# database configurations while maintaining security through external configuration management.

# DB specifies the database driver type for migration and connection operations
# PostgreSQL is chosen as the default because it provides excellent features for web applications
# including ACID transactions, concurrent access, advanced indexing, and JSON support
DB        ?= postgres

# Database connection parameters with secure defaults for local development
# These settings provide a working database configuration for immediate development use
# while supporting override through environment variables for different environments
DB_HOST   ?= localhost
DB_PORT   ?= 5432
DB_USER   ?= app
DB_NAME   ?= appdb
DB_SSLMODE?= disable

# MIG defines the migration directory location for database schema management
# Organizing migrations in a dedicated directory supports version control of schema changes
# and enables collaborative development where multiple developers can contribute schema
# modifications without conflicts through sequential migration numbering
MIG       ?= ./migrations

# Compose DSN (Data Source Name) from individual configuration parameters
# This demonstrates the Builder pattern for constructing complex configuration strings
# from component parts, enabling flexible configuration while maintaining compatibility
# with PostgreSQL connection string requirements. The DSN composition handles optional
# parameters and supports various deployment scenarios from local development to production
DSN ?= host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)

# ---- Database migration helpers ----
# GOOSE implements the Migration Tool Integration pattern for database schema management.
# This variable demonstrates how Makefiles can integrate external tools (goose) with
# project-specific configuration, providing a bridge between the migration tool's
# capabilities and the project's configuration management approach.
#
# The GOOSE variable sets up environment variables that the goose migration tool reads:
# - GOOSE_DRIVER: Specifies which database driver to use (postgres, mysql, etc.)
# - GOOSE_DBSTRING: Provides the complete database connection string
# - GOOSE_MIGRATION_DIR: Tells goose where to find migration files
#
# This configuration approach enables the migration tool to work with project-specific
# settings while maintaining the tool's standard command interface for developers
GOOSE = GOOSE_DRIVER=$(DB) GOOSE_DBSTRING="$(DSN)" GOOSE_MIGRATION_DIR=$(MIG) goose

# ---- Phony target declarations ----
# This section implements the Build Target Documentation pattern for Make target management.
# The .PHONY declaration tells Make that these targets don't create files with matching names,
# preventing conflicts with potential files and ensuring targets always execute when requested.
# This is crucial for targets like "clean" or "test" that perform actions rather than create files.
#
# Phony targets improve build reliability by:
# - Preventing target skipping when files with matching names exist
# - Clearly documenting which targets are actions versus file creation
# - Supporting consistent target execution regardless of directory contents
# - Enabling build system optimization through clear target classification
.PHONY: help up up1 down down1 redo reset goto to version status create fix run build br dev clean fmt vet tidy test build-linux build-windows build-macos

# ---- Application build targets ----
# These targets implement the Build Pipeline pattern for Go application compilation.
# Build targets provide different compilation strategies for various development and
# deployment scenarios, from quick development builds to optimized production binaries
# for multiple platforms and architectures.

# build creates optimized production binary with configured flags and settings
# This target demonstrates the Production Build pattern with linker optimization flags
# that reduce binary size and remove debugging information for deployment environments.
# The build target provides the foundation for containerized deployments and production releases
build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP) $(MAIN)

# run combines build and execution for convenient development workflows
# This target implements the Build-and-Run pattern that streamlines the development cycle
# by automatically rebuilding the application before execution, ensuring that developers
# always run the latest code changes without manual build management
run: build 
	DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) ./$(APP)

# br provides abbreviated alias for run target following developer convenience patterns
# Short aliases improve developer productivity by reducing typing for frequently used commands
# while maintaining the full command names for clarity in documentation and automation scripts
br: run

# dev implements the Development Mode pattern for rapid iteration during coding
# This target uses "go run" for immediate execution without intermediate binary creation,
# providing faster feedback during development at the cost of slightly slower execution.
# Development mode prioritizes quick iteration over execution performance
dev:
	go run $(MAIN)

# clean implements the Cleanup pattern for build artifact management
# This target removes generated files to ensure clean builds and prevent issues
# with stale binaries or cross-platform build artifacts. Clean builds are essential
# for troubleshooting build issues and ensuring consistent deployment artifacts
clean:
	rm -f $(APP) $(APP).exe $(APP)-linux $(APP)-darwin

# ---- Code quality and maintenance targets ----
# These targets implement the Code Quality Automation pattern for maintaining code standards.
# Automated code quality checks ensure consistent formatting, catch potential bugs, and
# maintain dependency hygiene across the development team without manual oversight.

# fmt applies Go's standard code formatting automatically across the entire project
# Consistent code formatting improves readability, reduces diff noise in version control,
# and eliminates formatting-related discussions during code review processes
fmt: ; go fmt ./...

# vet runs Go's built-in static analysis tool to detect potential bugs and suspicious code
# Static analysis catches common programming errors like unreachable code, incorrect
# struct tags, and misuse of standard library functions before runtime testing
vet: ; go vet ./...

# tidy manages Go module dependencies and removes unused dependencies
# Module tidying keeps dependency management clean and ensures reproducible builds
# by maintaining accurate dependency tracking and removing unused imports
tidy: ; go mod tidy

# test executes all tests in the project with Go's standard testing framework
# Automated testing validates application functionality and prevents regressions
# during development while providing confidence for deployment and refactoring
test: ; go test ./...

# ---- Cross-platform build targets ----
# These targets implement the Cross-Platform Build pattern for multi-architecture deployment.
# Cross-platform builds enable deployment to different operating systems and architectures
# from a single development environment, supporting diverse deployment scenarios including
# cloud containers, edge devices, and multi-platform distribution.

# build-linux creates optimized binary for Linux x86_64 systems
# Linux builds support containerized deployments, cloud servers, and traditional server infrastructure
# The amd64 architecture covers most server and desktop Linux installations
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP)-linux $(MAIN)

# build-windows creates optimized binary for Windows x86_64 systems  
# Windows builds support desktop development environments and Windows Server deployments
# The .exe extension follows Windows executable naming conventions
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP).exe $(MAIN)

# build-macos creates optimized binary for macOS x86_64 systems
# macOS builds support developer workstations and macOS server deployments
# The darwin target name reflects macOS's underlying Darwin operating system
build-macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP)-darwin $(MAIN)

# help provides self-documenting interface for available build targets
# Self-documenting build systems improve developer onboarding and reduce the need
# for separate documentation that might become outdated as the build system evolves
help:
	@echo "Targets:"
	@echo "  up|up1|down|redo|reset|status|version|goto v=NNN|create n=name [t=sql|go]"

# ---- Database migration targets ----
# These targets implement the Database Migration Management pattern for schema evolution.
# Migration management enables collaborative development with database schema changes
# while maintaining data integrity and supporting rollback scenarios for deployment safety.
# The goose migration tool provides robust migration capabilities with version tracking.

# up applies all pending migrations to bring database schema to latest version
# This target implements the Forward Migration pattern for deploying schema changes
# incrementally while maintaining data integrity and supporting concurrent development
up:        ; $(GOOSE) up

# up1 applies exactly one pending migration for careful schema change deployment
# Single-step migrations enable cautious deployment of schema changes with immediate
# rollback capability if issues are discovered during or after migration application
up1:       ; $(GOOSE) up-by-one

# down rolls back the most recent migration for error recovery scenarios
# Migration rollback provides safety net for schema changes that cause problems
# in production, enabling rapid recovery while maintaining data consistency
down:      ; $(GOOSE) down

# down1 provides alias for down target following consistent naming patterns
# Command aliases improve developer experience by providing multiple ways to invoke
# the same functionality while maintaining clear primary command documentation
down1:     ; $(GOOSE) down

# redo rolls back and reapplies the most recent migration for development iteration
# Migration redo supports rapid development iteration when refining schema changes
# without creating additional migration versions during development cycles
redo:      ; $(GOOSE) redo

# reset rolls back all migrations to empty database state for clean development
# Database reset enables clean slate development and testing scenarios where
# starting from empty database state is needed for comprehensive testing
reset:     ; $(GOOSE) reset

# status displays current migration state for database version awareness
# Migration status provides visibility into database schema version and pending
# changes, supporting deployment planning and troubleshooting schema issues
status:    ; $(GOOSE) status

# version displays current database schema version for deployment tracking
# Version information supports deployment verification and helps coordinate
# application deployments with corresponding database schema requirements
version:   ; $(GOOSE) version

# goto migrates database to specific version for precise schema management
# Targeted migration enables precise schema version control for testing specific
# application versions, reproducing production issues, or coordinating deployments.
# Usage: make goto v=20250823170000
goto:
ifndef v
	$(error Provide version with v=NNN)
endif
	$(GOOSE) goto $(v)

# to provides convenient alias for goto target with consistent naming patterns
# Short aliases reduce typing for frequently used commands while maintaining
# clear documentation and supporting both abbreviated and full command forms
to: goto

# create generates new migration files for schema change development
# Migration creation provides standardized templates and naming conventions
# for schema changes while supporting both SQL and Go migration formats.
# Usage: make create n=add_users_table [t=sql]  (or t=go)
create:
ifndef n
	$(error Provide name with n=your_migration_name)
endif
	$(GOOSE) create $(n) $(if $(t),$(t),sql)

# ---- Development and debugging helpers ----
# These targets implement the Development Support pattern for troubleshooting and configuration.
# Development helpers provide visibility into build configuration and system state,
# supporting efficient troubleshooting and onboarding of new team members.

# fix displays current configuration without exposing sensitive information
# Configuration debugging helps troubleshoot build and connection issues while
# maintaining security by avoiding exposure of passwords or other sensitive data
# in build output or log files that might be shared for troubleshooting
fix:
	@echo DB=$(DB)
	@echo DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE)
	@echo MIG=$(MIG)
	@which goose || true