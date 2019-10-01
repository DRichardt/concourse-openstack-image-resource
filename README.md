# Concourse Resource for Openstack Images

checks and manages openstack images by a concourse

https://godoc.org/github.com/DRichardt/concourse-openstack-image-resource/pkg/resource

## Adding to your pipeline

To use Concourse Resource , you have to add a new resource_type for the concourse openstack resource

```
resource_types:
- name: image
  type: docker-image
  source:
    repository: will be filled when resource is finished
```

## Source Configuration

* `OS_AUTH_URL`: *Required.* The url for authentication (Keystone)

* `OS_USERNAME`: *Required.* The username to authenticate to openstack. 

* `OS_PASSWORD`: *Required.* The Password to authenticate to openstack. 

* `OS_REGION`: *Required.* The Openstack Region to use.

* `OS_PROJECT_NAME`: *Required.* The Openstack Project to use.

* `Imagename`: *Required.* The Imagename to filter
