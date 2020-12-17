# Rezolvr

Resolve complex deployment details in a containerized, microservices-based world

## Introduction

When it comes to modern workload deployment, there's lot of different "stuff" that's jammed into a single (YAML) file. This makes it hard to separate operational stuff from developer stuff. 

Rezolvr asks: Why jam all of the details into the same file? You wouldn't put all of your source code in one file, so why would you put all of your deployment details in one?

Rezolver creates a well-defined interface between development and ops teams. Each team is responsible for a subset of the information. Rezolver then "compiles" this information together into a complete set of deployment files - without violating the principle of separation of concerns.

Rezolvr also goes beyond traditional template engines: It allows a team to identify dependencies which must be resolved before a deployment is considered valid. As an example, a microservice may depend upon a relational database and an additional microservice. There's nothing in a traditional YAML file that ensures these other resources exist, so it's possible to have a bad deploy. Rezolver prevents this situation by ensuring the target environment can meet all of a component's dependencies.


## Installation and usage

As of now, there are no pre-built executables for Rezolvr. However, it's pretty straightforward to build an executable from source. Assuming the most recent version of `go` is available, the installation process is as follows:

    `make install`

This creates the executable (`rezolvr`) as well as two plugins - one for Kubernetes, and one for Docker. By default, plugins are stored in a user's home directory (`~/.rezolvr`).

## Concepts and Examples

A complete set of examples can be found in the `/examples` subdirectory. However, at a high level, Rezolvr combines four different things:

 - Components
 - Environment / Platform specifics
 - A platform-specific driver
 - State

### Components
A development team defines a `Component`. A component is a single YAML file which contains three types of resources:
   - `provides` - Resources which are made available to other components within a larger deployment
   - `uses` - Resources used by the component
   - `needs` - Resources which must exist for the component to be considered valid

   The following sample demonstrates a properly-formatted component:

   ```
    name: welcome
    type: component.web.app
    description: A simple welcome web app component
    provides:
    - type: service.web.app
        name: welcomeappservice
        description: A basic application service to ensure the site is running
        params:
        - name: path
            value: /message
        - name: port
            value: 3000
        - name: imageName
            formula: '{{.Component.Name}}'
    uses:
    - type: environment
        params:
        - name: APP_MSG
            formula: '{{with(index .Needs "environment.properties:appEnvProps")}}{{.Params.app_message.Value}}{{end}}'
    needs:
    - type: environment.properties
        name: appEnvProps
        description: Application properties
        params:
        - name: app_message
            required: true
   ```

Briefly, this component `needs` an environment variable named `app_message`. It `uses` that environment variable internally (as an environmet variable named `APP_MSG`). Finally, it `provides` an endpoint that others may reference within their components.

### Environment / platform specifics

An operations team defines one or more environment files. (Generally speaking, one file per target environment.)

An environment file is a single YAML file which contains two types of resources:
   - `provides` - Environment-specific details (e.g. registry details, port numbers, etc.).
   - `uses` - Environment / platform-specific deployment details (e.g. number of instances to run, pull policies, etc.).

An environment file also specifies the name of a `driver` (or plugin) which is used to generate deployment files.

The following sample demonstrates a properly-formatted component:

   ```
    name: devMinikubeEnv
    type: resource.environment
    driver: kube
    description: Environment variables provided by the local development environment
    provides:
    - type: environment.properties
        name: appEnvProps
        description: Application Properties
        params:
        - name: app_message
            value: 'Hello from Rezolvr!'
    uses: # This is really ".Platform"
    - type: platform.settings
        name: default
        params:
        - name: numInstances
            value: 1
        - name: imagePullPolicy
            value: Never
        - name: serviceType
            value: NodePort
   ```

### Plugins

Rezolvr is not tied to a single platform (such as Docker Compose, Kubernetes, etc.). Instead, a plugin transforms input data into platform-specific deployment files. In the above environment sample, the `kube` plugin has been specified.

By default, Rezolvr comes with both Kubernetes and Docker plugins.

### State

A complete deployment will generally consist of several components. To manage the relationships between the components, Rezolver maintains state information. This information is stored in a YAML file. The state file contains:
 - All previously consumed components
 - Environment specifics
 - All resolved values for parameters

### Sample usage

To combine / compile the separate files, invoke Rezolver as follows:

`rezolvr apply -a ./rezolvr-welcome.yaml -e ./env-dev-kube.yaml -s ./state.yaml -o ./out/`

Comments:
 - `apply` - Apply changes to the current state of the system
 - `-a ./rezolvr-welcome.yaml` - Add a component to the system
 - `-e ./env-dev-kube.yaml` - Target the development instance of our Kubernetes environment(s)
 - `-s ./state.yaml` - Save state information in a file named `state.yaml`. Create the file if it doesn't already exist
 - `-o ./out/` - All fully-resolved deploymet files should be placed in the `./out/` directory.

## Typical workflow

Rezolvr is built to work with CI/CD pipelines. It's generally used just after the CI steps and just before the CD steps.
A typical workflow is as follows:
 - Developers commit code to their component's repository
 - The code is built and tested by the component's CI pipeline
 - The code contains one or more component files, generally in a top-level directory named `/rezolvr`
 - A pipeline checks out both the component's repo as well as a deployment repo
 - The pipeline combines Rezolvr details from the component repo and the deployment repo, and generates platform-specific deployment files
 - State is updated as well
 - Deployment files and the state are committed to the deployment repo
 - The overall CD pipeline is triggerd by the git commit, and the deployment files are applied to the target platform

For additional information - including a sample Jenkinsfile - please refer to the `pipeline` example.

## Additional examples

Additional examples can be found in the `/examples` directory.

## Changelog

0.0.1 - Initial alpha version