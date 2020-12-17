# Volume One

## Overview
This example includes both a front end and a database. In addition, the images are not
stored locally, but located on a "remote" registry server. Also, the database makes use of a persistent volume.

This sample builds upon prior samples, so it may be a good idea to review them first.

## Registry Configuration and Adding Data

See the "registry" readme for additional information about loading data and configuring an unsecured registry.

## Docker Compose:

Note: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Steps:
1. If not previously completed, clone and build the sample image: `docker build -t catalog .`
2. Push the relevant images to the registry:
    ```
    docker pull postgres
    docker tag postgres host.minikube.internal:5000/postgres
    docker push host.minikube.internal:5000/postgres
    docker tag catalog host.minikube.internal:5000/catalog
    docker push host.minikube.internal:5000/catalog
    ```
3. Verify that the images have been pushed: `curl -X GET http://host.minikube.internal:5000/v2/_catalog`
4. Run Rezolvr to resolve all components:
`rezolvr apply -a rezolvr-volume.yaml -a ./rezolvr-registry.yaml -a ./rezolvr-db.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml`

... or ...

```
rezolvr apply -a ./rezolvr-registry.yaml -e ./env-dev.yaml-docker -s ./state.yaml
rezolvr apply -a ./rezolvr-volume.yaml -e ./env-dev.yaml-docker -s ./state.yaml
rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev.yaml-docker -s ./state.yaml
rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml
```

5. Navigate to the output directory, and run the following command: `docker-compose up`
6. Ensure data is loaded into the database.
7. Launch a browser and navigate to http://localhost:3001/charters. You should see three entries.
8. Stop and remove the containers. Once they're completely gone, re-run `docker-compose up`. 
9. Launch a browser and navigate to http://localhost:3001/charters. You should still see three entries - even though you didn't load any data. That's because the data persisted between containers.
10. Close the docker window and remove the old image. (`docker ps -a`, `docker rm <all containers>`)

## Kubernetes (Minikube)

Note: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Steps:
1. Ensure the Kubernetes cluster (Minikube) is running, and make sure that an unsecured registry is configured. (See the registry sample for details.)
2. Locally build the sample image: `docker build -t catalog .`
3. Push the relevant images to the registry:
    ```
    docker pull postgres
    docker tag postgres host.minikube.internal:5000/postgres
    docker push host.minikube.internal:5000/postgres
    docker tag catalog host.minikube.internal:5000/catalog
    docker push host.minikube.internal:5000/catalog
    ```
4. Verify that the images have been pushed: `curl -X GET http://host.minikube.internal:5000/v2/_catalog`
5. Run Rezolvr to resolve all components:
`rezolvr apply -a ./rezolvr-volume.yaml -a ./rezolvr-registry.yaml -a ./rezolvr-db.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml`

... or ...

```
rezolvr apply -a ./rezolvr-registry.yaml -e ./env-dev.yaml-kube -s ./state.yaml
rezolvr apply -a ./rezolvr-volume.yaml -e ./env-dev.yaml-kube -s ./state.yaml
rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev.yaml-kube -s ./state.yaml
rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml
```

6. Navigate to the output directory, and run the following commands:
```
kubectl apply -f dbvolume.yaml
kubectl apply -f dbvolumeclaim.yaml
kubectl apply -f mydb.yaml
kubectl apply -f catalogapp.yaml
```
7. Load the data into the Postgres database. (It may be necessary to first expose the postgres service to do this.
   See the `registry` sample for details.)
8. Get the IP address of the deployed service: `minikube service --url catalogapp-service`.
9. Launch a browser and navigate to: http://<ip addr from step 8>:<port from step 8>/charters. You should see three entries.
10. Delete the mydb.yaml and catalogapp.yaml deployments:
    ```
    kubectl delete -f catalogapp.yaml
    kubectl delete -f mydb.yaml
    ```
11. Re-load the deployments and reload the browser page. You should still see the three charters - even though the deployment / pods were deleted.
    ```
    kubectl apply -f mydb.yaml
    kubectl apply -f catalogapp.yaml
    ```
10. To remove the deployed workload:
    ```
    kubectl delete -f catalogapp.yaml
    kubectl delete -f mydb.yaml
    ```
