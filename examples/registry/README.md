# Registry One

## Overview
This example includes both a front end and a database. In addition, the container images are not
stored locally, but located on a remote registry server. The registry has been added as a component of the overall solution.
(See `rezolvr-registry.yaml` for the specifics.)

This sample builds upon the `basicdb` sample. If you've not yet completed that sample, it's recommended that you complete it now.

Items of note:
 - Both `env-dev-docker.yaml` and `env-dev-kube.yaml` have two sets of environment properties: registryProps and dbEnvProps.
   This is for organizational purposes.
 - `rezolvr-db.yaml` depends on both sets of properties.

## Adding Data

At this point, since no persistent volumes are being used, it's necessary to reload data every time. After the database has been
started, connect to the `catalog` database and execute the following SQL commands:

```
DROP TABLE CHARTERS;
CREATE TABLE CHARTERS (
    charter_id SERIAL PRIMARY KEY,
    charter_name VARCHAR(128) NOT NULL,
    charter_descr VARCHAR(254) NOT NULL
);

INSERT INTO CHARTERS (charter_name, charter_descr) VALUES('Inshore 1', 'Inshore half day fishing with guide.');
INSERT INTO CHARTERS (charter_name, charter_descr) VALUES('Inshore 2', 'Inshore full day fishing with guide. Lunch provided.');
INSERT INTO CHARTERS (charter_name, charter_descr) VALUES('Offshore 1', 'Offshore half day fishing with guide.');
```

This content exists as a file in the `catalog` source code repository. If you have this locally, the following command can load
the data: `psql -p 5432 -h localhost -U abetterusername -d catalog -f init.sql`

For minikube / kubernetes, it may be necessary to expose postgres via a service. If so, the following will help:

 - Run: `kubectl expose deployment postgres --type=NodePort --name=postgres-service`
 - Get the IP address and port from kubernetes: `minikube service --url postgres-service`
 - Connect to the database and create the charters database:
    - `psql -p 31172 -h 192.168.99.101 -U abetterusername -d postgres`
    - `CREATE DATABASE CATALOG;`
    - `\q`
 - Load data from the `init.sql` file: `psql -p 31172 -h 192.168.99.101 -U abetterusername -d catalog -f init.sql`

Data should now be available.

## Registry Madness

Setting up an HTTP (insecure) local registry is painful. Hopefully, this helps:

1. Configure Docker to support an insecure registry
 - Edit (sudo) `/etc/hosts` and create an entry for host.minikube.internal. Set the IP address to be your local IP address. (Something like 192.168.1.155)
 - Update docker to allow an insecure registry for that:
   * Docker Preferences -> Docker Engine. You should see the following:
        ```
        {
          "debug": true,
          "experimental": false,
        }
        ```
    * Update it to this:
        ```
        {
          "debug": true,
          "experimental": false,
          "insecure-registries": [
            "host.minikube.internal:5000"
          ]
        }
        ```
  - Apply and restart
2. Start an instance of a Docker registry locally: `docker run -d -p 5000:5000 --restart=always --name registry registry:2`
3. Tag and push an image to this registry:
    ```
    docker tag catalog host.minikube.internal:5000/catalog
    docker push host.minikube.internal:5000/catalog
    ```
4. Check that it's there: `curl -X GET http://host.minikube.internal:5000/v2/_catalog`. The response should look like this:
    ```
    {"repositories":["catalog"]}
    ```
5. Configure minikube to support an insecure registry
  - COMPLETELY DELETE minikube and start with an insecure registry:
    `minikube start --driver="virtualbox" --insecure-registry="host.minikube.internal:5000"`
6. Test that it works:
    ```
    minikube ssh
    docker images
    # pick an image and tag it. (We will use the welcome sample images)
    docker tag welcome host.minikube.internal:5000/welcome
    # push to registry
    docker push host.minikube.internal:5000/welcome
    ```
If you don't get an error, then you're good to go

You should now be able to both push to the insecure registry from your local machine, and pull from minikube. :-/

Note: When done, consider removing the entry in the `hosts` file and the insecure registry from Docker and minikube.

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
`rezolvr apply -a ./rezolvr-registry.yaml -a ./rezolvr-db.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml`

... or ...

```
rezolvr apply -a ./rezolvr-registry.yaml -e ./env-dev.yaml-docker -s ./state.yaml
rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev.yaml-docker -s ./state.yaml
rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml
```

Note: Running the above three commands out of order will result in an unresolved error. This is because each dependency builds upon
the previous dependency. To avoid this issue include all component files in a single command.

5. Navigate to the output directory, and review the file named `docker-compose.yaml`. Note that the docker images are no longer
   expected to appear in a local registry.
6. Run the following command: `docker-compose up`
7. Launch a browser and navigate to http://localhost:3001/dbtime. You should see the current date and time.
8. Add a table and three rows of data to the database. (See the "adding data" section.)
9. Launch a browser and navigate to http://localhost:3001/charters. You should see three entries.
10. The response should contain the current time. Reloading the page will update the time.
11. Close the docker window and remove the old containers. (`docker ps -a`, `docker rm <all containers>`)


## Kubernetes (Minikube)

Note: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Steps:
1. Ensure the Kubernetes cluster (Minikube) is running, and make sure that an insecure registry is configured. (See above for details.)
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
`rezolvr apply -a ./rezolvr-registry.yaml -a ./rezolvr-db.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml`

... or ...

```
rezolvr apply -a ./rezolvr-registry.yaml -e ./env-dev.yaml-kube -s ./state.yaml
rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev.yaml-kube -s ./state.yaml
rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml
```

6. Navigate to the output directory, and run the following commands:
```
kubectl apply -f mydb.yaml
kubectl apply -f catalogapp.yaml
```
7. Load data into the database. (See above.)
8. Get the IP address of the deployed service: `minikube service --url catalogapp-service`.
9. Launch a browser and navigate to: http://<ip addr from step 7>:<port from step 7>/charters. You should see data about three charters.
10. To remove the deployed workload:
    ```
    kubectl delete -f mydb.yaml
    kubectl delete -f catalogapp.yaml
    ```

## View a Dependency Error

To view an error, ensure `state.yaml` does not exist, and run the following command:

`rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev-docker.yaml -s ./state.yaml`

The resulting error should state that a dependency within db.yaml has not been satisfied. Specifically, db.yaml has a need for a registry, which does not exist.

This error is resolved by loading `./registry.yaml` as described below.

