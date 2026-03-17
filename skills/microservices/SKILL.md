# Microservices Deployment Skill

## TRIGGER
Read this file before designing, implementing, or reviewing ANY microservice architecture or component. Use this when the user asks to "build a microservice", "design system architecture", or "deploy a new service".

---
name: microservices-architecture
description: >
  Production-grade 5-layer microservices architecture pattern.
  Focuses on API Gateway, gRPC sync communication, Kafka async event bus with Outbox Pattern,
  Consumer services with Saga rollback, and full Infrastructure observability.
---

# 🏗️ Microservices General Architecture (5 Layers)

## IDENTITY
You are a Senior System Architect specializing in distributed systems and microservices.
You follow the "Database per Service" principle and prioritize eventual consistency with guaranteed delivery.

---

## 📐 THE 5-LAYER ARCHITECTURE

### Layer 1: Client & API Gateway (REST/HTTPS) - The Entry Point
- **Responsibility**: Single entry point for all clients (Web, Mobile, 3rd-party).
- **Communication**: REST/HTTPS only.
- **Core Functions**:
  - **Authentication/Authorization**: Centralized JWT verification.
  - **Rate Limiting**: Protect downstream services from DDoS/abuse.
  - **Request Routing**: Proxying requests to correct internal services.
  - **Load Balancing**: Distributing traffic across service instances.
- **Rule**: No business logic in the Gateway. It is a traffic coordinator.

### Layer 2: Core Services (gRPC, Sync) - The Source of Truth
- **Key Services**: User, Order, Product, Auth.
- **Communication**: Internal sync calls via **gRPC**.
- **Data Persistence**: Each service has its own dedicated database (e.g., PostgreSQL).
- **Architecture Highlights**:
  - **Protobuf**: Strict schema definition for all internal APIs.
  - **Auth Service**: Always called synchronously to verify every request.
  - **Outbox Pattern**: When a state changes (e.g., `order.created`), the service writes to its own DB *and* an `outbox` table in the same transaction to guarantee event delivery to Kafka.

### Layer 3: Kafka Message Bus (Async, Event-Driven)
- **Role**: Decouples Core Services from Consumer Services.
- **Guarantee**: "At least once" delivery ensures events are never lost.
- **Typical Topics**: `order.created`, `user.deleted`, `payment.success`, `stock.reserved`.
- **Infrastructure**: Distributed log for high throughput and scalability.

### Layer 4: Consumer Services (Kafka Subscribe, Async)
- **Key Services**: Payment, Inventory, Notification, Search.
- **Workflow**:
  - Listen to relevant Kafka topics.
  - Process events independently and asynchronously.
- **Error Handling**:
  - **Compensating Events**: If a step fails (e.g., Payment fails after Order created), publish a rollback event (e.g., `order.rollback`) to revert the state in Core Services.
  - **Idempotency**: Consumers must handle duplicate events without side effects.

### Layer 5: Infrastructure & Observability (Cross-Cutting)
- **Service Discovery**: Consul or Kubernetes DNS for service tracking.
- **Monitoring**: Prometheus (metrics collection) + Grafana (visualization).
- **Tracing**: OpenTelemetry/Jaeger for end-to-end request tracking across layers.
- **Deployment**: Dockerized services orchestrated by **Kubernetes**.

---

## 🛠️ MANDATORY ARCHITECTURAL RULES

1. **Strict Layering**: Clients MUST NEVER bypass the API Gateway to talk to Core Services.
2. **Database Isolation**: One service = One Database. Shared databases are FORBIDDEN.
3. **Communication Protocol**:
   - Client → Gateway: **REST/HTTPS**
   - Internal Sync (Immediate Result): **gRPC**
   - Internal Async (State Change): **Kafka (Event Bus)**
4. **Reliability (The Outbox Pattern)**:
   - DO NOT publish to Kafka directly from service code.
   - DO write to an `outbox` table in the same DB transaction as the business logic.
   - Use a separate "Relay" process to move events from `outbox` to Kafka.
5. **Resiliency (Saga Pattern)**:
   - Use compensating transactions to handle distributed failures.
   - If Service B fails after Service A succeeds, Service B must trigger a cleanup in Service A via a "rollback" event.

---

## 🔍 OBSERVABILITY STANDARDS

- **Trace Propagation**: Every request must carry a `TraceID` from Gateway to the deepest Consumer.
- **Health Checks**: Every service must expose `/health` and `/ready` endpoints.
- **Structured Logging**: Use JSON format with `service_name`, `trace_id`, and `level` fields.

---

# 📚 DETAILED REFERENCES & PATTERN SPECIFICATIONS

## 1. The Outbox Pattern (Guaranteed Delivery)
To avoid the "Dual Write" problem (DB succeeds but Kafka fails), follow this flow:
1. **Transaction Start**: Begin DB transaction.
2. **Business Logic**: Update domain tables (e.g., `orders`).
3. **Outbox Insert**: Insert event payload into `outbox` table in the *same* transaction.
4. **Transaction Commit**: Finish DB transaction.
5. **Relay Strategy**: A separate process (e.g., Debezium or a custom poller) reads from `outbox` and publishes to Kafka.
6. **Mark Sent**: Once Kafka acknowledges, the relay marks the outbox entry as `processed`.

## 2. Distributed Transactions (Saga Pattern)
Since we use "Database per Service", we use **Choreography-based Saga**:
- **Happy Path**: Service A publishes `Event.Success` → Service B listens and acts.
- **Failure Path**: Service B fails → publishes `Event.Failed` → Service A listens and runs a **Compensating Transaction** (e.g., `Update Status to CANCELLED`).
- **Rule**: Never use 2PC (Two-Phase Commit) in high-scale microservices.

## 3. Communication Protocols (Deep Dive)
- **gRPC (Layer 2)**: 
  - Use for real-time dependencies (e.g., Order service needs User info *now*).
  - Define `.proto` files in a shared `proto/` repository.
  - Advantage: Bi-directional streaming, heavy optimization via HTTP/2.
- **Kafka (Layer 3 & 4)**:
  - Use for "Fire and Forget" or data synchronization.
  - Implement **Dead Letter Queues (DLQ)** for events that fail to process after 3 retries.
  - Consumer groups must be used for scaling.

## 4. Security Enforcement
- **Gateway**: Validates JWT signature using Public Key (RSA/EdDSA).
- **Service-to-Service**: Use mTLS (Mutual TLS) within the cluster via Service Mesh (Istio/Linkerd) or verify internal tokens.
- **Secret Management**: Never hardcode keys. Use HashiCorp Vault or Kubernetes Secrets.

## 5. Technology Stack Mapping (Reference)
| Component | Recommended Tool | Alternative |
| :--- | :--- | :--- |
| **Gateway** | Kong / KrakenD | Nginx / APISIX |
| **Service Mesh** | Istio / Linkerd | Consul Connect |
| **Database** | PostgreSQL | MySQL / MongoDB |
| **Message Bus** | Kafka | RabbitMQ / Pulsar |
| **Monitoring** | Prometheus / Grafana | Datadog / New Relic |
| **Tracing** | Jaeger / OpenTelemetry | Zipkin |
| **Platform** | Kubernetes (k8s) | Nomad / ECS |
