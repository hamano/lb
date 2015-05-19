#!/usr/bin/env python2
# -*- coding: utf-8 -*-

import sys
import click

@click.group()
@click.pass_context
def cli(ctx):
    ctx.obj = {}

import jobs

if __name__ == '__main__':
    cli()
