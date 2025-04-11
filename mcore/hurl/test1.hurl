
GET {{myurl}}/health/
HTTP 200

POST {{myurl}}/add-msg/
content-type: application/json
secret: test
{
  "chatId": 1,
  "userName": "testUser",
  "text": "y https://www.youtube.com/watch?v=mWTeo9sQr6w",
  "messageId": 13
}

HTTP 200
[Asserts]
jsonpath "$.status" ==  "OK"
jsonpath "$.taskMsg.text" ==  "task created"

#### Get ytdl task

# POST {{myurl}}/get-task/
# content-type: application/json
# secret: test
# {
#   "taskType": 2
# }

# HTTP 200
# [Asserts]
# jsonpath "$.status" ==  "OK"
# [Captures]
# taskId_type2 : jsonpath "$.data.id"

# POST {{myurl}}/report-task/
# content-type: application/json
# secret: test

# {
#   "taskId": {{taskId_type2}},
#   "status": 3,
#   "textMsg": "ok"
# }
# HTTP 200
# [Asserts]
# jsonpath "$.status" ==  "OK"

#### Get bot task

# POST {{myurl}}/get-task/
# content-type: application/json
# secret: test
# {
#   "taskType": 1
# }

# HTTP 200
# [Asserts]
# jsonpath "$.status" ==  "OK"
# [Captures]
# taskId_type1 : jsonpath "$.data.id"

# POST {{myurl}}/report-task/
# content-type: application/json
# secret: test

# {
#   "taskId": {{taskId_type1}},
#   "status": 3,
#   "textMsg": "ok"
# }
# HTTP 200
# [Asserts]
# jsonpath "$.status" ==  "OK"