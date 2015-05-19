import sys
import threading
import time
import ldap
from ..bundle import events

class BaseJob(threading.Thread):
    def __init__(self, cfg, id):
        super(BaseJob, self).__init__()
        self.cfg = cfg
        self.tid = id
        self.result = {}
        self.result['name'] = 'thread%d' % id
        self.result['count'] = 0
        self.result['success'] = 0
        self.result['error'] = 0
        self.requests = cfg['requests_per_thread']
        self.ldap = ldap.initialize(cfg['url'])

    def run(self):
        startTime = time.time()
        for i in xrange(self.cfg['requests_per_thread']):
            if events['abort'].is_set():
                return
            if events['timeout'].is_set():
                break
            if self.request():
                self.result['success'] += 1
            self.result['count'] += 1
        endTime = time.time()
        elapsedTime = endTime - startTime
        self.result['rps'] = float(self.result['count']) / elapsedTime
        self.result['elapsedTime'] = elapsedTime
        self.result['complete'] = True

    def request(self):
        # dummy
        time.sleep(0.1)
        return True

def search(l, result):
#    res = l.search_s(BASE_DN, ldap.SCOPE_BASE, '(objectClass=*)', None)
    res = l.search_s(BASE_DN, ldap.SCOPE_SUBTREE, '(objectClass=*)', None)
    result['count'] += 1
    if len(res) == 1:
        result['success'] += 1
    else:
        result['error'] += 1

def bind(l, result):
    res = l.simple_bind_s(BIND_DN, BIND_PW)
    result['count'] += 1
    if res[0] == 97:
        result['success'] += 1
    else:
        result['error'] += 1

