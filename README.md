# Go REST API

This project is a RESTful API built with Go (Golang) using the Gin framework. It provides endpoints to manage plans and their associated cost shares and services. The data is stored in Redis.

## Features

- Create, retrieve, and delete plans
- Validate JSON payloads
- Generate and validate ETags for conditional requests
- Custom validation for specific fields

## Prerequisites

- Go 1.16 or later
- Redis server

## Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/go-rest-api.git
    cd go-rest-api
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Run the Redis server (if not already running):

    ```sh
    redis-server
    ```

## Usage

1. Run the application:

    ```sh
    go run main.go
    ```

2. The API will be available at `http://localhost:8080`.

## Endpoints

### Create a Plan

- **URL**: `/api/v1/plans`
- **Method**: `POST`
- **Request Body**:

    ```json
    {
        "objectId": "12345",
        "objectType": "plan",
        "_org": "example.com",
        "planType": "inNetwork",
        "planCostShares": {
            "deductible": 1000,
            "copay": 20,
            "_org": "example.com",
            "objectId": "67890",
            "objectType": "membercostshare"
        },
        "linkedPlanServices": [
            {
                "linkedService": {
                    "objectId": "11111",
                    "objectType": "service",
                    "_org": "example.com",
                    "name": "Service Name"
                },
                "planserviceCostShares": {
                    "deductible": 500,
                    "copay": 10,
                    "_org": "example.com",
                    "objectId": "22222",
                    "objectType": "membercostshare"
                }
            }
        ]
    }
    ```

- **Response**:

    ```json
    {
        "objectId": "12345",
        "objectType": "plan",
        "_org": "example.com",
        "planType": "inNetwork",
        "creationDate": "2023-10-01",
        "planCostShares": {
            "deductible": 1000,
            "copay": 20,
            "_org": "example.com",
            "objectId": "67890",
            "objectType": "membercostshare"
        },
        "linkedPlanServices": [
            {
                "linkedService": {
                    "objectId": "11111",
                    "objectType": "service",
                    "_org": "example.com",
                    "name": "Service Name"
                },
                "planserviceCostShares": {
                    "deductible": 500,
                    "copay": 10,
                    "_org": "example.com",
                    "objectId": "22222",
                    "objectType": "membercostshare"
                }
            }
        ]
    }
    ```

### Get a Plan

- **URL**: `/api/v1/plans/:id`
- **Method**: `GET`
- **Response**:

    ```json
    {
        "objectId": "12345",
        "objectType": "plan",
        "_org": "example.com",
        "planType": "inNetwork",
        "creationDate": "2023-10-01",
        "planCostShares": {
            "deductible": 1000,
            "copay": 20,
            "_org": "example.com",
            "objectId": "67890",
            "objectType": "membercostshare"
        },
        "linkedPlanServices": [
            {
                "linkedService": {
                    "objectId": "11111",
                    "objectType": "service",
                    "_org": "example.com",
                    "name": "Service Name"
                },
                "planserviceCostShares": {
                    "deductible": 500,
                    "copay": 10,
                    "_org": "example.com",
                    "objectId": "22222",
                    "objectType": "membercostshare"
                }
            }
        ]
    }
    ```

### Delete a Plan

- **URL**: `/api/v1/plans/:id`
- **Method**: `DELETE`
- **Response**: `204 No Content`

