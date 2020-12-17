# An Introductory Test

## Overview
This example uses a very simple web app to display a message. Since it's an introductory example,
there are limited dependencies. However, there are additional examples which illustrate some of
the more advanced features of rezolvr.

The key rezolvr file for this example is named `rezolvr-welcome.yaml`. That file defines a single component. Within the component, one resource is needed, one resource is used, and one is provided. 

Some items of note:
 - It's considered a best practice to have a single component in a file (line 1)
 - The only needed resources are environment properties. 
 - A parameter can use a go-style formula to create values at runtime. (Lines 14 & 19)

## Clone the Application Repo

Steps:
1. Clone the welcome starter application `git clone https://github.com/tcrowleyibm/welcome.git`
2. Navigate into the welcome subdirectory
3. Locally build the sample container: `docker build -t welcome .`

## Docker Compose:

Note: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.

Steps:
1. From the `intro` directory, execute: `rezolvr apply -a ./rezolvr-welcome.yaml -e ./env-dev-docker.yaml -s ./state.yaml`. This will generate a `docker-compose.yaml` file in a subdirectory named `out`.

      **NOTE:** Rezolvr doesn't automatically create the output directory. This is intentional. Before running the examples, you should
      create an output directory (e.g. `mkdir out`)


2. Navigate to the `out` directory, and run the following command: `docker-compose up`
3. Launch a browser and navigate to http://localhost:3000/message
4. You should see the following text: Hello from Rezolvr!
5. Close the browser window and remove the old image. (`docker ps -a`, `docker rm <containername>`)


## Kubernetes (Minikube)

Note 1: For a completely clean run, be sure to remove existing state by deleting `state.yaml` from the current directory.
Note 2: These instructions assume that VirtualBox is being used within minikube. If that's not the case, then the process
for obtaining the IP address (step 5) may differ.

Steps:
1. Ensure the Kubernetes cluster (Minikube) is running. If not, then execute: `minikube start`.
2. Point to the cluster's registry and manually push the image
    ```
    eval $(minikube docker-env)
    # Nav to source directory (e.g. cd welcome)
    # Build the image
    docker build -t welcome .
    # Verify the image is there
    docker images
    ```
3. From the `intro` directory, execute: `rezolvr apply -a ./rezolvr-welcome.yaml -e ./env-dev-kube.yaml -s ./state.yaml`. This will generate a file named `welcomeappservice.yaml` in a subdirectory named `out`.

      **NOTE:** Rezolvr doesn't automatically create the output directory. This is intentional. Before running the examples, you should
      create an output directory (e.g. `mkdir out`)


4. Navigate to the output directory, and run the following command: `kubectl apply -f welcomeappservice.yaml`.
5. Get the IP address of the deployed service: `minikube service --url welcomeappservice-service`.
6. Launch a browser and navigate to: http://<ip addr from step 5>:<port from step 5>/message
7. You should see the following text: Hello from Rezolvr!
8. To remove the deployed workload: `kubectl delete -f welcomeappservice.yaml`.


## Remove a Component

Execute: `rezolvr apply -d component.web.app:welcome -e ./env-dev-kube.yaml -s ./state.yaml`. This will remove the component from the state. However, because there are no longer any components defined in the system, output files will not be created.

Note: This applies to both the Kubernetes and Docker drivers.
