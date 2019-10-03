#! /usr/bin/env python3

import argparse
import datetime
import glob
import os
import re
import textwrap

from string import Template

copyright_template = Template(
    """${stc} ${header}
${c} Copyright ${yeartext} ${holder}. All Rights Reserved.
${c}
${c} Licensed under ${license_name} (the "License").  You may not use
${c} this file except in compliance with the License.  You can obtain a copy
${c} in the file LICENSE in the source distribution or at
${c} ${license_link}
${c} ${footer}
${endc}
"""
)

# The header / footer values are chosen to be unlikely to occur in normal
# code.
values = dict(
    header="----- ---- --- -- -",
    footer="- -- --- ---- -----",
    year=datetime.datetime.now().year,
    holder="Oneiro NA, Inc",
    license_name="the Apache License 2.0",
    license_link="https://www.apache.org/licenses/LICENSE-2.0.txt",
)

comments = dict(
    go=dict(stc="//", c="//", endc="//"),
    js=dict(stc="/*", c=" *", endc=" */"),
    py=dict(stc="##", c="##", endc="##"),
    chasm=dict(stc=";;", c=";;", endc=";;"),
)


def setupArgs():
    """ sets up the argument parser with all of our options """
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=textwrap.dedent(
            """
            This program can be given a set of files (specified as one or more glob
            patterns) and can update those files with new copyright information.

            It looks for comments previously inserted by this program and will
            overwrite them if they're out of date. Otherwise, it inserts them
            at the top of the file (but after any hashbang line).

            It attempts to infer the comment style from the file extension, but
            can be forced to use a particular file type with the lang switch.

            It's smart about years and can update files with a new year.
            """
        ),
        epilog=textwrap.dedent(
            """
            Examples:
                addCopyright.py --files "./**/*.go"
            """  # noqa: E501
        ),
    )
    parser.add_argument(
        "--check",
        default=False,
        action="store_true",
        help="Makes no changes, but checks the files given to see if their copyright"
        "information is up to date; if not, exits with errorlevel=2.",
    )
    parser.add_argument(
        "--files",
        nargs="*",
        default=[],
        help="Specifies filenames or glob patterns of files to examine and update",
    )
    parser.add_argument(
        "--exts",
        nargs="*",
        default=[],
        help="Specifies file extensions (like 'go' or 'js') to examine",
    )
    parser.add_argument(
        "--verbose",
        default=False,
        action="store_true",
        help="prints the filenames that require updates",
    )
    parser.add_argument(
        "--year",
        type=int,
        default=datetime.datetime.now().year,
        help="ensure this year is listed in the copyright text (default this year)",
    )
    parser.add_argument(
        "--lang",
        default="",
        help="forces the language type for comments (go/js/py/chasm)",
    )
    return parser


def generateFileList(globs):
    allfiles = set()
    for g in globs:
        gs = glob.glob(g, recursive=True)
        allfiles.update(gs)
    return allfiles


def findCopyrightInfo(lines):
    ctr = 0
    headerline = -1
    for ctr in range(len(lines)):
        if headerline == -1 and values["header"] in lines[ctr]:
            headerline = ctr
            break

    if headerline == -1:
        return None

    # we found the header, now we can look for the years and footer
    footerline = -1
    yearline = -1
    years = []
    yearpat = re.compile("20[12][0-9]")
    for ctr in range(len(lines)):
        if footerline == -1 and values["footer"] in lines[ctr]:
            # account for the extra comment line after the footer
            footerline = ctr + 1
            break
        if yearline == -1:
            years = [int(y) for y in yearpat.findall(lines[ctr])]
            if years:
                yearline = ctr

    if footerline == -1:
        # we found a header but not a footer
        return None

    return dict(
        headerline=headerline,
        footerline=footerline,
        yearline=yearline,
        years=years,
        coprBlock=lines[headerline:footerline + 1],
    )


def createCopyright(values):
    values["yeartext"] = ", ".join([str(y) for y in values["years"]])
    block = copyright_template.substitute(values)
    return block.splitlines()


def different(sa1, sa2):
    " compares two string arrays and returns True if they differ at all "
    if len(sa1) != len(sa2):
        return True
    for i in range(len(sa1)):
        if sa1[i] != sa2[i]:
            return True

    return False


def maybeUpdateFile(filename, ext, values, readonly):
    """
    Checks if a file has the header and the footer. If not, inserts
    a new copyright block and rewrites the file.

    If so, reads the years and checks to see if it includes the current year.
    If it doesn't, it updates the copyright block and rewrites the file.

    If it does, then it tests if the full copyright block is identical
    to the current one. If they're the same, it does not update the file, otherwise
    it does.

    It returns True if the file was updated, False otherwise.
    """
    coprInfo = None
    with open(filename) as f:
        lines = f.read().splitlines()
        if len(lines) == 0:
            return False
        coprInfo = findCopyrightInfo(lines)
        values["years"] = [values["year"]]

    update = False
    if coprInfo is None:
        update = True
    else:
        years = coprInfo["years"]
        if values["year"] not in years:
            years.append(values["year"])
            values["years"] = years
            update = True

    newBlock = createCopyright(values)
    if coprInfo is not None and different(coprInfo["coprBlock"], newBlock):
        update = True

    if update and not readonly:
        with open(filename, "w") as f:
            if coprInfo is None:
                if lines[0].startswith("#!"):
                    lines = lines[:1] + newBlock + lines[1:]
                else:
                    lines = newBlock + lines
            else:
                lines = (
                    lines[: coprInfo["headerline"]]
                    + newBlock
                    + lines[coprInfo["footerline"] :]
                )

            f.write("\n".join(lines))

    return update


def updateFiles(files, ext, values, readonly):
    changed = []
    for file in files:
        if ext:
            if ext.startswith("."):
                ext = ext[1:]
            lang = comments[ext]
        else:
            _, ext = os.path.splitext(file)
            if ext.startswith("."):
                ext = ext[1:]
            if ext in comments:
                lang = comments[ext]

        values.update(lang)
        updated = maybeUpdateFile(file, ext, values, readonly)
        if updated:
            changed.append(file)
    return changed


if __name__ == "__main__":
    parser = setupArgs()
    args = parser.parse_args()

    for x in args.exts:
        args.files.append(f"./**/*.{x}")

    files = generateFileList(args.files)

    changedFiles = updateFiles(files, args.lang, values, args.check)
    if args.check:
        if len(changedFiles) != 0:
            if args.verbose:
                print("\n".join(changedFiles))
            exit(2)
    else:
        if args.verbose:
            print("\n".join(changedFiles))
