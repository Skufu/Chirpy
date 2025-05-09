# Chirpy API Usage Examples

This document provides `curl` examples for interacting with the Chirpy server API. It's recommended to use a tool like `jq` to pretty-print JSON responses.

## Prerequisites

1.  **Chirpy Server Running:** Ensure your Chirpy server is running, typically at `http://localhost:8080`.
2.  **`jq` (Optional):** For formatted JSON output. Install it if you haven't already.
3.  **Environment Variables (for script execution):** If you plan to script these, define them:
    ```bash
    BASE_URL="http://localhost:8080"
    # For storing tokens, use actual variables in your script
    ACCESS_TOKEN=""
    REFRESH_TOKEN=""
    USER_EMAIL="test@example.com"
    USER_PASSWORD="password123"
    POLKA_API_KEY="your_polka_api_key_here" # Match your .env POLKA_KEY
    ```

## Users

### 1. Create User

Creates a new user account.

```bash
curl -X POST $BASE_URL/api/users \
-H "Content-Type: application/json" \
-d '{
  "email": "'"$USER_EMAIL"'",
  "password": "'"$USER_PASSWORD"'"
}' | jq
```

**Expected Response (Success - 201 Created):**
```json
{
  "id": "<user_id>",
  "email": "test@example.com",
  "is_chirpy_red": false
}
```

### 2. Login User

Logs in an existing user and returns JWT access and refresh tokens.

```bash
curl -X POST $BASE_URL/api/login \
-u "$USER_EMAIL:$USER_PASSWORD" | jq
```

**Expected Response (Success - 200 OK):**
```json
{
  "id": "<user_id>",
  "email": "test@example.com",
  "is_chirpy_red": false,
  "token": "<access_token_string>",
  "refresh_token": "<refresh_token_string>"
}
```

*Store the `token` (access token) and `refresh_token` for subsequent requests.*

### 3. Update User

Updates the authenticated user's email and/or password. Requires a valid Access Token.

*(First, obtain ACCESS_TOKEN from login)*
```bash
ACCESS_TOKEN="your_access_token_here"
NEW_EMAIL="updated_test@example.com"
NEW_PASSWORD="newpassword456"

curl -X PUT $BASE_URL/api/users \
-H "Authorization: Bearer $ACCESS_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "email": "'"$NEW_EMAIL"'",
  "password": "'"$NEW_PASSWORD"'"
}' | jq
```

**Expected Response (Success - 200 OK):**
```json
{
  "id": "<user_id>",
  "email": "updated_test@example.com",
  "is_chirpy_red": false
}
```

## Chirps

*(Ensure you have a valid ACCESS_TOKEN from login before creating/deleting chirps)*

### 1. Create Chirp

Creates a new chirp for the authenticated user.

```bash
ACCESS_TOKEN="your_access_token_here"
CHIRP_BODY="Hello, Chirpy world! This is my first chirp."

curl -X POST $BASE_URL/api/chirps \
-H "Authorization: Bearer $ACCESS_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "body": "'"$CHIRP_BODY"'"
}' | jq
```

**Expected Response (Success - 201 Created):**
```json
{
  "id": "<chirp_id>",
  "body": "Hello, Chirpy world! This is my first chirp.",
  "user_id": "<user_id_of_author>"
  // Note: The actual API might return created_at, updated_at as well
}
```

### 2. List Chirps

Retrieves a list of all chirps. No authentication required.

```bash
curl -X GET $BASE_URL/api/chirps | jq
```

**Expected Response (Success - 200 OK):**
```json
[
  {
    "id": "<chirp_id_1>",
    "body": "This is chirp one.",
    "user_id": "<user_id_1>"
  },
  {
    "id": "<chirp_id_2>",
    "body": "Another chirp here!",
    "user_id": "<user_id_2>"
  }
  // ... and so on
]
```

### 3. Get Specific Chirp

Retrieves a single chirp by its ID. No authentication required.

```bash
CHIRP_ID_TO_GET="your_chirp_id_here" # Replace with an actual chirp ID

curl -X GET $BASE_URL/api/chirps/$CHIRP_ID_TO_GET | jq
```

**Expected Response (Success - 200 OK):**
```json
{
  "id": "<chirp_id_to_get>",
  "body": "Content of the specific chirp.",
  "user_id": "<user_id_of_author>"
}
```

### 4. Delete Chirp

Deletes a specific chirp. Requires authentication, and the authenticated user must be the author of the chirp.

```bash
ACCESS_TOKEN="your_access_token_here_of_chirp_author"
CHIRP_ID_TO_DELETE="your_chirp_id_to_delete_here"

curl -X DELETE $BASE_URL/api/chirps/$CHIRP_ID_TO_DELETE \
-H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response (Success - 204 No Content):**
*(No JSON body, just HTTP status 204)*

## Tokens

*(Ensure you have a valid REFRESH_TOKEN from login)*

### 1. Refresh Access Token

Issues a new Access Token using a valid Refresh Token.

```bash
REFRESH_TOKEN="your_refresh_token_here"

curl -X POST $BASE_URL/api/refresh \
-H "Authorization: Bearer $REFRESH_TOKEN" | jq
```

**Expected Response (Success - 200 OK):**
```json
{
  "token": "<new_access_token_string>"
}
```
*Store this new `token` (access token).* 

### 2. Revoke Refresh Token

Invalidates a Refresh Token.

```bash
REFRESH_TOKEN_TO_REVOKE="your_refresh_token_to_revoke_here"

curl -X POST $BASE_URL/api/revoke \
-H "Authorization: Bearer $REFRESH_TOKEN_TO_REVOKE"
```

**Expected Response (Success - 204 No Content):**
*(No JSON body, just HTTP status 204)*

## Polka Webhook (Mock)

Example of sending a request to the mock Polka webhook endpoint.

```bash
# Ensure POLKA_API_KEY matches the one in your .env file
POLKA_API_KEY="your_polka_api_key_here"
USER_ID_TO_UPGRADE="user_id_for_chirpy_red"

curl -X POST $BASE_URL/api/polka/webhooks \
-H "Authorization: ApiKey $POLKA_API_KEY" \
-H "Content-Type: application/json" \
-d '{
  "event": "user.upgraded",
  "data": {
    "user_id": "'"$USER_ID_TO_UPGRADE"'"
  }
}'
```

**Expected Response (Success - 204 No Content or 200 OK depending on implementation):**
*(If it only processes and doesn't return data, 204 is common. If it returns a confirmation, it might be 200 OK with a body.)*

## Admin Endpoints

### 1. Get Metrics

Retrieves server metrics (hit count for static files).

```bash
curl -X GET $BASE_URL/admin/metrics
```

**Expected Response (Success - 200 OK):**
```html
<html>
    <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited <strong>0</strong> times!</p>
    </body>
</html>
```
*(The count will vary)*

### 2. Reset Server Data (Development Only)

Resets metrics and clears the database. Only works if `PLATFORM=dev` is set in `.env`.

```bash
curl -X POST $BASE_URL/admin/reset
```

**Expected Response (Success - 204 No Content):**
*(No JSON body, just HTTP status 204)* 