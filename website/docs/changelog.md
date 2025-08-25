# Changelog

## v2.2 - June 11, 2025

### New controllers

- Floating IP
- Server Group

### New features

- The subnet controller gained support for updating subnets. The rest of the controllers will follow in the next releases.
- Added ability to specify which project to create resources in, for all networking controllers.

### Bug fixes

- Add an optional dependency of subnet on router: a subnet will now wait for the referenced router to be available before proceeding with creation. Likewise, the router won't be deleted until all subnets that depend on it are themselves deleted (Fixes [#316](https://github.com/k-orc/openstack-resource-controller/issues/316)).
- Fix the import of images, where the status may not reflect the real status of the resource.
- Make the deletion of router interface more robust (Fixes [#378](https://github.com/k-orc/openstack-resource-controller/issues/378)).
- Add API validation to limit ExternalGateways to one until multiple gateways is effectively implemented in ORC (Fixes [#416](https://github.com/k-orc/openstack-resource-controller/issues/416)).
- Selectively stop populating status fields with zero values (Fixes [#188](https://github.com/k-orc/openstack-resource-controller/issues/188)).

### Update considerations

Although we don't guarantee that ORC runs fine on OpenStack versions that are no longer maintained by the OpenStack community, we've merged a change in this release that requires Nova from Stein. To the best of our knowledge, OpenStack Stein now becomes the minimum required version of OpenStack.

## v2.1 - May 02, 2025

Release 2.1 marks the continuation of our efforts to stabilize and consolidate ORC.

We are now building multi-platform container images, supporting `amd64`, `arm64`, `ppc64le` and `s390x` platforms.

This release also brings initial support for Keystone resources in the form of
the Project controller.

### New controllers

- Project

### New features

- ORC now passes the Kubernetes ReconcileID as the Request ID for all OpenStack API calls allowing to correlate the ORC and OpenStack logs, making troubleshooting much easier.
- The image controller is now more consistent with other controllers.
- Support setting new properties in the image controller: the `hw_rng_model` and `hw_qemu_guest_agent` hardware properties, and the `os_distro`, `os_version`, `architecture` and `hypervisor_type` properties.
- The port controller now has the ability to set port security and vnic type.

### Bug fixes

- The user-agent header now includes the ORC versions, helping identify specific versions in logs for better support and debugging.
- Add ability to create router interfaces for HA or DVR routers (Fixes [#330](https://github.com/k-orc/openstack-resource-controller/issues/330)).
- The status of servers is now reflected more accurately (Fixes [#280](https://github.com/k-orc/openstack-resource-controller/issues/280)).


## v2.0 - Mar 28, 2025

This release introduces several new controllers, expanding ORC's capabilities
beyond the original image controller. With this update, ORC now offers
a robust, stable core and a comprehensive end-to-end (e2e) test suite, making
it easier to create new controllers while ensuring quality and reliability.

Version 2.0 highlights the capabilities of ORC and the direction the project
wants to take. The API is still alpha and may change frequently.

### New controllers

- Flavor
- Network
- Port
- Router
- Security Group
- Server
- Subnet

### Breaking changes

```
github.com/k-orc/openstack-resource-controller/api/v1alpha1
  Incompatible changes:
  - ImageFilter.Name: changed from *string to *OpenStackName
  - ImageFilter: old is comparable, new is not
  - ImageProperties.MinDiskGB: changed from *int to *int32
  - ImageProperties.MinMemoryMB: changed from *int to *int32
  - ImagePropertiesHardware.CPUCores: changed from *int to *int32
  - ImagePropertiesHardware.CPUSockets: changed from *int to *int32
  - ImagePropertiesHardware.CPUThreads: changed from *int to *int32
  - ImageResourceStatus.Status: changed from *string to string
  - ImageResourceStatus: old is comparable, new is not
  - ImageStatus.DownloadAttempts: changed from *int to *int32
  - ImageStatusExtra.DownloadAttempts: changed from *int to *int32
  - OpenStackDescription: removed
```

## v1.0 - Dec 19, 2024

First public version for a standalone ORC.

This preliminary release is not intended for general consumption. Its primary
purpose is to satisfy the existing use case of
[cluster-api-provider-openstack](https://github.com/kubernetes-sigs/cluster-api-provider-openstack)
without creating any new APIs.

ORC v1.0.0 contains an API and controller for creating and deleting Glance images.

### New controllers

- Image
