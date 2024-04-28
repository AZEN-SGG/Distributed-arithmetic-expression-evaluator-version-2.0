# Distributed Arithmetic Expression Evaluator Version 2.0

## Overview

The Distributed Arithmetic Expression Evaluator Version 2.0 is an enhanced system for asynchronous arithmetic expression computations, supporting multi-user operations. This system integrates user registration and authentication, manages expressions, calculates results, and preserves them in a database.

## Getting Started

### Prerequisites

- Go 1.22 or higher
- SQLite database for storing user data and expressions

### Installation and Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/AZEN-SGG/Distributed-arithmetic-expression-evaluator-version-2.0.git
   ```
2. Navigate to the project directory:
   ```bash
   cd distributed-arithmetic-expression-evaluator-v2.0
   ```
3. Start the server:
   ```bash
   go run main.go
   ```

## System Components

### Server

The server is the core component that handles:
- User registration and authentication
- Addition, processing, and storage of arithmetic expressions
- Management of expression statuses and results
- Execution of arithmetic operations with specified timing

### Clients

This module allows users to interact with the system, supporting registration, authentication, and requests for expression evaluation.

## HTTP Interfaces

### User Registration
**POST** `/register`
- Accepts parameters `username` and `password`.
- Registers a new user in the system.

### User Authentication
**GET** `/login`
- Accepts parameters `username` and `password`.
- Returns a JWT token for accessing protected routes.

### Adding an Arithmetic Expression
**POST** `/expression`
- Accepts parameters `expression`, `id`, and `username`.
- Adds an arithmetic expression to the database and initiates its calculation.

### Retrieving the Result of an Expression
**POST** `/get`
- Accepts parameters `id` and `username`.
- Returns the result of the computed expression, if available.

### List All Expressions for a User
**GET** `/list`
- Accepts parameter `username`.
- Returns a list of all expressions belonging to the user along with their statuses.

### Managing Operation Execution Time
**GET/POST** `/math`
- GET returns the current execution times of operations.
- POST allows updating the execution times of operations (parameters `addition`, `subtraction`, `multiplication`, `division`).

### Viewing and Managing Computing Processes
**GET** `/processes`
- Returns information about current computing processes.

## Usage Examples

### Register a New User
```bash
curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d "{\"username\":\"user1\", \"password\":\"pass123\"}"
```

### Authenticate a User
```bash
curl -X GET http://localhost:8080/login -H "Content-Type: application/json" -d "{\"username\":\"user1\", \"password\":\"pass123\"}"
```

### Add an Expression
```bash
curl -X POST http://localhost:8080/expression -H "Content-Type: application/json" -d "{\"username\":\"user1\", \"id\":\"user_id_123\", \"content\":\"2 + 2\", \"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidXNlcjEiLCJwYXNzd29yZCI6InBhc3MxMjMifQ.l_K3jRZhOYg8l8zEgWJPUlTnEaiNiyBm13ExDACtZxk\"}"
```

### List Expressions for a User
```bash
curl -X GET http://localhost:8080/list -H "Content-Type: application/json" -d "{\"username\":\"user1\", \"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidXNlcjEiLCJwYXNzd29yZCI6InBhc3MxMjMifQ.l_K3jRZhOYg8l8zEgWJPUlTnEaiNiyBm13ExDACtZxk\"}"
```

### Retrieve the Result of an Expression
```bash
curl -X POST http://localhost:8080/get -H "Content-Type: application/json" -d "{\"username\":\"user1\", \"id\":\"user_id_123\", \"token\":\"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidXNlcjEiLCJwYXNzd29yZCI6InBhc3MxMjMifQ.l_K3jRZhOYg

8l8zEgWJPUlTnEaiNiyBm13ExDACtZxk\"}"
```

## Server Interaction

The server provides a REST API for interactions. Requests can be made via any HTTP client, such as `curl` or API testing tools like Postman.

## Scalability and Reliability

The system supports scaling by adding more computing resources. All data is securely stored in a database, allowing the system to resume operations without data loss after failures.

## Security and Protection

The system ensures data and operation security through:
- **Password Encryption**: User passwords are stored in an encrypted format.
- **JWT Authentication**: Access to operations requiring authorization is controlled via JWT tokens, ensuring each request is authenticated.
- **Access Restrictions**: Users can only interact with their expressions, preventing access to others' data.

## Monitoring and Management

### Processes
The system monitors active and pending expressions, offering real-time status updates to users. It provides information on current computing resources, helping optimize resource load and operation times.

## Backup and Recovery

Regular data backups are performed, ensuring business continuity and data protection:
- **Database Backups**: Regular archiving of data allows system recovery in the event of failures.
- **Recovery after Failures**: In case of system crashes or data loss, the orchestrator can quickly restore operational state using the latest backups.

## Conclusion

The Distributed Arithmetic Expression Evaluator Version 2.0 is a robust, scalable, and secure system for asynchronous arithmetic expression calculations, offering users a flexible tool for handling large data volumes in a multi-user environment.