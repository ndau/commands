#!/usr/bin/python

"""
This script enforces custom rules prior to commiting changes.

If you want it for personal use only:
- Put this file anywhere you want.
- Run this: ln -s <where-you-put-it>/pre-commit.py <your-repo>/.git/hooks/pre-commit

If you want everyone to get the rules enforced:
- Put this file at the root of your repo, commit it.
- Then put the following in your .git/hooks/pre-commit file:
#!/bin/sh
./pre-commit.py
"""

__author__ = 'Eric Schmidt'
__copyright__ = 'Copyright 2018, ndev'
__version__ = '1.0'


import os
import re
import subprocess
import sys


# We'll prevent any file containing REMOVE ME (no space) from being committed.
# Split up that string here so we don't break the rule ourselves.
REMOVE_ME = 'REMOVE' + 'ME'


def search(regex, fullpath):
    """
    Return True if and only if the given regex matches any line of the given file.
    """

    p = re.compile(regex)
    for line in open(fullpath):
        if p.search(line):
            return True

    return False


def die(message, filename):
    """
    Log an error message and the given file name, then exit with a failure error code.
    """

    print('Cannot commit due to rule violation in file:', filename)
    print(message)

    sys.exit(1)


def main():
    """
    Enforce our rules for every staged file.
    """

    # Get list of files staged for commit.
    process_args = [
        'git',
        'diff-index',
        'HEAD',
        '--cached',
        '--name-only'
    ]
    output = subprocess.check_output(process_args)
    files = output.split()

    # Get repo root directory.
    process_args = [
        'git',
        'rev-parse',
        '--show-toplevel'
    ]
    output = subprocess.check_output(process_args)
    rootdir = output.rstrip('\n')

    for filename in files:
        fullpath = os.path.join(rootdir, filename)

        # Ignore deleted files in the change list.
        if not os.path.exists(fullpath):
            continue

        # lint
        if fullpath.endswith('.go'):
            process_args = [
                'golint',
                fullpath
            ]
            output = subprocess.check_output(process_args)
            if output != '':
                die('Go lint is not satisfied:\n{0}'.format(output), filename)

        # REMOVE_ME
        if fullpath.endswith('.go') or \
           fullpath.endswith('.sh') or \
           fullpath.endswith('.py') or \
           fullpath.endswith('.ipynb'):
            if search('\\b{0}\\b'.format(REMOVE_ME), fullpath):
                die('A line containing "REMOVE' + 'ME" was found', filename)


# Call main() if we're being run from the command line.
if __name__ == '__main__':
    main()
