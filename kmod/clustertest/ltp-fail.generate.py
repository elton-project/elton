#!/usr/bin/env python3
import glob
import os
from jinja2 import Environment, FileSystemLoader
env = Environment(loader=FileSystemLoader('./', encoding='utf8'))
tpl = env.get_template('ltp-fail.tmpl.yaml')

warn_message = '''
##################  DO NOT EDIT  ##################
#  This file generated by ltp-fail.generate.py.   #
#  You MUST NOT edit this file.                   #
###################################################
'''


# このテストは、カーネルがハングしてしまい、clustertestが完走しない。
# execveat03

# FAILしたテストケース
names = '''
fcntl24
fcntl24_64
fcntl25
fcntl25_64
fcntl26
fcntl26_64
fcntl33
fcntl33_64
'''.strip().splitlines()

# CONFで失敗したテストケース
names += '''
bdflush01
cacheflush01
chown01_16
chown02_16
chown03_16
chown04_16
chown05_16
eventfd01
fallocate01
fallocate03
fchown01_16
fchown02_16
fchown03_16
fchown04_16
fchown05_16
fcntl06
fcntl06_64
get_mempolicy01
getegid01_16
getegid02_16
geteuid01_16
geteuid02_16
getgid01_16
getgid03_16
getgroups01_16
getgroups03_16
getresgid01_16
getresgid02_16
getresgid03_16
getresuid01_16
getresuid02_16
getresuid03_16
getrusage04
getuid01_16
getuid03_16
getxattr05
io_cancel01
io_destroy01
io_getevents01
io_setup01
io_submit01
lchown01_16
lchown02_16
lchown03_16
'''.strip().splitlines()


# Remove generated files.
files = glob.glob('ltp-fail.generated.*.yaml')
for f in files:
    os.unlink(f)

# Generate files.
for name in names:
    tpl_data = {
        'r': {
            'vmname': name.replace('_', '-'),
            'cmd': name,
        },
        'warn': warn_message,
    }
    with open(f'ltp-fail.generated.{name}.yaml', 'w') as f:
        data = tpl.render(tpl_data)
        f.write(data)
