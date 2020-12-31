# Working with CI/CD Pipelines

## Overview
This sample demonstrates integration with a Jenkins Pipeline. It provides a more robust example
of how to leverage Rezolvr in a CI/CD setting.

This example also checks in the fully-resolved deployment files so that a CD tool can pick up the changes and apply them to the target environment.


## The Repositories

To simulate a real world use case, this sample leverages a main deployment repository and the previousy described catalog registry. The URLs are as follows:

Catalog:

`https://github.com/tcrowleyibm/catalog.git`

Catalog Deploy:

`https://github.com/tcrowleyibm/catalog_deploy.git`


## Registry Details

As with the other samples, this sample uses an "insecure" (non-HTTPS) registry. This should NOT be used in production; it's for educational purposes only.

Prior samples provide the details for setting up an insecure registry, but the basic steps:

1. Set up a DNS entry which can be resolved locally (via the `hosts` file).
   It can be any value, but `host.minikube.internal` can provide some compatability
   with Kubernetes-related work.

   The `hosts` file entry should resolve to the localhost IP address.

2. Start up an HTTP-based instance of a Container registry:

   `docker run -d -p 5000:5000 --restart=always --name registry registry:2`


## Jeninks in Docker containers

This sample uses a container-based version of Jenkins. And although the official container image works well by itself, it's necessary to make a custom image in order to build images within the container. (This uses an image named 'DIND' - Docker In Docker.)

The steps for creating and running custom Jenkins containers are as follows:

1. Create a custom Jenkins image similar to the one specified in this sample's `Dockerfile`:

    ```
    FROM jenkins/jenkins:2.263.1-lts-slim
    USER root
    RUN mkdir -p /usr/share/rezolvr/plugins/docker
    RUN mkdir -p /usr/share/rezolvr/plugins/kube
    RUN apt-get update && apt-get install -y apt-transport-https \
          ca-certificates curl gnupg2 \
          software-properties-common
    RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
    RUN apt-key fingerprint 0EBFCD88
    RUN add-apt-repository \
          "deb [arch=amd64] https://download.docker.com/linux/debian \
          $(lsb_release -cs) stable"
    RUN apt-get update && apt-get install -y docker-ce-cli
    COPY ./bin/rezolvr_linux_amd64 /usr/local/bin/rezolvr
    COPY ./bin/plugindocker_linux_amd64.so /usr/share/rezolvr/plugins/docker/plugindocker.so
    COPY ./bin/pluginkube_linux_amd64.so /usr/share/rezolvr/plugins/kube/pluginkube.so
    COPY ./bin/plugins/docker/templates/*.template /usr/share/rezolvr/plugins/docker/templates/
    COPY ./bin/plugins/kube/templates/*.template /usr/share/rezolvr/plugins/kube/templates/
    ENV REZOLVR_PLUGINDIR=/usr/share/rezolvr/plugins/
    USER jenkins
    RUN jenkins-plugin-cli --plugins blueocean:1.24.3
    ```

   Note: This Dockerfile assumes that rezolvr binaries exist in the `/bin` subdirectory.

2. From the Rezolvr root directory, build the file with a command similar to the following:
    
   `docker build -t myjenkins-blueocean -f examples/pipeline/Dockerfile .`

3. Create a network for dind and Jenkins to communicate:

   `docker network create jenkins`

4. Create an instance of dind. Be sure to expose the dev registry as insecure.
(See the `registry` sample for more details.)

    ```
    docker run --name jenkins-docker --rm --detach \
      --privileged --network jenkins --network-alias docker \
      --env DOCKER_TLS_CERTDIR=/certs \
      --volume /Users/tcrowley/apps/jenkins/certs:/certs/client \
      --volume /Users/tcrowley/apps/jenkins/jenkins_home:/var/jenkins_home \
      --publish 2376:2376 docker:dind --insecure-registry host.minikube.internal:5000
    ```

    Comments:
     - dind requires a special registry parameter to let it know not to fail for non-HTTPS connections
     - dind creates private/public keys which must be shared with jenkins. Store them in a volume
     - The Jenkins home directory should also use a volume
     - dind will communicate across port 2376

5. Start the custom Jenkins image:

    ```
    docker run --name jenkins-blueocean --rm --detach \
      --network jenkins --env DOCKER_HOST=tcp://docker:2376 \
      --env DOCKER_CERT_PATH=/certs/client --env DOCKER_TLS_VERIFY=1 \
      --publish 8080:8080 --publish 50000:50000 \
      --volume /Users/tcrowley/apps/jenkins/jenkins_home:/var/jenkins_home \
      --volume /Users/tcrowley/apps/jenkins/certs:/certs/client:ro \
      myjenkins-blueocean
    ```
    Comments:
    - Enable communication with dind via TCP port 2376
    - Instruct Jenkins to use the certificates for process communication
    - Use the same volumes as described above


6. Check the logs for the one time code to start Jenkins:

`docker logs jenkins-blueocean`

7. Continue to set up Jenkins in the normal manner.

8. Once Jenkins is fully initialized, create a new Pipeline and specify a script similar to this sample's `Jenkinsfile`:

    ```
    node {
        stage('Clone main repo') {
            cleanWs()
            sh 'git config --global user.email "someuser@somedomain.co"'
            sh 'git config --global user.name "Jenkie Jenkins"'
            sh 'mkdir -p deploy'
            dir("deploy") {
                checkout([$class: 'GitSCM',
                    branches: [[name: '*/master']],
                    doGenerateSubmoduleConfigurations: false,
                    extensions: [],
                    submoduleCfg: [],
                    userRemoteConfigs: [[credentialsId: '{your-jenkins-credentials-id-here}', 
                                        url: 'https://github.com/tcrowleyibm/catalog_deploy.git']]])
                sh 'git checkout master'
            }
        }
        
        stage('Clone catalog repo') {
            sh 'mkdir -p catalog'
            dir("catalog") {
                checkout([$class: 'GitSCM',
                    branches: [[name: '*/main']],
                    doGenerateSubmoduleConfigurations: false,
                    extensions: [],
                    submoduleCfg: [],
                    userRemoteConfigs: [[credentialsId: '{your-jenkins-credentials-id-here}',
                                        url: 'https://github.com/tcrowleyibm/catalog.git']]])
            }
        }
        
        stage('Resolve dependencies') {
            sh 'pwd'
            sh 'export REZOLVR_PLUGINDIR=/usr/share/rezolvr/plugins/'
            sh 'rezolvr apply -a ./catalog/rezolvr/catalog.yaml -a ./deploy/rezolvr/rezolvr-db.yaml -e ./deploy/rezolvr/env-dev-kube.yaml -s ./deploy/rezolvr/state.yaml -o ./deploy/deploy/'
        }

        stage('Push changes back to git') {
            dir("deploy") {
                withCredentials([usernamePassword(credentialsId: '{your-jenkins-credentials-id-here}',
                    usernameVariable: 'username',
                    passwordVariable: 'password')]){
                    sh 'git add .'
                    sh 'git commit -m "Jenkins update to the deployment files"'
                    sh("git push http://$username:$password@github.com/tcrowleyibm/catalog_deploy.git")
                }
            }
        }

    }
    ```

    Notes:
     - Substitute the credentialsId with your own credential ID. (This comes from Jenkins - not git. However, from within Jenkins, you will use a Git(Hub) personal access token to connect to git.)
     - Directory names can be changed as needed
     - It's possible to add additional repositories to the build process

10. Trigger a build of the Pipeline

11. Review the output

12. Navigate to the `catalog_deploy` repository, and note that the contents have been updated with the generated deployment files. Also, the state file (`state.yaml`) has been updated as well.

