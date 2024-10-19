aws dynamodb get-item --table-name dev-table --key '{"Id": {"S": "USER_INFO-prefix:u"}, "Range": {"S": "user3"}}' --endpoint-url http://localhost:8000
