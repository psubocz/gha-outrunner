# How to Build a Custom Windows VM Image

Build a Windows qcow2 golden image for use with the libvirt provisioner.

## Recommended: Packer with rgl base

The fastest path is to start from [rgl/windows-vagrant](https://github.com/rgl/windows-vagrant), which produces a clean Windows 11 qcow2 with VirtIO drivers, QEMU Guest Agent, and SSH already configured. Then layer your tools on top with Packer.

### Prerequisites

- Linux server with KVM and libvirt
- [Packer](https://developer.hashicorp.com/packer/downloads) installed
- OVMF firmware: `sudo apt install ovmf` (Ubuntu/Debian) or `sudo dnf install edk2-ovmf` (Fedora)

### 1. Build the rgl base image

```bash
git clone https://github.com/rgl/windows-vagrant.git
cd windows-vagrant
```

Follow the rgl README to build the libvirt/QEMU variant. This takes 2-4 hours (Windows install + updates). The output is a qcow2 with Windows 11, VirtIO drivers, guest agent, and SSH.

### 2. Layer your tools with Packer

Create `runner-image.pkr.hcl`:

```hcl
packer {
  required_plugins {
    qemu = {
      source  = "github.com/hashicorp/qemu"
      version = "~> 1"
    }
  }
}

variable "base_image_path" {
  type        = string
  description = "Path to the rgl base qcow2 image"
}

source "qemu" "runner" {
  disk_image       = true
  iso_url          = var.base_image_path
  iso_checksum     = "none"
  skip_resize_disk = true

  accelerator  = "kvm"
  machine_type = "q35"
  cpu_model    = "host"
  cpus         = 4
  memory       = 8192
  headless     = true

  disk_interface = "virtio-scsi"
  format         = "qcow2"
  net_device     = "virtio-net"

  efi_boot          = true
  efi_firmware_code = "/usr/share/OVMF/OVMF_CODE_4M.fd"
  efi_firmware_vars = "/usr/share/OVMF/OVMF_VARS_4M.fd"

  communicator = "ssh"
  ssh_username = "vagrant"
  ssh_password = "vagrant"
  ssh_timeout  = "30m"

  boot_wait    = "60s"
  boot_command = []

  shutdown_command = "shutdown /s /t 10 /f /d p:4:1"
  shutdown_timeout = "15m"

  output_directory = "/var/lib/libvirt/images/ci-runners/packer-output"
  vm_name          = "windows-runner.qcow2"
}

build {
  sources = ["source.qemu.runner"]

  provisioner "powershell" {
    inline = [
      "mkdir C:\\actions-runner; cd C:\\actions-runner",
      "Invoke-WebRequest -Uri https://github.com/actions/runner/releases/download/v2.333.1/actions-runner-win-x64-2.333.1.zip -OutFile runner.zip",
      "Expand-Archive runner.zip -DestinationPath .",
      "Remove-Item runner.zip"
    ]
  }

  provisioner "powershell" {
    inline = [
      "# Add your tools here, e.g.:",
      "# choco install -y git cmake",
      "# choco install -y visualstudio2022buildtools",
      "# choco install -y visualstudio2022-workload-vctools"
    ]
  }

  provisioner "powershell" {
    inline = [
      "Remove-Item -Recurse -Force $env:TEMP\\* -ErrorAction SilentlyContinue",
      "Remove-Item -Recurse -Force C:\\Windows\\Temp\\* -ErrorAction SilentlyContinue",
      "Optimize-Volume -DriveLetter C -Defrag"
    ]
  }
}
```

### 3. Build

```bash
packer init runner-image.pkr.hcl
packer build -var "base_image_path=/path/to/rgl-base.qcow2" runner-image.pkr.hcl
```

### 4. Move the image and use it

```bash
mv /var/lib/libvirt/images/ci-runners/packer-output/windows-runner.qcow2 \
   /var/lib/libvirt/images/ci-runners/windows-runner.qcow2
rm -rf /var/lib/libvirt/images/ci-runners/packer-output
```

```yaml
runners:
  windows:
    labels: [self-hosted, windows]
    libvirt:
      path: /var/lib/libvirt/images/ci-runners/windows-runner.qcow2
      runner_cmd: 'C:\actions-runner\run.cmd'
      cpus: 4
      memory: 8192
```

## Layered images

For complex setups, build images in layers so you don't rebuild everything when one tool changes:

```
Layer 0: rgl base (Windows 11 + VirtIO + guest agent + SSH)
    |
Layer 1: Build tools (VS, CMake, Git, runner) - rebuild on tool updates
    |
Layer 2: Project-specific (language runtimes, caches) - rebuild frequently
```

Each layer is a Packer build with `disk_image = true` on top of the previous layer's output.

## Alternative: manual build

If you prefer not to use Packer, you can build an image manually via VNC:

```bash
qemu-img create -f qcow2 /var/lib/libvirt/images/ci-runners/windows.qcow2 60G

virt-install \
  --name win-setup \
  --ram 8192 --vcpus 4 \
  --os-variant win11 \
  --disk path=/var/lib/libvirt/images/ci-runners/windows.qcow2,bus=scsi \
  --cdrom /path/to/windows.iso \
  --disk path=/path/to/virtio-win.iso,device=cdrom \
  --network network=default,model=virtio \
  --controller type=scsi,model=virtio-scsi \
  --graphics vnc,listen=0.0.0.0 \
  --boot firmware=efi
```

Connect via VNC, install Windows, load VirtIO drivers, install the QEMU Guest Agent and Actions runner manually. See the [image requirements](../reference/image-requirements.md) for what needs to be in the image.

## Tips

- **Disk size:** 60 GB is usually enough. qcow2 only uses space for written data.
- **Disable Windows Update** in the golden image to avoid update noise during CI jobs.
- **Test the guest agent** before finalizing: `virsh qemu-agent-command win-setup '{"execute":"guest-ping"}'`
- **Compress the final image:** `qemu-img convert -O qcow2 -c input.qcow2 output.qcow2`
