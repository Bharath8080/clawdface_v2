FROM 605982976738.dkr.ecr.us-east-2.amazonaws.com/golang:1.24.7-alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . /app

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /main

EXPOSE 3077

# Run
CMD ["/main"]