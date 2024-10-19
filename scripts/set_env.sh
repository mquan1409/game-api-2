export BASE_WORK_DIR="$HOME/game-api-2"
table=dev-table
endpoint=http://localhost:8000
region=us-east-1
go_test() {
    DYNAMODB_ENDPOINT="$endpoint" DYNAMODB_TABLE="$table" DYNAMODB_REGION="$region" go test "$@"
}
