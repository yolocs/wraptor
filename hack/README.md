# Hack

## Build/publish a container quickly (No Docker needed)

1. Install https://ko.build/install/
2. `KO_DOCKER_REPO=[REPLACE_WITH_REPO] ko build ./cmd/repo-template-go --platform=all`

## Deploy to Cloud Run

- Run as a job `gcloud run jobs create JOB_NAME --image IMAGE_URL --command COMMAND --args ARG1,ARG-N`
- Run as a service `gcloud run deploy SERVICE_NAME --image IMAGE_URL --command COMMAND --args ARG1,ARG-N`
