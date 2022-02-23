#!/usr/bin/env python3

import os
import base64
import shutil
import pathlib
from argparse import ArgumentParser
from github import Github
from github import GithubException


def get_sha_for_tag(repo, tag):
    """
    Returns a commit PyGithub object for the specified repository and tag.
    """
    branches = repo.get_branches()
    matched_branches = [match for match in branches if match.name == tag]
    if matched_branches:
        return matched_branches[0].commit.sha

    tags = repo.get_tags()
    matched_tags = [match for match in tags if match.name == tag]
    if not matched_tags:
        raise ValueError('No Tag or Branch exists with that name')
    return matched_tags[0].commit.sha


def download_directory(repo, sha, server_path, working_dir):
    """
    Download all contents at server_path with commit tag sha in
    the repository.
    """
    print("Working dir", working_dir, "server path", server_path)
    if not os.path.isdir(working_dir):
        os.makedirs(working_dir)

    contents = repo.get_dir_contents(server_path, ref=sha)
    for content in contents:
        print("Processing %s" % content.path)
        if content.type == 'dir':
            print("creating folder", os.path.join(working_dir, content.path))
            if not os.path.exists(os.path.join(working_dir, content.path)):
                print("creating folder", os.path.join(working_dir, content.path))
                os.makedirs(os.path.join(working_dir, content.path))
            download_directory(repo, sha, content.path, working_dir)
        else:
            try:
                path = content.path
                file_content = repo.get_contents(path, ref=sha)
                file_data = base64.b64decode(file_content.content)
                with open(os.path.join(working_dir, content.path), "w+", encoding="utf-8") as file_out:
                    file_out.write(file_data.decode('ascii'))
            except (GithubException, IOError) as exc:
                print('Error processing', content.path, exc)


if __name__ == "__main__":

    parser = ArgumentParser(
        __file__,
        description='adding new bundle version to gpu-addon'
    )
    parser.add_argument(
        '-t', '--token',
        default="",
        help='github token'
    )

    parser.add_argument(
        '-o', '--org',
        required=True,
        help='Github org'
    )
    parser.add_argument(
        '-r', '--repo',
        required=True,
        help='Repo'
    )
    parser.add_argument(
        '-f', '--folder',
        required=True,
        help='folder to download'
    )
    parser.add_argument(
        '-b', '--branch',
        default="main",
        help='branch'
    )

    parser.add_argument(
        '-w', '--working-dir',
        default="",
        help='which directory to use as working dir'
    )

    args = parser.parse_args()
    github = Github(args.token or None)
    organization = github.get_organization(args.org)
    repository = organization.get_repo(args.repo)
    sha = get_sha_for_tag(repository, args.branch)

    if not args.working_dir:
        args.working_dir = pathlib.Path(__file__).parent.resolve()

    if args.folder.startswith("/"):
        args.folder = args.folder[1:]
    content_path = os.path.join(args.working_dir, pathlib.Path(args.folder).parts[0])
    if os.path.exists(content_path):
        shutil.rmtree(content_path)

    download_directory(repository, sha, args.folder, args.working_dir)
