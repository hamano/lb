import sys
import time
import signal
import threading

events = {}
events['abort'] = threading.Event()
events['timeout'] = threading.Event()

def signal_handler(signum, frame):
    if signum == signal.SIGINT:
        events['abort'].set()
    elif signum == signal.SIGALRM:
        events['timeout'].set()

class ThreadBundle():
    def __init__(self, cfg, ThreadClass):
        self.cfg = cfg
        self.threads = []
        for i in xrange(self.cfg['concurrency']):
            t = ThreadClass(self.cfg, i)
            self.threads.append(t)

    def run(self):
        signal.signal(signal.SIGINT, signal_handler)
        signal.signal(signal.SIGALRM, signal_handler)
        self.start()
        self.join()
        if events['abort'].is_set():
            if self.cfg['verbose'] > 0:
                print('Aborted.')
            return
        self.report()

    def start(self):
        for t in self.threads:
            if self.cfg['verbose'] > 0:
                print('Starting Thread: %s' % (t.tid))
            t.start()

    def join(self):
        for t in self.threads:
            while t.isAlive():
                time.sleep(1)
            t.join()

    def report(self):
        count = 0
        success = 0
        totalTime = 0
        for t in self.threads:
            result = t.result
            if 'complete' not in result:
                continue
            count += result['count'] # TODO: fix
            success += result['success']
            totalTime += result['elapsedTime']
            if self.cfg['verbose'] > 0:
                print 'Thread %s: %d req, %f req/sec, %f sec' % \
                    (t.tid, \
                     result['count'], \
                     result['rps'], \
                     result['elapsedTime'])
        avgTime = totalTime / self.cfg['concurrency']
        rpq = float(count) / avgTime
        print('Concurrency Level: %d' % (self.cfg['concurrency']))
        print('Total requests: %d' % (count))
        print('Success requests: %d' % (success))
        print('Success rate: %d%%' % (float(success) / count * 100))
        print('Requests per second: %.3f req/sec' % (rpq))
