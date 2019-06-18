#!/usr/bin/env python3
from jinja2 import Environment, FileSystemLoader
env = Environment(loader=FileSystemLoader('./', encoding='utf8'))
tpl = env.get_template('ltp-fail.tmpl.yaml')

names = '''
execveat03
fcntl24
fcntl24_64
fcntl25
fcntl25_64
fcntl26
fcntl26_64
fcntl33
fcntl33_64
inotify07
inotify08
keyctl02
'''.strip().splitlines()

tpl_data = {
	'rules': [
		{
			'vmname': name.replace('_', '-'),
			'cmd': name,
		}
		for name in names
	],
	'warn': '''
##################  DO NOT EDIT  ##################
#  This file generated by ltp-fail.generate.py.   #
#  You MUST NOT edit this file.                   #
###################################################
''',
}

with open('ltp-fail.yaml', 'w') as f:
	data = tpl.render(tpl_data)
	f.write(data)
