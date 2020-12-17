# Basic Test with an Associated Database

## Overview
This example creates a service which is backed by a relational database. No actual tables/rows are used, but an SQL
command is used to retrieve the date from the database.

The purpose of including the database is to demonstrate "chained dependencies" which must be resolved. Specifically:
 - The web application requires a database (see `rezolvr-catalog.yaml`)
 - The database component requires environment properties (see `rezolvr-db.yaml`)

Rezolvr ensures that all dependencies are resolved before deployment files are generated.

## Clone the Application Repo

Steps:
1. Clone the catalog application `git clone https://github.com/tcrowleyibm/catalog.git`
2. Navigate into the catalog subdirectory
3. Locally build the sample container: `docker build -t catalog .`

## Docker Compose:

Note: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Steps:
1. Ensure the container image exists by running the following command: `docker images`. You should see an image for "welcome".
2. Run Rezolvr to resolve all components:
`rezolvr apply -a ./rezolvr-db.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml`

Or, to see each component one at a time ...
```
rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev-docker.yaml -s ./state.yaml
rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml
```

3. Navigate to the `out` directory, and run the following command: `docker-compose up`
4. Launch a browser and navigate to http://localhost:3001/dbtime
5. The response should contain the current time. Reloading the page will update the time.
6. Close the docker window and remove the old image. (`docker ps -a`, `docker rm <all containers>`)

## Kubernetes (Minikube)

Note 1: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Note 2: This deployment will create two application pods and a single database pod. See the "uses" section
of `env-dev-kube.yaml` for additional details about how this works.

Steps:
1. Ensure the Kubernetes cluster (Minikube) is running. If not, then execute: `minikube start`.
2. Point to the cluster's registry and manually push the image
```
eval $(minikube docker-env)
# Nav to source directory
# Build the image
docker build -t catalog .
# Verify the image is there
docker images
```
3. From the `basicdb` directory, run Rezolvr to resolve all components:
`rezolvr apply -a ./rezolvr-db.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml`

Or, to see each component one at a time ...
```
rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev-kube.yaml -s ./state.yaml
rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml
```
After running, there will be two files in the `out` directory: `mydb.yaml` and `catalogapp.yaml`.
4. Navigate to the output directory, and run the following commands:
```
kubectl apply -f mydb.yaml
kubectl apply -f catalogapp.yaml
```
5. Get the IP address of the deployed service: `minikube service --url catalogapp-service`.
6. Launch a browser and navigate to: http://<ip addr from step 5>:<port from step 5>/dbtime
7. You should see a message similar to the following: `{"message":"2020-12-06T20:12:11.519Z"}`
8. To remove the deployed workload: `kubectl delete -f catalogapp.yaml`.
