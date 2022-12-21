# assignment

## To run:
`go run cmd/main.go`

This application by default runs at localhost:8080. It serves 3 APIs. Refer to docs/apispec.yaml for detailed specification.

The directory to upload images and server listening port are configurable with environment variables.
 * Set `LISTEN_ON` for listening port.
 * Set `UPLOAD_PATH` for image upload location.

## Limitation:
Supports one image upload/download per request.


## Example curl commands:
```
curl -X POST http://localhost:8080/api/v1/image -F "file=@/location/file.png" -H "Content-Type: multipart/form-data"

curl -X GET http://localhost:8080/api/v1/images

curl -X GET http://localhost:8080/api/v1/image/d3825644-1dfd-47fa-a75d-f725b01d8276?width=300 --output ~/image.png
```
