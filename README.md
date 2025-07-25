# Developer setup

## Starting a local database

```
docker run --name code-review-challenge -e POSTGRES_PASSWORD=mysecretpassword --publish 5432:5432 -d postgres
```

## Initial database Migration

-- Brew install assumes you're on a mac, and it should be stated to run flyway migrate from wihthin the database-migrations
-- directory

```
brew install flyway
flyway migrate
```

## Testing the service

The service can be easily tested using bruno (a local REST client), you can open `bruno-tests` from within
bruno and easily query all the available services. You can also run CURL commands as described below in the API section

# The API

## Endpoints

Below is the list of endpoints provided by the API.

--

### 1. **Get All Notes**

**Retrieve a list of all notes.**

- **URL**: `/notes`
- **Method**: `GET`
- **Headers**:

  - `Authorization: Basic <base64-encoded-credentials>`

- **Response Format**: JSON

#### Example Request:

```bash
curl -u your_username:your_password http://localhost:8080/notes
```

#### Example Response:

```json
[
  {
    "id": 1,
    "title": "First Note",
    "content": "This is the first note."
  },
  {
    "id": 2,
    "title": "Second Note",
    "content": "This is the second note."
  }
]
```

### 2. **Create a Note**

**Create a new note in the database.**

- **URL**: `/note` -- Unsure about this naming convention, I need to think of better options.
- **Method**: `POST`
- **Headers**:

  - `Content-Type: application/json`
  - `Authorization: Basic <base64-encoded-credentials>`

- **Request Body**:

  - `title` (string) - The title of the note (required).
  - `content` (string) - The content of the note (optional).

- **Response Format**: JSON

#### Example Request:

```bash
curl -u your_username:your_password -X POST http://localhost:8080/note \
-H "Content-Type: application/json" \
-d '{"title": "My New Note", "content": "This is the content of my new note."}'
```

#### Example Response:

```json
{
  "id": 3,
  "title": "My New Note",
  "content": "This is the content of my new note.",
  "archived": false
}
```

### 3. **Update a Note**

**Update an existing note.**

- **URL**: `/note/{id}` -- show a full example!!!
- **Method**: `PUT`
- **Headers**:

  - `Content-Type: application/json`
  - `Authorization: Basic <base64-encoded-credentials>`

- **Path Parameters**:

  - `id` (integer) - The ID of the note to update.

- **Request Body**: -- This should be consistent with the other examples and show an actual example json

  - `title` (string) - The updated title of the note (optional).
  - `content` (string) - The updated content of the note (optional).
  - `archived` (boolean) - If the note should be archived (optional).

- **Response Format**: JSON

#### Example Request:

```bash
curl -u your_username:your_password -X PUT http://localhost:8080/notes/1 \
-H "Content-Type: application/json" \
-d '{"title": "Updated Note", "content": "This is the updated content."}'
```

#### Example Response:

```json
{
  "id": 1,
  "title": "Updated Note",
  "content": "This is the updated content.",
  "archived": false
}
```

### 5. **Delete a Note** -- What happened to 4?

**Delete a note from the database.**

- **URL**: `/note/{id}` -- discrepancy between this and the example. The code says note is correct
- **Method**: `DELETE`
- **Headers**:

  - `Authorization: Basic <base64-encoded-credentials>`

- **Path Parameters**:
  - `id` (integer) - The ID of the note to delete.

#### Example Request:

```bash
curl -u your_username:your_password -X DELETE http://localhost:8080/notes/1   -- This is the wrong URL
```
