# Rezolvr

Resolve complex deployment details in a containerized, microservices-based world

## Introduction

Many teams have transformed their applications from a monolithic architecture to a microservices-based one. There are many
benefits associated with this type of transformation, such as:
 - Developers can independently develop and push changes without long, complicated build and test cycles.
 - Teams may collaborate on well-defined interfaces, which allow them to maintain their independence.
 - An operations team can perform finely tuned scaling on the system; it's possible to add resources to the parts of the
   system that need it most.
 - As the application progresses through the various environments (dev, test, prod, etc), supporting components can be swapped out,
   to better fit the purpose of a particular environment. For example, a development environment might need a disposable database
   or message queue, but the production environment requires far more resiliant instances of these components.

Although a microservices-based architecture addresses these concerns, other issues arise:
 - As features are added, and the application evolves, it's difficult to manage the dependencies between the various parts of the system.
 - Most organizations have a number of different technologies and environments, so a platform-specific solution will have
   a limited impact on the organization as a whole.
 - Needed components can differ significantly between various environments. Without a mechanism for managing dependencies,
   it's easy to lose track of the similarities and differences between them.
 - Complex communication between the development and operations teams becomes more critical. Without automation, this commuication
   becomes burdensome and unwieldy, generally devolving into countless meetings.

Stated another way: Most monolithic applications resolve their internal dependencies as part of the build process.
However, microservices-based applications don't have an equivalent mechanism; missing dependencies arise after changes have been deployed. In other words, it's possible to build, test and deploy a component of the solution successfully, and still cause the system to break.

As with compilers, there needs to be a way to catch dependency problems before it's too late. That's the purpose of `rezolvr`:
 - Resolve dependencies before deploying updates
 - Feel confident that all needs have been accounted for
 - Work with existing CI/CD tools and processes
 - Describe resources and their dependencies in a platform-neutral way

Rezolvr also goes beyond traditional template engines: It allows a team to identify dependencies which must be resolved before a deployment is considered valid. As an example, a microservice may depend upon a relational database and an additional microservice. There's nothing in a traditional YAML file that ensures these other resources exist, so it's possible to have a bad deploy. Rezolver prevents this situation by ensuring the target environment can meet all of a component's dependencies.

### How Does it Work?

A development team defines a `Component` within their source code repository. A component is a single YAML file
which contains three types of resources:
 - **Needs** - The resources that this component depends upon. This can be things like environment variables, secrets, or
   even databases or other microservices
 - **Provides** - The resources that this component makes available to other components within the system, such as an endpoint,
   or a service.
 - **Uses** - The resources internally used by the component. This consistes of environment, configuration, and secret-based
   data, but it can contain other resources as well.

A simple example of a configuration file is listed here:

```
name: welcome
type: component.web.app
description: A simple welcome web app component
needs:
  - type: environment.properties
    name: appEnvProps
    description: Application properties
    params:
      - name: app_message
        required: true
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
        formula: '{{.Component.Name}}'  # Formulas are used to calculate the value of a parameter. They use go template syntax.
uses:
  - type: environment
    params:
      - name: APP_MSG
        formula: '{{with(index .Needs "environment.properties:appEnvProps")}}{{.Params.app_message.Value}}{{end}}'
```

When executed, Rezolvr attempts to resolve all of the needs within the system. It achieves
this by combining the microservices configuration files with:
 - The configuration files associated with other microservices within the application.
 - A special configuration file that's environment-specific (called the environment file).
 - A file which represents the current state of the system (`state.yaml`).
 - A platform-specific plugin (e.g. Docker Compose, Kubernetes, or other platform).

The output of the executable is a collection of platform-specific deployment files.

As an example, to generate Kubernetes deployment file(s) from the above, a Kubernetes-specific environment file is used:

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
uses: # This is really ".Platform" - items which are specific to Kubernetes
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

The following command is used to resolve the dependencies:

`rezolvr apply -a welcome.yaml -e env-dev-kube.yaml -s state.yaml -o ./out/`

**Note:** If the `state.yaml` file doesn't already exist, `rezolvr` will create it.

After a successful run, a Kubernetes deployment file will be created in the `./out/` subdirectory.


## Installation and usage

As of now, there are no pre-built executables for Rezolvr. However, it's pretty straightforward to build an executable from source. Assuming the most recent version of `go` is available, the installation process is as follows:

    `make install`

This creates the executable (`rezolvr`) as well as two plugins - one for Kubernetes, and one for Docker. By default, plugins are stored in a user's home directory (`~/.rezolvr`).


## CI/CD - Typical Workflow

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