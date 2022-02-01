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
CSV_SUFFIX = "clusterserviceversion.yaml"

ROLES_TO_ADD = [{"apiGroups": ["operators.coreos.com"], "resources": ["clusterserviceversions"],
                 "verbs": ["get", "list"]},
                {"apiGroups": ["nfd.openshift.io"], "resources": ["nodefeaturediscoveries"],
                 "verbs": ["get", "list", "create", "patch", "update"]}]
INIT_CONTAINER = [{"name": "gpu-init-container",
                   "image": "quay.io/edge-infrastructure/nvidia-gpu-add-on-init-container:latest",
                   "command": ["/usr/bin/init_run"], }]


def get_csv_file(bundle_path):
    manifests = os.path.join(bundle_path, MANIFESTS)
    for file in os.listdir(manifests):
        if file.endswith(CSV_SUFFIX):
            return os.path.join(manifests, file)


def handle_csv(bundle_path, version, prev_version):
    print("Handling csv")
    csv_file = get_csv_file(bundle_path)
    with open(csv_file, "r") as _f:
        csv = yaml.safe_load(_f)

    csv["metadata"]["annotations"]["olm.skipRange"] = f">=0.0.1 <{version}"
    csv["metadata"]["name"] = f"gpu-operator-certified-addon.v{version}"
    csv["spec"]["replaces"] = f"gpu-operator-certified-addon.v{prev_version}"
    csv["spec"]["install"]["spec"]["deployments"][0]["spec"]["template"]["spec"]["initContainers"] = INIT_CONTAINER
    csv["spec"]["install"]["spec"]["permissions"][0]["rules"].extend(ROLES_TO_ADD)
    with open(csv_file, "w") as _f:
        yaml.dump(csv, _f)


def copy_deps(addon_path, version, prev_version):
    print("Copy dependency file from old version")
    shutil.copy(os.path.join(addon_path, f"{prev_version}/{DEPENDENCIES}"), os.path.join(addon_path,
                                                                                         f"{version}/{DEPENDENCIES}"))


def handle_annotations(bundle_path, channel, namespace):
    print("Handling annotations")
    with open(os.path.join(bundle_path, ANNOTATION_PATH), "r") as _f:
        annotations = yaml.safe_load(_f)
    annotations["annotations"]["operators.operatorframework.io.bundle.channels.v1"] = channel
    annotations["annotations"]["operators.operatorframework.io.bundle.channel.default.v1"] = channel
    annotations["annotations"]["operators.operatorframework.io.bundle.package.v1"] = ADDON_NAME
    annotations["annotations"]["operatorframework.io/suggested-namespace"] = namespace
    with open(os.path.join(bundle_path, ANNOTATION_PATH), "w") as _f:
        yaml.dump(annotations, _f)


def donwload_new_bundle_from_rh_certified(version, addon_path):
    operators_folder = "operators"
    gpu_folder = os.path.join(operators_folder, "gpu-operator-certified")
    current_folder = pathlib.Path(__file__).parent.resolve()
    subprocess.check_call(f"{current_folder}/github_download.py --org=redhat-openshift-ecosystem "
                          f"--repo=certified-operators --branch=main "
                          f"--folder='{gpu_folder}/v{version}' -w {addon_path}"
                          f" && mv {addon_path}/{gpu_folder}/v{version} {addon_path}/{version} "
                          f" && rm -rf {addon_path}/{operators_folder}", shell=True)


def download_new_bundle(version, addon_path):
    print(f"Downloading new bundle {version} to {addon_path}")
    current_folder = pathlib.Path(__file__).parent.resolve()
    move_to_folder = addon_path
    if version.startswith("v"):
        move_to_folder = os.path.join(addon_path, version[1:])

    subprocess.check_call(f"export WORKING_DIR={addon_path}; {current_folder}/gitlab_download.sh bundle/{version} "
                          f"&& mv $WORKING_DIR/bundle/{version} {move_to_folder} && rm -rf $WORKING_DIR/bundle", shell=True)


def create_new_bundle(args):
    version = args.version
    addon_path = os.path.join(args.manage_tenants_bundle_path, ADDON_PATH)
    if args.rh_certified:
        donwload_new_bundle_from_rh_certified(version, addon_path)
    else:
        download_new_bundle(version, addon_path)
    if version.startswith("v"):
        version = version[1:]

    bundle_path = os.path.join(addon_path, version)
    handle_annotations(bundle_path, args.channel, args.namespace)
    handle_csv(bundle_path, version, args.prev_version)
    copy_deps(addon_path, version, args.prev_version)


if __name__ == '__main__':
    parser = ArgumentParser(
        __file__,
        description='adding new bundle version to gpu-addon'
    )
    parser.add_argument(
        '-mP', '--manage-tenants-bundle-path',
        required=True,
        help='Path to managed tenants repo on the disk'
    )

    parser.add_argument(
        '-rh', '--rh-certified',
        action="store_true",
        help='Download bundle from redhat certified addon repo'
    )

    parser.add_argument(
        '-c', '--channel',
        default="alpha",
        help='Path to managed tenants repo on the disk'
    )
    parser.add_argument(
        '-n', '--namespace',
        default="redhat-gpu-operator",
        help='Target namespace'
    )
    parser.add_argument(
        '-v', '--version',
        required=True,
        help='New nvidia version'
    )
    parser.add_argument(
        '-pv', '--prev-version',
        required=True,
        help='Previous nvidia version'
    )

    args = parser.parse_args()
    create_new_bundle(args)
