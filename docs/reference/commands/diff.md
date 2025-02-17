<!-- This comment is uncommented when auto-synced to www-kluctl.io

---
title: "diff"
linkTitle: "diff"
weight: 10
description: >
    diff command
---
-->

## Command
<!-- BEGIN SECTION "diff" "Usage" false -->
Usage: kluctl diff [flags]

Perform a diff between the locally rendered target and the already deployed target
The output is by default in human readable form (a table combined with unified diffs).
The output can also be changed to output a yaml file. Please note however that the format
is currently not documented and prone to changes.
After the diff is performed, the command will also search for prunable objects and list them.

<!-- END SECTION -->

## Arguments
The following sets of arguments are available:
1. [project arguments](./common-arguments.md#project-arguments)
1. [image arguments](./common-arguments.md#image-arguments)
1. [inclusion/exclusion arguments](./common-arguments.md#inclusionexclusion-arguments)

In addition, the following arguments are available:
<!-- BEGIN SECTION "diff" "Misc arguments" true -->
```
Misc arguments:
  Command specific arguments.

      --force-apply                                 Force conflict resolution when applying. See documentation for
                                                    details
      --force-replace-on-error                      Same as --replace-on-error, but also try to delete and
                                                    re-create objects. See documentation for more details.
      --helm-insecure-skip-tls-verify stringArray   Controls skipping of TLS verification. Must be in the form
                                                    --helm-insecure-skip-tls-verify=<credentialsId>, where
                                                    <credentialsId> must match the id specified in the helm-chart.yaml.
      --helm-key-file stringArray                   Specify client certificate to use for Helm Repository
                                                    authentication. Must be in the form
                                                    --helm-key-file=<credentialsId>:<path>, where <credentialsId>
                                                    must match the id specified in the helm-chart.yaml.
      --helm-password stringArray                   Specify password to use for Helm Repository authentication.
                                                    Must be in the form
                                                    --helm-password=<credentialsId>:<password>, where
                                                    <credentialsId> must match the id specified in the helm-chart.yaml.
      --helm-username stringArray                   Specify username to use for Helm Repository authentication.
                                                    Must be in the form
                                                    --helm-username=<credentialsId>:<username>, where
                                                    <credentialsId> must match the id specified in the helm-chart.yaml.
      --ignore-annotations                          Ignores changes in annotations when diffing
      --ignore-labels                               Ignores changes in labels when diffing
      --ignore-tags                                 Ignores changes in tags when diffing
      --no-obfuscate                                Disable obfuscation of sensitive/secret data
  -o, --output-format stringArray                   Specify output format and target file, in the format
                                                    'format=path'. Format can either be 'text' or 'yaml'. Can be
                                                    specified multiple times. The actual format for yaml is
                                                    currently not documented and subject to change.
      --render-output-dir string                    Specifies the target directory to render the project into. If
                                                    omitted, a temporary directory is used.
      --replace-on-error                            When patching an object fails, try to replace it. See
                                                    documentation for more details.

```
<!-- END SECTION -->

`--force-apply` and `--replace-on-error` have the same meaning as in [deploy](./deploy.md).
