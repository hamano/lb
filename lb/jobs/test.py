import click

from .. import cli
from ..bundle import ThreadBundle
from . import common_option
from . import init_option
from base import BaseJob

@cli.command()
@common_option
def test(ctx, n, c, v, d, w, url):
    if v > 0:
        print('Execute: test')
    cfg = init_option(ctx, n, c, v, d, w, url)
    bundle = ThreadBundle(cfg, BaseJob)
    bundle.run()
