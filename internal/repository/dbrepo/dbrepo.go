package dbrepo

import (
	"database/sql"

	"github.com/bensabler/milos-residence/internal/config"
	"github.com/bensabler/milos-residence/internal/repository"
)

// postgresDBRepo implements the Repository pattern with PostgreSQL-specific optimizations.
// This struct provides the concrete implementation of database operations specifically
// tailored for PostgreSQL's capabilities, features, and performance characteristics.
// It demonstrates how the Repository pattern enables database-specific optimizations
// while maintaining the same interface contract across different database implementations.
//
// The struct composition approach (embedding dependencies rather than inheriting behavior)
// follows Go's design philosophy of favoring composition over inheritance. This enables
// the repository to access both application-wide configuration and database connections
// while maintaining clear separation of concerns and explicit dependency management.
//
// Design Pattern: Repository Implementation - concrete data access layer for PostgreSQL
// Design Pattern: Adapter - adapts PostgreSQL-specific operations to repository interface
// Design Pattern: Composition - combines configuration and database access dependencies
type postgresDBRepo struct {
	// App provides access to application-wide configuration and shared services
	// This includes loggers for operational monitoring, session management for user state,
	// and configuration flags that might affect database operation behavior (like debug modes).
	// Having access to app configuration enables sophisticated error logging, performance
	// monitoring, and integration with application lifecycle management.
	App *config.AppConfig

	// DB provides access to the PostgreSQL connection pool for executing queries
	// This *sql.DB instance represents the managed connection pool configured with
	// appropriate timeouts, connection limits, and performance settings for production use.
	// The repository uses this connection pool for all database operations while
	// relying on the pool's built-in management for connection lifecycle and health.
	DB *sql.DB
}

// testDBRepo implements the Test Double pattern for repository testing and development.
// This struct provides a lightweight alternative to the full PostgreSQL repository
// that enables fast, isolated testing without requiring database infrastructure.
// It demonstrates how the Repository pattern facilitates comprehensive testing strategies
// by enabling different implementations for different contexts (production vs testing).
//
// The Test Double pattern is essential for maintaining fast, reliable test suites because:
// 1. Tests run independently without requiring external database setup or cleanup
// 2. Test execution is deterministic and not affected by database state or network issues
// 3. Different test scenarios can use different mock behaviors without complex database manipulation
// 4. Continuous integration environments don't need full database infrastructure
// 5. Developers can run comprehensive tests quickly during development without database dependencies
//
// Design Pattern: Test Double - provides testing alternative to production implementation
// Design Pattern: Mock Object - enables controlled testing scenarios with predictable behavior
// Design Pattern: Null Object - provides safe default behaviors for testing contexts
type testDBRepo struct {
	// App provides access to application configuration for testing scenarios
	// Even in testing contexts, repositories may need access to configuration for
	// behaviors like error logging, feature flags, or test-specific settings that
	// control how mock operations behave or what data they return for specific scenarios.
	App *config.AppConfig

	// DB field exists for interface consistency but typically remains nil in test implementations
	// Some testing scenarios might use lightweight database alternatives (like SQLite in-memory)
	// while others rely entirely on mocked data and behaviors without any database interaction.
	// This flexibility enables different testing strategies based on specific testing goals.
	DB *sql.DB
}

// NewPostgresRepo implements the Factory Method pattern for creating PostgreSQL repository instances.
// This factory function handles the complete initialization process for production repository
// instances, ensuring that all dependencies are properly injected and the repository is
// ready for immediate use in production environments. It demonstrates how factory functions
// can encapsulate complex initialization logic while providing simple, clean interfaces.
//
// The Factory Method pattern is particularly valuable for repository creation because:
// 1. Repository initialization involves multiple dependencies that must be coordinated properly
// 2. Different environments (development, staging, production) might require different configurations
// 3. Error handling during initialization can be complex and should be centralized
// 4. The factory can perform validation to ensure repositories are created in valid states
// 5. Factory methods provide a stable interface that can evolve without breaking calling code
//
// This approach enables dependency injection at the application level while maintaining
// encapsulation of implementation details within the repository layer. The factory method
// serves as the integration point between the application's dependency management system
// and the repository's internal structure and requirements.
//
// Design Pattern: Factory Method - creates properly initialized repository instances
// Design Pattern: Dependency Injection - injects external dependencies into repository
// Design Pattern: Builder - assembles complex object with multiple dependencies
// Parameters:
//
//	conn: Database connection pool configured for production use with appropriate settings
//	a: Application configuration providing access to loggers, session management, and settings
//
// Returns: Fully configured repository instance ready for production data operations
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	// Create and return PostgreSQL repository instance with all dependencies properly injected
	// This construction ensures that the repository has immediate access to both database
	// operations (through the connection pool) and application services (through the config)
	// without requiring additional setup or configuration after creation.
	//
	// The explicit return type (repository.DatabaseRepo interface) enforces the contract
	// that this factory produces instances conforming to the repository interface,
	// enabling compile-time verification of interface compliance and supporting the
	// dependency inversion principle throughout the application architecture.
	return &postgresDBRepo{
		App: a,    // Application configuration for logging, session management, and settings
		DB:  conn, // Production database connection pool for executing queries and transactions
	}
}

// NewTestingRepo implements the Factory Method pattern for creating test repository instances.
// This factory function creates lightweight repository instances specifically designed for
// testing scenarios, enabling fast, isolated tests without requiring database infrastructure
// or complex setup procedures. It demonstrates how the same factory pattern can create
// different implementations based on the intended usage context.
//
// The testing repository factory enables several important testing strategies:
// 1. **Unit Testing**: Business logic can be tested in isolation from database concerns
// 2. **Integration Testing**: Different repository behaviors can be simulated without database manipulation
// 3. **Performance Testing**: Tests can focus on application logic without database performance variability
// 4. **Error Scenario Testing**: Various database error conditions can be simulated reliably
// 5. **Continuous Integration**: Test suites run quickly without external database dependencies
//
// This factory typically creates repositories with predictable, controllable behavior that
// enables comprehensive testing of business logic while avoiding the complexity, performance
// overhead, and reliability issues associated with testing against real databases.
//
// Design Pattern: Factory Method - creates test-appropriate repository instances
// Design Pattern: Test Double Factory - produces testing alternatives to production implementations
// Design Pattern: Null Object - provides safe default behaviors for testing scenarios
// Parameters:
//
//	a: Application configuration that may contain test-specific settings or behaviors
//
// Returns: Test repository instance with predictable behavior suitable for automated testing
func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	// Create and return test repository instance configured for testing scenarios
	// The test implementation typically provides predictable responses, controlled error
	// conditions, and fast execution without external dependencies, enabling comprehensive
	// testing of business logic that depends on repository operations.
	//
	// Note that the database connection is not provided to test repositories because
	// they typically don't interact with real databases. Instead, they use in-memory
	// data structures, predefined responses, or other testing techniques that provide
	// the speed, isolation, and predictability required for effective automated testing.
	return &testDBRepo{
		App: a, // Application configuration that may include test-specific settings
		// DB is intentionally not set, as test repositories typically don't use real database connections
	}
}
