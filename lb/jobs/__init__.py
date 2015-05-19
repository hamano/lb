import click
import math

# decorator chain for click common option
def common_option(f):
    f = click.option('-n',
                     default=1,
                     help='Number of requests to perform')(f)
    f = click.option('-c',
                     help='Number of multiple requests to make',
                     default=1)(f)
    f = click.option('-v',
                     default=0,
                     help='How much troubleshooting info to print')(f)
    f = click.option('-D',
                     help='',
                     default='cn=Manager,dc=example,dc=com')(f)
    f = click.option('-w',
                     help='')(f)

    f = click.argument('URL')(f)

    f = click.pass_context(f)

    return f

def init_option(ctx, n, c, v, d, w, url):
    cfg = {}
    cfg['url'] = url
    cfg['requests'] = n
    cfg['concurrency'] = c
    cfg['verbose'] = v
    cfg['bind_dn'] = d
    cfg['bind_pw'] = w
    cfg['requests_per_thread'] = int(math.ceil(float(n) / c))
    return cfg

import base
import test
import add
