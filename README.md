
# sidecar - update configmaps and secrets in running pod.

This is a docker container intended to run inside a kubernetes cluster to collect config maps and secrets with specified labels and store the included files in an local folder. When using ConfigMap as a `subPath` volume mount, there will be no changes until the pod is manually restarted. Then if you want update config map and keep running pods this container will help you.

## Quick Starts

- [Example](./example/example.yaml)
- [Image](https://hub.docker.com/repository/docker/vinhph2/sidecar)

# Features

- Extract files from config maps to floders
- Filter based on labels
- Update/Delete on change of configmap
- Enforce unique filenames

## Documentation
Example for a simple deployment can be found in [`example.yaml`](./example/example.yaml)
```
Usage:
  sidecar [flags]

Flags:
  -c, --config string        file config to define resource to watch
  -d, --debug                if the flag debug is true then application will slepp 1 minute before exit when it have error.
  -h, --help                 help for sidecar
  -k, --kube-config string   Where is the file kubeconfig? If you run the application in pod, please ingnore this flag.
  -T, --sleep-time int       How many seconds to next check (default 3)
```
## Configuration watch resoruces
file config define resource to watch is yaml type with define as follow:
- `resource`
    - description: is list resource for side will watch or get
    - required: true
    - type: list each elememt has folow attitudes:
        - `namespace`: is the k8s namespace which resource belong. If you want to get on all namespace just ingnore this attitude.
            - type: string
        - `path`: folder wich file will be save after get.
            - type: path
            - default: `\tmp`
        - `type`: type of resource. it must be `configmap`, `secret` or `both`.
            - type: string
            - default: `configmap`
        - `method`: which action will do with this resource. `get` will get in the first run. `watch` will get in the first run and watch change in later. 
            - type: string
            - default: `watch`
        - `labels`: define labels to match on resources. 
            - type: list which element is:
                - `name`: name of label
                    - type: string
                    - require: true
                - `value`: value of label
                    - type: string
                    - require: true
        - `script_inlines`: This is a list of command strings. They will excute after resource change. 
            - type: list with each element is string
## Frequently Asked Questions

If you have any question or insue please contact us at [support@vngcloud.vn](mailto:support@vngcloud.vn) or create new issue.

## License
[Mozilla Public License v2.0](./LICENSE)