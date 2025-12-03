package parsers

import (
	"testing"

	"github.com/petrarca/tech-stack-analyzer/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJavaParser(t *testing.T) {
	parser := NewJavaParser()
	assert.NotNil(t, parser, "Should create a new JavaParser")
	assert.IsType(t, &JavaParser{}, parser, "Should return correct type")
}

func TestParsePomXML(t *testing.T) {
	parser := NewJavaParser()

	tests := []struct {
		name         string
		content      string
		expectedDeps []types.Dependency
	}{
		{
			name: "valid pom.xml with dependencies",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>2.7.0</version>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-jpa</artifactId>
			<version>2.7.0</version>
		</dependency>
		<dependency>
			<groupId>junit</groupId>
			<artifactId>junit</artifactId>
		</dependency>
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{
				{Type: "maven", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "maven", Name: "org.springframework.boot:spring-boot-starter-data-jpa", Example: "2.7.0"},
				{Type: "maven", Name: "junit:junit", Example: "latest"},
			},
		},
		{
			name: "pom.xml with no dependencies",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
</project>`,
			expectedDeps: []types.Dependency{},
		},
		{
			name: "pom.xml with empty dependencies section",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<dependencies>
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{},
		},
		{
			name: "pom.xml with missing groupId or artifactId",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<!-- Missing artifactId -->
			<version>2.7.0</version>
		</dependency>
		<dependency>
			<!-- Missing groupId -->
			<artifactId>spring-boot-starter-data-jpa</artifactId>
			<version>2.7.0</version>
		</dependency>
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{}, // Should skip incomplete dependencies
		},
		{
			name: "invalid XML",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>2.7.0</version>
		</dependency>
	<!-- Missing closing dependency tag -->
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{
				{Type: "maven", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
			}, // XML parser is more lenient than expected
		},
		{
			name:         "empty content",
			content:      "",
			expectedDeps: []types.Dependency{},
		},
		{
			name: "pom.xml with properties and variable substitution",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<properties>
		<spring.version>2.7.0</spring.version>
		<quinoa.version>1.2.3</quinoa.version>
		<junit.version>5.8.2</junit.version>
	</properties>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>${spring.version}</version>
		</dependency>
		<dependency>
			<groupId>io.quarkiverse.quinoa</groupId>
			<artifactId>quarkus-quinoa</artifactId>
			<version>${quinoa.version}</version>
		</dependency>
		<dependency>
			<groupId>org.junit.jupiter</groupId>
			<artifactId>junit-jupiter</artifactId>
			<version>${junit.version}</version>
		</dependency>
		<dependency>
			<groupId>org.mockito</groupId>
			<artifactId>mockito-core</artifactId>
			<version>4.6.1</version>
		</dependency>
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{
				{Type: "maven", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "maven", Name: "io.quarkiverse.quinoa:quarkus-quinoa", Example: "1.2.3"},
				{Type: "maven", Name: "org.junit.jupiter:junit-jupiter", Example: "5.8.2"},
				{Type: "maven", Name: "org.mockito:mockito-core", Example: "4.6.1"},
			},
		},
		{
			name: "pom.xml with undefined property reference",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<properties>
		<spring.version>2.7.0</spring.version>
	</properties>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>${spring.version}</version>
		</dependency>
		<dependency>
			<groupId>io.quarkiverse.quinoa</groupId>
			<artifactId>quarkus-quinoa</artifactId>
			<version>${undefined.version}</version>
		</dependency>
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{
				{Type: "maven", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "maven", Name: "io.quarkiverse.quinoa:quarkus-quinoa", Example: "${undefined.version}"},
			},
		},
		{
			name: "pom.xml with empty properties section",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<properties>
	</properties>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>${spring.version}</version>
		</dependency>
	</dependencies>
</project>`,
			expectedDeps: []types.Dependency{
				{Type: "maven", Name: "org.springframework.boot:spring-boot-starter-web", Example: "${spring.version}"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParsePomXML(tt.content)

			require.Len(t, result, len(tt.expectedDeps), "Should return correct number of dependencies")

			for i, expectedDep := range tt.expectedDeps {
				assert.Equal(t, expectedDep.Type, result[i].Type, "Should have correct type")
				assert.Equal(t, expectedDep.Name, result[i].Name, "Should have correct name")
				assert.Equal(t, expectedDep.Example, result[i].Example, "Should have correct version")
			}
		})
	}
}

func TestParseGradle(t *testing.T) {
	parser := NewJavaParser()

	tests := []struct {
		name         string
		content      string
		expectedDeps []types.Dependency
	}{
		{
			name: "standard Gradle dependencies",
			content: `plugins {
	id 'java'
	id 'org.springframework.boot' version '2.7.0'
}

dependencies {
	implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
	implementation 'org.springframework.boot:spring-boot-starter-data-jpa:2.7.0'
	compile 'junit:junit:4.13.2'
	testImplementation 'org.mockito:mockito-core:4.6.1'
	api 'com.google.guava:guava:31.1-jre'
	compileOnly 'org.projectlombok:lombok:1.18.24'
	runtimeOnly 'mysql:mysql-connector-java:8.0.29'
	testRuntimeOnly 'org.junit.jupiter:junit-jupiter-engine:5.8.2'
}`,
			expectedDeps: []types.Dependency{
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-data-jpa", Example: "2.7.0"},
				{Type: "gradle", Name: "junit:junit", Example: "4.13.2"},
				{Type: "gradle", Name: "org.mockito:mockito-core", Example: "4.6.1"},
				{Type: "gradle", Name: "com.google.guava:guava", Example: "31.1-jre"},
				{Type: "gradle", Name: "org.projectlombok:lombok", Example: "1.18.24"},
				{Type: "gradle", Name: "mysql:mysql-connector-java", Example: "8.0.29"},
				{Type: "gradle", Name: "org.junit.jupiter:junit-jupiter-engine", Example: "5.8.2"},
			},
		},
		{
			name: "Gradle with parentheses notation",
			content: `dependencies {
	implementation("org.springframework.boot:spring-boot-starter-web:2.7.0")
	compile("junit:junit:4.13.2")
	testImplementation("org.mockito:mockito-core:4.6.1")
}`,
			expectedDeps: []types.Dependency{
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "gradle", Name: "junit:junit", Example: "4.13.2"},
				{Type: "gradle", Name: "org.mockito:mockito-core", Example: "4.6.1"},
			},
		},
		{
			name: "Gradle dependencies without versions",
			content: `dependencies {
	implementation 'org.springframework.boot:spring-boot-starter-web'
	compile 'junit:junit'
	testImplementation 'org.mockito:mockito-core'
}`,
			expectedDeps: []types.Dependency{
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-web", Example: "latest"},
				{Type: "gradle", Name: "junit:junit", Example: "latest"},
				{Type: "gradle", Name: "org.mockito:mockito-core", Example: "latest"},
			},
		},
		{
			name: "Gradle with comments and empty lines",
			content: `// Spring Boot dependencies
dependencies {
	implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
	
	/* Test dependencies */
	testImplementation 'org.mockito:mockito-core:4.6.1'
	// JUnit for testing
	compile 'junit:junit:4.13.2'
	
	* Another comment
	api 'com.google.guava:guava:31.1-jre'
}`,
			expectedDeps: []types.Dependency{
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "gradle", Name: "org.mockito:mockito-core", Example: "4.6.1"},
				{Type: "gradle", Name: "junit:junit", Example: "4.13.2"},
				{Type: "gradle", Name: "com.google.guava:guava", Example: "31.1-jre"},
			},
		},
		{
			name: "Gradle with no dependencies",
			content: `plugins {
	id 'java'
}

repositories {
	mavenCentral()
}`,
			expectedDeps: []types.Dependency{},
		},
		{
			name:         "empty Gradle file",
			content:      "",
			expectedDeps: []types.Dependency{},
		},
		{
			name: "Gradle with invalid dependency format",
			content: `dependencies {
	implementation 'invalid-dependency-format'
	compile 'another-invalid'
}`,
			expectedDeps: []types.Dependency{}, // Should not match invalid format
		},
		{
			name: "Kotlin DSL (build.gradle.kts)",
			content: `plugins {
	java
	id("org.springframework.boot") version "2.7.0"
}

dependencies {
	implementation("org.springframework.boot:spring-boot-starter-web:2.7.0")
	testImplementation("org.junit.jupiter:junit-jupiter:5.8.2")
}`,
			expectedDeps: []types.Dependency{
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "gradle", Name: "org.junit.jupiter:junit-jupiter", Example: "5.8.2"},
			},
		},
		{
			name: "Gradle with annotationProcessor",
			content: `dependencies {
	implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
	annotationProcessor 'org.projectlombok:lombok:1.18.24'
	testAnnotationProcessor 'org.projectlombok:lombok:1.18.24'
}`,
			expectedDeps: []types.Dependency{
				{Type: "gradle", Name: "org.springframework.boot:spring-boot-starter-web", Example: "2.7.0"},
				{Type: "gradle", Name: "org.projectlombok:lombok", Example: "1.18.24"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseGradle(tt.content)

			require.Len(t, result, len(tt.expectedDeps), "Should return correct number of dependencies")

			for i, expectedDep := range tt.expectedDeps {
				assert.Equal(t, expectedDep.Type, result[i].Type, "Should have correct type")
				assert.Equal(t, expectedDep.Name, result[i].Name, "Should have correct name")
				assert.Equal(t, expectedDep.Example, result[i].Example, "Should have correct version")
			}
		})
	}
}

func TestJavaParser_Integration(t *testing.T) {
	parser := NewJavaParser()

	// Test Maven integration
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>my-app</artifactId>
	<version>1.0.0</version>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>2.7.0</version>
		</dependency>
	</dependencies>
</project>`

	mavenDeps := parser.ParsePomXML(pomContent)
	assert.Len(t, mavenDeps, 1, "Should parse 1 Maven dependency")
	assert.Equal(t, "maven", mavenDeps[0].Type, "Maven dependency should have correct type")
	assert.Equal(t, "org.springframework.boot:spring-boot-starter-web", mavenDeps[0].Name)
	assert.Equal(t, "2.7.0", mavenDeps[0].Example)

	// Test Gradle integration
	gradleContent := `dependencies {
	implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
	testImplementation 'junit:junit:4.13.2'
}`

	gradleDeps := parser.ParseGradle(gradleContent)
	assert.Len(t, gradleDeps, 2, "Should parse 2 Gradle dependencies")

	// Verify first dependency
	assert.Equal(t, "gradle", gradleDeps[0].Type, "Gradle dependency should have correct type")
	assert.Equal(t, "org.springframework.boot:spring-boot-starter-web", gradleDeps[0].Name)
	assert.Equal(t, "2.7.0", gradleDeps[0].Example)

	// Verify second dependency
	assert.Equal(t, "gradle", gradleDeps[1].Type, "Gradle dependency should have correct type")
	assert.Equal(t, "junit:junit", gradleDeps[1].Name)
	assert.Equal(t, "4.13.2", gradleDeps[1].Example)
}

func TestJavaParser_ComplexScenarios(t *testing.T) {
	parser := NewJavaParser()

	// Test complex Maven with multiple dependency scopes
	complexPom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.example</groupId>
	<artifactId>complex-app</artifactId>
	<version>1.0.0</version>
	
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
			<version>2.7.0</version>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-jpa</artifactId>
			<version>2.7.0</version>
		</dependency>
		<dependency>
			<groupId>org.postgresql</groupId>
			<artifactId>postgresql</artifactId>
			<version>42.3.3</version>
		</dependency>
		<dependency>
			<groupId>org.projectlombok</groupId>
			<artifactId>lombok</artifactId>
		</dependency>
	</dependencies>
</project>`

	mavenDeps := parser.ParsePomXML(complexPom)
	assert.Len(t, mavenDeps, 4, "Should parse 4 Maven dependencies")

	// Create dependency map for verification
	depMap := make(map[string]types.Dependency)
	for _, dep := range mavenDeps {
		depMap[dep.Name] = dep
	}

	assert.Equal(t, "maven", depMap["org.springframework.boot:spring-boot-starter-web"].Type)
	assert.Equal(t, "2.7.0", depMap["org.springframework.boot:spring-boot-starter-web"].Example)
	assert.Equal(t, "latest", depMap["org.projectlombok:lombok"].Example) // No version specified

	// Test complex Gradle with various configurations
	complexGradle := `plugins {
	id 'java'
	id 'org.springframework.boot' version '2.7.0'
}

dependencies {
	// Spring Boot starters
	implementation('org.springframework.boot:spring-boot-starter-web:2.7.0')
	implementation('org.springframework.boot:spring-boot-starter-data-jpa:2.7.0')
	
	// Database
	runtimeOnly 'org.postgresql:postgresql:42.3.3'
	
	// Testing
	testImplementation 'org.springframework.boot:spring-boot-starter-test:2.7.0'
	testImplementation 'org.mockito:mockito-core:4.6.1'
	
	// Utilities
	compileOnly 'org.projectlombok:lombok:1.18.24'
	annotationProcessor 'org.projectlombok:lombok:1.18.24'
	
	// API dependencies
	api 'com.google.guava:guava:31.1-jre'
}`

	gradleDeps := parser.ParseGradle(complexGradle)
	assert.Len(t, gradleDeps, 8, "Should parse 8 Gradle dependencies including annotationProcessor")

	// Verify specific dependencies
	gradleDepMap := make(map[string]types.Dependency)
	for _, dep := range gradleDeps {
		gradleDepMap[dep.Name] = dep
	}

	assert.Equal(t, "gradle", gradleDepMap["org.postgresql:postgresql"].Type)
	assert.Equal(t, "42.3.3", gradleDepMap["org.postgresql:postgresql"].Example)
	assert.Equal(t, "gradle", gradleDepMap["org.projectlombok:lombok"].Type)
	assert.Equal(t, "1.18.24", gradleDepMap["org.projectlombok:lombok"].Example)
}
