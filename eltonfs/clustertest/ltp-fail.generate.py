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


# このテストは、ファイルシステム以外の要因により失敗している。
# 他のファイルシステム (ext4とtmpfs) でも失敗することを確認した。
# msgstress03

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
execveat03
'''.strip().splitlines()

# CONFで失敗したテストケース
names += '''
bdflush01
cacheflush01
fallocate01
fallocate03
fcntl06
fcntl06_64
getrusage04
lseek11
migrate_pages02
migrate_pages03
modify_ldt01
modify_ldt02
modify_ldt03
move_pages01
move_pages02
move_pages03
move_pages04
move_pages05
move_pages06
move_pages07
move_pages08
move_pages09
move_pages10
move_pages11
move_pages12
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
