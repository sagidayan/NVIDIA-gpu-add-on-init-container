#!/usr/bin/env python3

import os
import shutil
import subprocess
import yaml
import pathlib
from argparse import ArgumentParser

ANNOTATION_PATH = "metadata/annotations.yaml"
ADDON_PATH = "addons/gpu-operator/main"
MANIFESTS = "manifests"
ADDON_NAME = "gpu-operator-certified-addon"
DEPENDENCIES = "metadata/dependencies.yaml"
ROLE_YAML = os.path.join(MANIFESTS, "gpu-operator_rbac.authorization.k8s.io_v1_role.yaml")
ROLEBINDING_YAML = os.path.join(MANIFESTS, "gpu-operator_rbac.authorization.k8s.io_v1_rolebinding.yaml")
CLUSTER_ROLEBINDING_YAML = os.path.join(MANIFESTS, "gpu-operator_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml")
CSV = os.path.join(MANIFESTS, "gpu-operator.clusterserviceversion.yaml")

ROLES_TO_ADD = [{"apiGroups": ["operators.coreos.com"], "resources": ["clusterserviceversions"],
                 "verbs": ["get", "list"]},
                {"apiGroups": ["nfd.openshift.io"], "resources": ["nodefeaturediscoveries"],
                 "verbs": ["get", "list", "create", "patch", "update"]}]
INIT_CONTAINER = [{"name": "gpu-init-container",
                   "image": "quay.io/itsoiref/gpu_init_container:latest",
                   "command": ["/usr/bin/init_run"], }]


def handle_csv(bundle_path, version, prev_version):
    print("Handling csv")
    with open(os.path.join(bundle_path, CSV), "r") as _f:
        csv = yaml.safe_load(_f)
    csv["metadata"]["name"] = f"gpu-operator-certified-addon.v{version}"
    csv["spec"]["replaces"] = f"gpu-operator-certified-addon.v{prev_version}"
    csv["spec"]["install"]["spec"]["deployments"][0]["spec"]["template"]["spec"]["initContainers"] = INIT_CONTAINER
    with open(os.path.join(bundle_path, CSV), "w") as _f:
        yaml.dump(csv, _f)


def copy_deps(addon_path, version, prev_version):
    print("Copy dependency file from old version")
    shutil.copy(os.path.join(addon_path, f"{prev_version}/{DEPENDENCIES}"), os.path.join(addon_path,
                                                                                         f"{version}/{DEPENDENCIES}"))


def handle_annotations(bundle_path, channel):
    print("Handling annotations")
    with open(os.path.join(bundle_path, ANNOTATION_PATH), "r") as _f:
        annotations = yaml.safe_load(_f)
    annotations["annotations"]["operators.operatorframework.io.bundle.channels.v1"] = channel
    annotations["annotations"]["operators.operatorframework.io.bundle.channel.default.v1"] = channel
    annotations["annotations"]["operators.operatorframework.io.bundle.package.v1"] = ADDON_NAME
    with open(os.path.join(bundle_path, ANNOTATION_PATH), "w") as _f:
        yaml.dump(annotations, _f)


def update_role(bundle_path):
    print("Updating role")
    with open(os.path.join(bundle_path, ROLE_YAML), "r") as _f:
        roles = yaml.safe_load(_f)

    roles["rules"].extend(ROLES_TO_ADD)
    with open(os.path.join(bundle_path, ROLE_YAML), "w") as _f:
        yaml.dump(roles, _f)


# TODO remove for 1.9.0
def update_rolebinding_namespaces(bundle_path, namespace):
    role_files = [ROLEBINDING_YAML, CLUSTER_ROLEBINDING_YAML]
    print("Setting namespace in %s", role_files)
    for role_file in role_files:
        with open(os.path.join(bundle_path, role_file), "r") as _f:
            roles = yaml.safe_load(_f)

        roles["subjects"][0]["namespace"] = namespace
        with open(os.path.join(bundle_path, role_file), "w") as _f:
            yaml.dump(roles, _f)


def download_new_bundle(version, addon_path):
    print(f"Downloading new bundle {version} to {addon_path}")
    current_folder = pathlib.Path(__file__).parent.resolve()
    subprocess.check_call(f"export WORKING_DIR={addon_path}; {current_folder}/gitlab_download.sh bundle/{version} "
                          f"&& mv $WORKING_DIR/bundle/{version} $WORKING_DIR && rm -rf $WORKING_DIR/bundle", shell=True)


def create_new_bundle(args):
    addon_path = os.path.join(args.manage_tenants_bundle_path, ADDON_PATH)
    download_new_bundle(args.version, addon_path)
    bundle_path = os.path.join(addon_path, args.version)
    update_rolebinding_namespaces(bundle_path, args.namespace)
    update_role(bundle_path)
    handle_annotations(bundle_path, args.channel)
    handle_csv(bundle_path, args.version, args.prev_version)
    copy_deps(addon_path, args.version, args.prev_version)


if __name__ == '__main__':
    parser = ArgumentParser(
        __file__,
        description='adding new bundle version to gpu-addon'
    )
    parser.add_argument(
        '-mP', '--manage-tenants-bundle-path',
        type=str,
        required=True,
        help='Path to managed tenants repo on the disk'
    )

    parser.add_argument(
        '-c', '--channel',
        type=str,
        default="alpha",
        help='Path to managed tenants repo on the disk'
    )
    parser.add_argument(
        '-n', '--namespace',
        default="redhat-gpu-operator",
        type=str,
        help='Target namespace'
    )
    parser.add_argument(
        '-v', '--version',
        required=True,
        type=str,
        help='New nvidia version'
    )
    parser.add_argument(
        '-pv', '--prev-version',
        required=True,
        type=str,
        help='Previous nvidia version'
    )

    args = parser.parse_args()
    create_new_bundle(args)
