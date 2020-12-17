# REZOLVR - Separating Concerns in a Containerized, Cloud World

This README file contains background information regarding the purpose of Rezolvr. For a more concise overview,
and a tutorial on how to use Rezolvr, please refer to README.md

## What Is It?

Over the past several years, there have been a few major changes in the way software is developed
and deployed.

 - Monolithic systems have given way to microservices-based systems
 - Applications have become "containerized"
 - The operations ecosystem has become more automated, via infrastructure as code

While these changes have significantly improved the way software is developed and operationalized, some
issues have arisen. One key issue has been the mixing of concerns; most notably through YAML configuration
files.

There isn't really a problem with the YAML specification, but more of a problem with how they're being used.
Specifically, a single file contains multiple concerns, such as:

 - Services which other components may use
 - Hardware requirements (RAM, storage, connectivity)
 - Platform-specific details (dev, test, prod)
 - Relationships between components

Intermingling development concerns and operational concerns results in tight coupling. This tight coupling
makes it difficult to coordinate how teams work with each other. And even though the concerns are tightly coupled,
most modern platforms don't verify that all of a workload's dependencies are met. In other words, it's possible
to have a perfectly valid set of deployment files, which deploy successfully, but result in a broken system.

Rezolvr attempts to address this by separating workload details:
- Developers focus on what they do best - write code and define services
- Dev ops engineers focus on what they do best - operational automation and management

Rezolvr combines this information and identifies unmet dependencies. By combining (or resolving) the various
pieces of the system, Rezolvr helps ensure that workload deployments are successful.

## How Does It Work?

 - Developers create a special configuration (as code) file. The file contains:
   - Resources provided by the code / microservice
   - Resources used internally by the code / microservice
   - Dependencies, or resources needed to operate correctly
 - Operational teams create an additional configuration file, which contain:
   - Target envrionment specifics (dev / test / etc)
   - Hardware needs
   - Operational settings
 - A set of plugins provide platform-specific details to generate target YAML files
   - Platform specifics
   - Support for Docker and Kubernetes
   - Additional plugins can be added via Rezolvr's open architecture

Along with the state of the target platform, Rezolvr combines these parts, and generates deployment files which
are fully resolved, and ready to be deployed.

## Between CI and CD

Rezolvr can be used at any point in the development / delivery process. However, for devops pipelines, it sits
between the traditional continuous integration step and the continuous delivery step. That is, it uses source materials
from development, and state / configuration information for a target platform to generate deployment files. The
deployment files can then be used for a traditional continuous delivery process. The files may also be checked into
git for gitops-based pipelines.

