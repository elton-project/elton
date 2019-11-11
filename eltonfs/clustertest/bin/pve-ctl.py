#!/usr/bin/env python3
import os
import pathlib
import time
import typing
import subprocess
import tempfile
import re
import logging
import contextlib
import proxmoxer

# Gateway address
GATEWAY = '192.168.189.1'
# IP address used on the setup node.
SETUP_NODE = '192.168.189.149'
# Netmask used on the setup node.
SETUP_NODE_NETMASK = 24
# Path to bash script file.
SETUP_SCRIPT_FILE = 'eltonfs/clustertest/node-setup.sh'
# Memory size allocated to the setup node (in megabytes).
MAX_MEMORY_SIZE = 8192
MEMORY_SIZE = 4096
# Name of main storage.
STORAGE = 'ssd'
# Storage options.
STORAGE_OPT = 'discard=on,ssd=on,cache=unsafe'
# Additional disk size allocated to the setup node (in megabytes).
ADDITIONAL_DISK_SIZE = '+20G'
# Number of CPU cores allocated to the setup node.
VCPUS = 8
# URL to latest cloud image of the Ubuntu 19.04.
UBUNTU_IMAGE_URL = 'https://cloud-images.ubuntu.com/disco/current/disco-server-cloudimg-amd64.img'
# Path to temporary file location.
UBUNTU_IMAGE_PATH = '/var/tmp/ubuntu-19.04.img'
#
SHARED_IMAGE = '/mnt/pve/nas/vm_image.qcow2'

p = proxmoxer.ProxmoxAPI(
    'elton-pve.internal.t-lab.cs.teu.ac.jp',
    user=os.environ.get('PROXMOX_USER'),
    password=os.environ.get('PROXMOX_PW'),
    verify_ssl=False,
)


# fn()が真と評価される値を返すまで、interval間隔で再試行する。
def wait_for(fn: typing.Callable, interval=1):
    while True:
        result = fn()
        if result:
            return result
        time.sleep(interval)


class RemoteCommand(typing.NamedTuple):
    ip: str
    cmd: typing.List[str]

    def execute(self):
        with tempfile.TemporaryFile() as stdin:
            stdin.write(self._input_bytes)
            stdin.seek(0, os.SEEK_SET)
            proc = subprocess.run(self._args, stdin=stdin, capture_output=True)
            proc.check_returncode()
            return proc.stdout

    @property
    def _args(self) -> typing.List[str]:
        return ['ssh', f'root@{self.ip}', '--', 'xargs', '-0', 'env']

    @property
    def _input_bytes(self) -> bytes:
        return b'\0'.join(bytes(c, 'utf-8') for c in self.cmd)


class TaskFailed(Exception): pass


class Task(typing.NamedTuple):
    node: str
    upid: str

    def wait(self):
        while self.is_running():
            time.sleep(1)
        if not self.is_ok():
            raise TaskFailed(self)

    def log(self) -> str:
        lines = []
        n = 0
        limit = 2000
        while True:
            start = n
            ls = self._client().log.get(
                start=start,
                limit=limit,
            )
            for i in range(start, start + limit):
                lines.append(ls[str(i)])
            if len(ls) < limit:
                break
            n += limit
        return '\n'.join(lines)

    def is_running(self) -> bool:
        return self._client().status.get()['status'] == 'running'

    def is_ok(self) -> bool:
        return self._client().status.get()['exitstatus'] == 'OK'

    def _client(self):
        return p.nodes(self.node).tasks(self.upid)


class Node(typing.NamedTuple):
    name: str

    @staticmethod
    def list() -> typing.Generator['Node', None, None]:
        for node in p.nodes.get():
            yield Node(name=node['node'])

    @property
    def ip(self):
        nodes = p.cluster.status.get()
        for n in nodes:
            if n['name'] == self.name:
                return n['ip']
        raise ValueError('not found node')


class Pool(typing.NamedTuple):
    name: str

    def list(self) -> typing.Generator['VM', None, None]:
        for item in p.pools(self.name).get()['members']:
            if item['type'] != 'qemu':
                continue
            yield VM(node=item['node'], vmid=item['vmid'])


class VM(typing.NamedTuple):
    node: str
    vmid: int

    @staticmethod
    def list() -> typing.Generator['VM', None, None]:
        for node in p.nodes.get():
            for vm in p.nodes(node['node']).qemu.get():
                yield VM(node['node'], int(vm['vmid']))

    @staticmethod
    def remove_vms() -> typing.List[Task]:
        tasks = []
        for vm in VM.list():
            print(vm)
            if vm.is_protected or vm.is_template:
                continue
            tasks.append(vm.remove())
        return tasks

    @staticmethod
    def remove_all() -> typing.List[Task]:
        tasks = []
        for vm in VM.list():
            print(vm)
            if vm.is_protected:
                continue
            tasks.append(vm.remove())
        return tasks

    def remove(self) -> Task:
        return Task(node=self.node, upid=self._client.delete())

    def clone(self, newid: int, **kwargs) -> Task:
        return Task(node=self.node, upid=self._client.clone.post(newid=newid, **kwargs))

    def resize(self, disk: str, size: str):
        self._client.resize.put(disk=disk, size=size)

    def start(self) -> Task:
        return Task(node=self.node, upid=self._client.status.start.post())

    def stop(self) -> Task:
        return Task(node=self.node, upid=self._client.status.stop.post())

    def migrate(self, to: Node) -> Task:
        return Task(node=self.node, upid=self._client.migrate.post(target=to.name))

    def is_running(self) -> bool:
        return self._client.status.current.get()['status'] == 'running'

    def is_stopped(self) -> bool:
        return self._client.status.current.get()['status'] == 'stopped'

    def is_exists(self) -> bool:
        try:
            self.is_running()
            return True
        except proxmoxer.core.ResourceException:
            return False

    def set_template(self):
        self._client.template.post()

    @staticmethod
    def next_id() -> int:
        return int(p.cluster.nextid.get())

    @property
    def is_protected(self) -> bool:
        return bool(int(self.config.get('protection', '0')))

    @property
    def is_template(self) -> bool:
        return bool(int(self.config.get('template', '0')))

    @property
    def config(self) -> dict:
        return self._client.config.get()

    @config.setter
    def config(self, diff: dict):
        # Use synchronous API.
        self._client.config.put(**diff)

    @property
    def ip(self):
        s = self.config['ipconfig0']
        for e in s.split(','):
            match = re.fullmatch('ip=(?P<ip>.+)/[0-9]+', e)
            if match:
                return match.group('ip')
        raise ValueError(f'IP address is not allocated for {self}')

    @property
    def unsafe_ssh_command(self):
        return ['ssh', '-T', '-o', 'UserKnownHostsFile=/dev/null', '-o', 'StrictHostKeyChecking=no', f'root@{self.ip}']

    @property
    def _client(self):
        return p.nodes(self.node).qemu(self.vmid)


class TemplateBuilder(typing.NamedTuple):
    pool: Pool
    base: VM
    script_name: pathlib.Path
    template_name: str
    output: VM

    def remove(self):
        if self.output.is_exists():
            if not self.output.is_stopped():
                self.output.stop().wait()
            self.output.config = {'protection': 0}
            self.output.remove().wait()

    def build(self):
        self.base.clone(
            newid=self.output.vmid,
            name=self.template_name,
            pool=self.pool.name,
            full=1,
        ).wait()
        self.output.config = {'protection': 0}

        self._set_storage(self.output)
        self.output.config = {
            'ipconfig0': f'gw={GATEWAY},ip={SETUP_NODE}/{SETUP_NODE_NETMASK}',
            'agent': 'enabled=0',
            'memory': MAX_MEMORY_SIZE,
            'balloon': MEMORY_SIZE,
            'sockets': 1,
            'cores': VCPUS,
            'vcpus': VCPUS,
        }
        self.output.start().wait()
        wait_for(lambda: self._is_ready(self.output))
        self._run_script(self.output, self.script_name)
        wait_for(lambda: not self.output.is_running())

    def _set_storage(self, vm: VM):
        ip = Node(name=vm.node).ip
        # Download latest image.
        RemoteCommand(ip=ip, cmd=['rm', '-f', UBUNTU_IMAGE_PATH, ]).execute()
        RemoteCommand(ip=ip, cmd=['wget', UBUNTU_IMAGE_URL, '-O', UBUNTU_IMAGE_PATH]).execute()

        # Set dist to the VM.
        RemoteCommand(ip=ip, cmd=['qm', 'importdisk', str(vm.vmid), UBUNTU_IMAGE_PATH, STORAGE, '--format', 'qcow2']
                      ).execute()
        vm.config = {
            'scsi0': f'{STORAGE}:{vm.vmid}/vm-{vm.vmid}-disk-0.qcow2,{STORAGE_OPT}',
        }
        # Increase disk size.
        vm.resize('scsi0', ADDITIONAL_DISK_SIZE)

    def _is_ready(self, vm: VM) -> bool:
        try:
            proc = subprocess.run(
                [*vm.unsafe_ssh_command, 'systemctl', 'status', 'cloud-final'],
                stdin=None,
                capture_output=True,
            )
            proc.check_returncode()
            return b'Active: active (exited) since' in proc.stdout
        except subprocess.CalledProcessError:
            return False

    def _run_script(self, vm: VM, script_path: pathlib.Path):
        with script_path.open('rb') as stdin:
            subprocess.run(
                [*vm.unsafe_ssh_command, 'bash'],
                stdin=stdin,
            ).check_returncode()
        try:
            subprocess.run(
                [*vm.unsafe_ssh_command, 'poweroff'],
                stdin=None,
            ).check_returncode()
        except subprocess.CalledProcessError:
            logging.warning("ignore error when executing poweroff command")


class TemplateDistributor(typing.NamedTuple):
    pool: Pool
    # VM Template
    template: VM
    # VM for disk image template.
    disk_image: VM

    def remove_all(self):
        while True:
            tasks = []
            for vm in self.pool.list():
                match = False
                for line in vm.config.get('description', '').splitlines():
                    if line.strip() == self._vm_property_in_description:
                        match = True
                        break
                if match:
                    if vm.is_stopped():
                        vm.config = {'protection': 0}
                        tasks.append(vm.remove())
                    else:
                        tasks.append(vm.stop())

            if len(tasks) == 0:
                # All VMs are deleted.
                return

            for t in tasks:
                with contextlib.suppress(TaskFailed):
                    t.wait()
            # Some tasks may be failed.  Should retry until tasks is empty.

    def distribute(self):
        target_name = self.disk_image.config['name']
        target_desc = self.disk_image.config.get('description', '')

        # Copy target disk image to shared storage.
        ip = Node(name=self.disk_image.node).ip
        RemoteCommand(ip=ip, cmd=['rm', SHARED_IMAGE]).execute()
        RemoteCommand(ip=ip, cmd=['cp', self._disk_path, SHARED_IMAGE]).execute()

        vms = []
        for node in Node.list():
            vm = VM(node=node.name, vmid=VM.next_id())
            cloned_vm = VM(node=self.template.node, vmid=vm.vmid)
            self.template.clone(vm.vmid,
                                name=f'{target_name}-{node.name}',
                                description=f'{target_desc}\n{self._vm_property_in_description}',
                                pool=self.pool.name,
                                full=1,
                                storage=STORAGE).wait()

            # Unset cloud-init drive to prevent fail of VM migration with local storage.
            cloned_vm.config = {'ide0': 'none'}
            if vm.node != cloned_vm.node:
                cloned_vm.migrate(node).wait()
            RemoteCommand(ip=node.ip,
                          cmd=['qm', 'importdisk', str(vm.vmid), SHARED_IMAGE, STORAGE, '--format', 'qcow2']
                          ).execute()
            vm.config = {
                'scsi0': f'{STORAGE}:{vm.vmid}/vm-{vm.vmid}-disk-0.qcow2,{STORAGE_OPT}',
                'ide0': f'{STORAGE}:cloudinit',
                'protection': '0',
            }
            vms.append(vm)

        for vm in vms:
            vm.set_template()
            # templateは非同期APIだが、完了したことを検出できないため、待機処理はしない。
            # 作成したVMを即座に利用すると失敗する可能性があるので、後続の処理ではsleepするべき。

    @property
    def _disk_path(self) -> str:
        vmid = self.disk_image.vmid
        return f'/mnt/ssd/images/{vmid}/vm-{vmid}-disk-0.qcow2'

    @property
    def _vm_property_in_description(self):
        return f'clustertest.based_on={self.template.vmid}/{self.disk_image.vmid}'


base = VM('elton-pve1', 9000)
out = VM('elton-pve1', 9100)
pool = Pool('clustertest')
builder = TemplateBuilder(base=base, script_name=pathlib.Path(SETUP_SCRIPT_FILE),
                          output=out, pool=pool, template_name='template-ubuntu-19.04-ltp')
dist = TemplateDistributor(template=base, disk_image=out, pool=pool)

dist.remove_all()
builder.remove()
builder.build()
dist.distribute()
