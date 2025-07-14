# Calculator Service

**Calculator Service** - Part of PizzaMaker Microservices Architecture: A microservice for pizza ingredient calculations based on Domain-Driven Design (DDD) architecture with complete observability and monitoring.

## Main Features

- **Automatic Ingredient Calculation**: Automatically calculates ingredient quantities for pizza dough
- **Multi-Recipe Support**: Handles different pizza recipe types
- **Business Metrics**: Collects domain-specific metrics (accuracy, weight, hydration)

## Technologies

- **Go** - Primary language
- **gRPC** - Inter-service communication
- **Prometheus** - Metrics and monitoring
- **OpenTelemetry + Jaeger** - Distributed tracing
- **Logrus** - Structured logging
- **Docker** - Containerization

## Endpoints

### gRPC Services
- **Port**: 50051
- **Service**: `DoughCalculatorServer`
- **Methods**: 
  - `CalculateDough(DoughRequest) -> DoughResponse`
  - `ValidateIngredients(IngredientsRequest) -> ValidationResponse`

### HTTP Endpoints
- **Port**: 8080
- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check

## Observability

### Structured Logging
- **Correlation ID** for cross-service request tracking
- **Structured JSON** for easy parsing
- **Configurable levels** (Debug, Info, Warn, Error)

### Distributed Tracing
- **OpenTelemetry** for instrumentation
- **Jaeger** for trace visualization
- **Automatic spans** for gRPC operations

### Prometheus Metrics
The service exposes both **business** and **technical** metrics:

#### Business Metrics
- `calculator_calculations_total` - Total number of calculations by type
- `calculator_calculation_duration_seconds` - Calculation duration
- `calculator_dough_accuracy` - Calculated dough accuracy
- `calculator_dough_weight_grams` - Dough weight in grams
- `calculator_dough_hydration_percentage` - Hydration percentage
- `calculator_ingredient_validations_total` - Ingredient validations
- `calculator_recipe_types_total` - Count by recipe type

#### Technical Metrics
- `calculator_grpc_requests_total` - Total gRPC requests
- `calculator_grpc_request_duration_seconds` - gRPC request duration
- `calculator_active_calculations` - Active concurrent calculations
