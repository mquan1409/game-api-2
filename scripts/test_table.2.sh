aws dynamodb get-item --table-name dev-table --key '{"Id": {"S": "GameStat.user1"}, "Range": {"S": "soccer"}}' --endpoint-url http://localhost:8000
