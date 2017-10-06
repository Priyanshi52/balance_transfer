# Sample calls to REST API server

## enroll

```bash
curl -s -X POST http://localhost:4000/users -H "content-type: application/x-www-form-urlencoded" -d 'username=Oleg&orgName=org1'
```
save jwt token returned into env variable for future calls
```bash
export JWT=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDcyODMyNjYsInVzZXJuYW1lIjoiT2xlZyIsIm9yZ05hbWUiOiJvcmcxIiwiaWF0IjoxNTA3MjQ3MjY2fQ.Mkmhcu_65dBEotPJdXPRqsNrXKwCY0-j7kvIXNMetF8
```
## create channel
```bash
curl -s -X POST http://localhost:4000/channels \
  -H "authorization: Bearer $JWT" \
  -H "content-type: application/json" \
  -d '{
	"channelName":"mychannel",
	"channelConfigPath":"../../artifacts/channel/mychannel.tx"
}'
```
## join to channel
```bash
curl -s -X POST http://localhost:4000/channels/mychannel/peers \
  -H "authorization: Bearer $JWT" \
  -H "content-type: application/json" \
  -d '{
	"peers": ["org1/peer1"]
}'
```
## install chaincode
```bash
curl -s -X POST http://localhost:4000/chaincodes \
  -H "authorization: Bearer $JWT" \
  -H "content-type: application/json" \
  -d '{
	"peers": ["org1/peer1"],
	"chaincodeName":"mycc",
	"chaincodePath":"github.com/example_cc",
	"chaincodeVersion":"v0"
}'
```
## instantiate chaincode
```bash
curl -s -X POST \
  http://localhost:4000/channels/mychannel/chaincodes \
  -H "authorization: Bearer $JWT" \
  -H "content-type: application/json" \
  -d '{
	"chaincodeName":"mycc",
	"chaincodeVersion":"v0",
	"functionName":"Init",
	"args":["a","100","b","200"]
}'
```
