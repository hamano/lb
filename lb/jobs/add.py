import time
import ldap
import base
import traceback
import click
import libuuid

from .. import cli
from ..bundle import ThreadBundle
from . import common_option
from . import init_option
from base import BaseJob

template_dn = 'cn=%s,dc=example,dc=com'
template_entry = [
    ('objectClass', ['person']),
    ('userPassword', ['secret']),
]

class AddJob(BaseJob):
    def __init__(self, cfg, id):
        super(AddJob, self).__init__(cfg, id)
        rc = self.ldap.simple_bind_s(cfg['bind_dn'], cfg['bind_pw'])
        self.i = 0

    def request(self):
        cn = str(libuuid.uuid1())
#        cn = '%s_%s' % (self.tid, self.i)
        sn = str(self.tid)
        dn = template_dn % (cn)
        entry = list(template_entry)
        entry.append(('cn', [cn]))
        entry.append(('sn', [sn]))
        self.i += 1
        try:
            (rc, msg) = self.ldap.add_s(dn, entry)
            if rc == 105:
                return True
            else:
                print "error: " + str(rc)
                return False
        except ldap.LDAPError as le:
            print le.message
        except Exception as e:
            print traceback.format_exc()
        return False

@cli.command()
@common_option
def add(ctx, n, c, v, d, w, url):
    cfg = init_option(ctx, n, c, v, d, w, url)
    bundle = ThreadBundle(cfg, AddJob)
    bundle.run()
