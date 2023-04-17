# Reverse Proxy for Cloud Run

This project provides a reverse proxy that forwards incoming requests to either public or private Google Cloud Run services.

## Prerequisites

- [Go](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/dbut2/cloud-run-reverse-proxy.git
cd cloud-run-reverse-proxy
```

2. Set the required environment variables:

```bash
export PUBLIC_URL="https://your-public-service.run.app"
export PRIVATE_URL="https://your-private-service.run.app"
export PRIVATE_CLIENT_ID="127.0.0.1"
```

3. Build the Docker image:

```bash
docker build -t reverse-proxy -f reverse-proxy.Dockerfile .
```

4. Run the Docker container:

```bash
docker run -it --rm -p 8080:8080 \
--env PUBLIC_URL=$PUBLIC_URL \
--env PRIVATE_URL=$PRIVATE_URL \
--env PRIVATE_CLIENT_ID=$PRIVATE_CLIENT_ID \
reverse-proxy
```

The reverse proxy is now running on `http://localhost:8080` and will forward incoming requests to the appropriate Cloud Run service based on the client's IP address.

## Deployment

To deploy the reverse proxy on Google Cloud Run, follow these steps:

1. Push the Docker image to Google Container Registry:

```bash
docker tag reverse-proxy gcr.io/YOUR_PROJECT_ID/reverse-proxy
docker push gcr.io/YOUR_PROJECT_ID/reverse-proxy
```

2. Deploy the Cloud Run service:

```bash
gcloud run deploy reverse-proxy \
--image gcr.io/YOUR_PROJECT_ID/reverse-proxy \
--platform managed \
--allow-unauthenticated \
--region YOUR_REGION \
--update-env-vars PUBLIC_URL=$PUBLIC_URL,PRIVATE_URL=$PRIVATE_URL,PRIVATE_CLIENT_ID=$PRIVATE_CLIENT_ID
```

This will deploy the reverse proxy to a new Cloud Run service. The `--allow-unauthenticated` flag is used to allow public access to the reverse proxy. Make sure to replace `YOUR_PROJECT_ID` and `YOUR_REGION` with your Google Cloud project ID and the desired region, respectively.
