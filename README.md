# game-api

## API Specifications

### User Service

1. `GET /users/{id}`
   - Get user by ID
2. `GET /users?prefix={prefix}`
   - Get users with usernames starting with the given prefix
3. `GET /users/{userId}/games/{gameId}/stats`
   - Get game statistics for a specific user and game
4. `POST /users`
   - Create a new user
   - Input model:
     ```json
     {
       "UserID": "string",
       "Username": "string",
       "Email": "string",
       "GamesPlayed": ["string"]
     }
     ```
5. `PUT /users/{id}`
   - Update an existing user
   - Input model:
     ```json
     {
       "Username": "string",
       "Email": "string",
       "GamesPlayed": ["string"]
     }
     ```
6. `DELETE /users/{id}`
   - Delete a user

### Game Service

1. `GET /games/{id}`
   - Get game by ID
2. `GET /games/{gameId}/leaderboard/{attribute}`
   - Get full leaderboard for a game and attribute
3. `GET /games/{gameId}/leaderboard/{attribute}?limit={limit}`
   - Get bounded leaderboard for a game and attribute
4. `POST /games`
   - Create a new game
   - Input model:
     ```json
     {
       "GameID": "string",
       "Description": "string",
       "Attributes": ["string"],
       "RankedAttributes": ["string"]
     }
     ```
5. `PUT /games/{id}`
   - Update an existing game
   - Input model:
     ```json
     {
       "Description": "string",
       "Attributes": ["string"],
       "RankedAttributes": ["string"]
     }
     ```
6. `DELETE /games/{id}`
   - Delete a game

### Match Service

1. `GET /matches/{gameId}/{matchId}/{dateId}`
   - Get a specific match
2. `GET /matches?game={gameId}&date={dateId}`
   - Get all matches for a game on a specific date
3. `POST /matches`
   - Create a new match
   - Input model:
     ```json
     {
       "MatchID": "string",
       "DateID": "string",
       "GameID": "string",
       "TeamNames": ["string"],
       "TeamScores": [number],
       "TeamMembers": [["string"]],
       "PlayerAttributesMap": {
         "UserID1": {
           "AttributeName1": number,
           "AttributeName2": number
         },
         "UserID2": {
           "AttributeName1": number,
           "AttributeName2": number
         }
       }
     }
     ```
4. `PUT /matches/{id}`
   - Update an existing match
   - Input model:
     ```json
     {
       "DateID": "string",
       "GameID": "string",
       "TeamNames": ["string"],
       "TeamScores": [number],
       "TeamMembers": [["string"]],
       "PlayerAttributesMap": {
         "UserID1": {
           "AttributeName1": number,
           "AttributeName2": number
         },
         "UserID2": {
           "AttributeName1": number,
           "AttributeName2": number
         }
       }
     }
     ```
5. `DELETE /matches/{gameId}/{matchId}/{dateId}`
   - Delete a match

Note: All endpoints return appropriate HTTP status codes and error messages. Authentication and authorization mechanisms are not specified in this API and should be implemented separately.

## Testing

To run the API locally:

1. Run `./scripts/run_dynamodb.sh` to start a local DynamoDB instance.
2. Run `./scripts/create_table.sh` to create the necessary tables.
3. Run `./scripts/seed_table.sh` to seed the tables with some data.
4. Run `source ./scripts/set_env.sh` to set the environment variables.
5. Run `./scripts/sam_build.sh` to build the SAM application.
6. Run `./scripts/sam_run.sh` to run the SAM application.

Notes:
- You can modify the `./scripts/set_env.sh` file to set the environment variables to your desired values based on your AWS IAM user config.