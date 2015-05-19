#!/usr/bin/env python

"""Setup script for lb."""

import setuptools

import os
if os.path.exists('README.md'):
    README = open('README.md').read()
else:
    README = ""  # a placeholder, readme is generated on release

CHANGES = ''

setuptools.setup(
    name='lb',
    version=0.1,

    description="lb is LDAP benchmarking tool like an Apache Bench",
    url='https://github.com/hamano/lb',
    author='HAMANO Tsukasa',
    author_email='hamano@osstech.co.jp',

    packages=setuptools.find_packages(),

    entry_points={'console_scripts': []},

    long_description=(README + '\n' + CHANGES),
    license='MIT',
    classifiers=[
        'Development Status :: 1 - Planning',
        'Operating System :: OS Independent',
    ],
    scripts=["bin/lb"],
    install_requires=open('requirements.txt').readlines(),
)
