package serve

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gizzahub/gzh-manager-go/pkg/api"
	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
	"github.com/spf13/cobra"
)

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start REST API server",
	Long: `Start the GZH Manager REST API server to provide HTTP access to all functionality.

The server provides a comprehensive REST API with endpoints for:
- Bulk repository cloning operations
- Plugin management and execution
- System configuration and monitoring
- Internationalization support
- Health checks and metrics

Examples:
  gz serve                              # Start on default port 8080
  gz serve --port 3000                  # Start on custom port
  gz serve --host 0.0.0.0 --port 8080   # Bind to all interfaces
  gz serve --env production --auth       # Production mode with auth
  gz serve --swagger --debug             # Development mode with Swagger docs`,
	Run: runServe,
}

var (
	host          string
	port          int
	environment   string
	corsOrigins   string
	enableAuth    bool
	enableSwagger bool
	readTimeout   int
	writeTimeout  int
	logLevel      string
	rateLimit     int
)

func init() {
	ServeCmd.Flags().StringVar(&host, "host", "localhost", "Server host")
	ServeCmd.Flags().IntVar(&port, "port", 8080, "Server port")
	ServeCmd.Flags().StringVar(&environment, "env", "development", "Environment (development, staging, production)")
	ServeCmd.Flags().StringVar(&corsOrigins, "cors-origins", "*", "CORS allowed origins")
	ServeCmd.Flags().BoolVar(&enableAuth, "auth", false, "Enable authentication")
	ServeCmd.Flags().BoolVar(&enableSwagger, "swagger", true, "Enable Swagger documentation")
	ServeCmd.Flags().IntVar(&readTimeout, "read-timeout", 30, "Read timeout in seconds")
	ServeCmd.Flags().IntVar(&writeTimeout, "write-timeout", 30, "Write timeout in seconds")
	ServeCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, silent)")
	ServeCmd.Flags().IntVar(&rateLimit, "rate-limit", 100, "Rate limit per minute")
}

func runServe(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 Starting GZH Manager API Server...")

	// Create server configuration
	config := &api.Config{
		Host:          host,
		Port:          port,
		Environment:   environment,
		CORSOrigins:   corsOrigins,
		EnableAuth:    enableAuth,
		EnableSwagger: enableSwagger,
		ReadTimeout:   readTimeout,
		WriteTimeout:  writeTimeout,
		LogLevel:      logLevel,
		RateLimit:     rateLimit,
	}

	// Check for environment variables
	if envHost := os.Getenv("GZH_SERVER_HOST"); envHost != "" {
		config.Host = envHost
	}
	if envPort := os.Getenv("GZH_SERVER_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			config.Port = p
		}
	}
	if envEnv := os.Getenv("GZH_ENVIRONMENT"); envEnv != "" {
		config.Environment = envEnv
	}
	if envAuth := os.Getenv("GZH_ENABLE_AUTH"); envAuth == "true" {
		config.EnableAuth = true
	}
	if envSwagger := os.Getenv("GZH_ENABLE_SWAGGER"); envSwagger == "false" {
		config.EnableSwagger = false
	}

	// Create GZH client
	clientConfig := gzhclient.DefaultConfig()
	clientConfig.LogLevel = logLevel
	clientConfig.EnablePlugins = true

	client, err := gzhclient.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("❌ Failed to create GZH client: %v", err)
	}
	defer client.Close()

	// Create and start server
	server := api.NewServer(config, client)

	// Print server information
	fmt.Printf("🏠 Server configuration:\n")
	fmt.Printf("  Host: %s\n", config.Host)
	fmt.Printf("  Port: %d\n", config.Port)
	fmt.Printf("  Environment: %s\n", config.Environment)
	fmt.Printf("  Authentication: %v\n", config.EnableAuth)
	fmt.Printf("  Swagger docs: %v\n", config.EnableSwagger)
	fmt.Printf("  Log level: %s\n", config.LogLevel)
	fmt.Printf("  Rate limit: %d req/min\n", config.RateLimit)

	if config.EnableSwagger {
		fmt.Printf("📚 Swagger documentation: http://%s:%d/swagger/\n", config.Host, config.Port)
	}
	fmt.Printf("🚑 Health check: http://%s:%d/health\n", config.Host, config.Port)
	fmt.Printf("🔗 API base URL: http://%s:%d/api/v1\n", config.Host, config.Port)

	fmt.Println("\n📌 Available endpoints:")
	fmt.Println("  POST   /api/v1/bulk-clone          - Execute bulk clone operation")
	fmt.Println("  GET    /api/v1/bulk-clone/status/:id - Get operation status")
	fmt.Println("  GET    /api/v1/plugins             - List available plugins")
	fmt.Println("  POST   /api/v1/plugins/:name/execute - Execute plugin method")
	fmt.Println("  GET    /api/v1/plugins/:name       - Get plugin information")
	fmt.Println("  PUT    /api/v1/plugins/:name/enable - Enable plugin")
	fmt.Println("  PUT    /api/v1/plugins/:name/disable - Disable plugin")
	fmt.Println("  GET    /api/v1/config              - Get current configuration")
	fmt.Println("  PUT    /api/v1/config              - Update configuration")
	fmt.Println("  POST   /api/v1/config/validate     - Validate configuration")
	fmt.Println("  GET    /api/v1/system/info         - Get system information")
	fmt.Println("  GET    /api/v1/system/metrics      - Get system metrics")
	fmt.Println("  GET    /api/v1/system/logs         - Get system logs")
	fmt.Println("  GET    /api/v1/i18n/languages      - Get supported languages")
	fmt.Println("  PUT    /api/v1/i18n/language/:lang - Set current language")
	fmt.Println("  GET    /api/v1/i18n/messages/:lang - Get localized messages")

	// Start server (this blocks until shutdown)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}

	fmt.Println("🚑 Server shutdown complete")
}
