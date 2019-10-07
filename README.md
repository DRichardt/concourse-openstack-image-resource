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


## Params

* `file`: *Rquired* Path to file to upload

* `delete_broken_images` *Optional*  Default: true deletes Image if Checksumcheck fails

* `container_format`: *Required* Container Format

* `disk_format`: *Required* Disk Format

* `visibility` *Required* Visability (private|shared|public|community)

* `protected` *Required* Proteced (Bool)

*  `min_disk`: Minimal RAM needed (Integer)

*  `min_ram`: Minimal Disk needed (Integer)

*  `properties_by` : *Required* How to get properties (direct | file)

*  `properties`: JSON Object of additional properties (string)

    properties_by = direct (`"{\"architecture\": \"x86_64\",\"buildnumber\": \"20190926.2\",\"git_branch\":\"master\",\"hw_disk_bus\":\"scsi\",\"hw_video_ram\":\"16\",\"hw_vif_model\":\"VirtualVmxnet3\",\"hypervisor_type\":\"vmware\",\"os-version\":\"15.3\",\"vmware_adaptertype\":\"paraVirtual\",\"vmware_disktype\":\"streamOptimized\",\"vmware_ostype\":\"sles12_64Guest\"}"`)

    properties_by = string (/tmp/properties.json)