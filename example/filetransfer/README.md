# TCQ File Transfer Example

This example adds a small QUIC file transfer tool for LAN and Tailscale use.

## Server on Ubuntu

```bash
export TCQ_TRANSFER_TOKEN="change-me"
go run ./example/filetransfer/server -listen 0.0.0.0:4242 -root /srv/tcq-transfer
```

Open UDP port `4242` on the Ubuntu firewall. The same listener is reachable on
the LAN address `192.168.1.5:4242` and the Tailscale address
`100.91.211.76:4242` when those interfaces are configured on the server.

## Client on Windows

```powershell
$env:TCQ_TRANSFER_TOKEN="change-me"
go run ./example/filetransfer/client -network lan upload C:\Users\Alice\Videos uploads
go run ./example/filetransfer/client -network tailscale download uploads\Videos D:\Restore
```

You can override the target directly:

```powershell
go run ./example/filetransfer/client -addr 192.168.1.5:4242 upload C:\Temp\report.pdf .
go run ./example/filetransfer/client -addr 100.91.211.76 download uploads D:\Backup
```

The client uploads files and folders as tar streams. Downloads are also returned
as tar streams and extracted into the destination folder.
