# Managed (External) Resources

## Overview
This sample is very similar to the `basicdb` sample, except the relational database is externally managed;
it's hosted by a public cloud provider. Because it's external, output files
should not be generated.

Within both `env-dev-docker.yaml` and `env-dev-kube.yaml`, there exists a section for platform settings.
It can be found in the "uses" section:

```
uses:
  - type: platform.settings
    name: mydb
    params:
      - name: isExternal
        value: true
```

If the value is `true`, then output files will not be generated. This can be verified by reviewing the output
of the rezolvr command. When running this sample, look for a line in the output similar to the following:

`Based on the platform settings, a template will not be generated for: mydb. (isExternal=true)`


## Clone the Application Repo

Steps:
1. Clone the catalog application `git clone https://github.com/tcrowleyibm/catalog.git`
2. Navigate into the catalog subdirectory
3. Locally build the sample container: `docker build -t catalog .`

## Create a Managed Database

Using a public cloud provider, create a managed instance of a Postgres database. This can be any public cloud provider.

1. Access a public cloud provider and create an instance of Postgres.
2. Take note of the database credentials, and update the relevant properties in the files named `env-dev-docker.yaml` and `env-dev-kube.yaml`.
2. Seed the database with some data. (See the `registry` sample for more details.)


## Docker Compose:

Note: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Steps:
1. Ensure the container image exists by running the following command: `docker images`. You should see an image for "welcome".
2. Run Rezolvr to resolve all components:
    `rezolvr apply -a ./rezolvr-externaldb.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml`

      Or, to see each component one at a time ...
      ```
      rezolvr apply -a ./rezolvr-db.yaml -e ./env-dev-docker.yaml -s ./state.yaml
      rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-docker.yaml -s ./state.yaml
      ```

3. Navigate to the `out` directory, and run the following command: `docker-compose up`
4. Launch a browser and navigate to http://localhost:3001/charters
5. The response should contain the three rows that were loaded into the database.
6. Close the docker window and remove the old image. (`docker ps -a`, `docker rm <all containers>`)

## Kubernetes (Minikube)

Note 1: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Note 2: This deployment will create two application pods. See the "uses" section
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
3. From the `managed` directory, run Rezolvr to resolve all components:
      `rezolvr apply -a ./rezolvr-externaldb.yaml -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml`

      Or, to see each component one at a time ...
      
      ```
      rezolvr apply -a ./rezolvr-externaldb.yaml -e ./env-dev-kube.yaml -s ./state.yaml
      rezolvr apply -a ./rezolvr-catalog.yaml -e ./env-dev-kube.yaml -s ./state.yaml
      ```

      After running, there will be a file in the `out` directory named `catalogapp.yaml`.

4. Navigate to the output directory, and run the following commands:
    ```
    kubectl apply -f catalogapp.yaml
    ```
5. Get the IP address of the deployed service: `minikube service --url catalogapp-service`.
6. Launch a browser and navigate to: http://<ip addr from step 5>:<port from step 5>/charters
7. The response should contain the three rows that were loaded into the database.
8. To remove the deployed workload: `kubectl delete -f catalogapp.yaml`.
