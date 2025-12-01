package scanner

import "github.com/petrarca/tech-stack-analyzer/internal/types"

// ComponentTypes defines which technology types should create components vs just be listed as dependencies
// This classification determines whether a detected technology appears in the 'tech' field (primary technologies)
// or only in the 'techs' array (all technologies including tools/libraries)

// ShouldCreateComponent determines if a rule should create a component
// Returns true if component should be created, false otherwise
func ShouldCreateComponent(rule types.Rule) bool {
	// If is_component is explicitly set, use that value
	if rule.IsComponent != nil {
		return *rule.IsComponent
	}

	// Otherwise, use type-based logic (existing behavior)
	return !IsNotAComponent(rule.Type)
}

// IsNotAComponent returns true if the given type should NOT create a separate component
// These types represent tools, libraries, or utilities rather than architectural components
func IsNotAComponent(techType string) bool {
	// Types that should NOT create components:
	// - Development tools and utilities (ci, builder, linter, test, validation)
	// - Language and runtime environments (language, runtime)
	// - Code organization tools (framework, orm, package_manager)
	// - UI utilities (ui, ui_framework, iconset)
	// - Infrastructure as Code (iac)
	notAComponent := map[string]bool{
		"ci":              true, // CI/CD tools (GitHub Actions, Jenkins, etc.)
		"language":        true, // Programming languages (Python, JavaScript, etc.)
		"runtime":         true, // Runtime environments (Node.js runtime, Python runtime, etc.)
		"tool":            true, // General development tools
		"framework":       true, // Application frameworks (React, Django, Spring, etc.)
		"validation":      true, // Validation libraries
		"builder":         true, // Build tools (Webpack, Vite, etc.)
		"linter":          true, // Code linters (ESLint, Pylint, etc.)
		"test":            true, // Testing frameworks (Jest, Pytest, etc.)
		"orm":             true, // Object-relational mappers (SQLAlchemy, Prisma, etc.)
		"package_manager": true, // Package managers (npm, pip, cargo, etc.)
		"ui":              true, // UI component libraries
		"ui_framework":    true, // UI frameworks
		"iac":             true, // Infrastructure as Code tools (Terraform, Pulumi, etc.)
		"iconset":         true, // Icon libraries (Lucide, Font Awesome, etc.)
	}

	return notAComponent[techType]
}

// GetComponentTypes returns a list of all technology types that DO create components
// These represent architectural decisions and infrastructure choices
func GetComponentTypes() []string {
	return []string{
		"db",            // Databases (PostgreSQL, MongoDB, Redis, etc.)
		"hosting",       // Hosting platforms (Vercel, Netlify, AWS, etc.)
		"cloud",         // Cloud providers (AWS, GCP, Azure, etc.)
		"saas",          // SaaS platforms
		"cms",           // Content management systems
		"monitoring",    // Monitoring services (Datadog, Sentry, etc.)
		"communication", // Communication tools (Slack, Discord, etc.)
		"analytics",     // Analytics platforms (Google Analytics, Mixpanel, etc.)
		"etl",           // ETL tools
		"app",           // Applications (Nginx, Redis, etc.)
		"auth",          // Authentication services (Auth0, Clerk, etc.)
		"payment",       // Payment processors (Stripe, PayPal, etc.)
		"storage",       // Storage services (S3, Cloudinary, etc.)
		"notification",  // Notification services (SendGrid, Twilio, etc.)
		"queue",         // Message queues (RabbitMQ, Kafka, etc.)
		"automation",    // Automation tools
		"security",      // Security services
		"maps",          // Map services (Google Maps, Mapbox, etc.)
		"crm",           // CRM systems
		"network",       // Network services
		"collaboration", // Collaboration platforms
		"ssg",           // Static site generators (when used as main tech)
		"ai",            // AI services (OpenAI, Anthropic, etc.) - architectural decisions
	}
}
