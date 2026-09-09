//	Project: TCQ Network Protocol (Thread Controlled QUIC)
//	Author: Trần Nguyên Hiền (c)
//	Major: Electronic And Communication Engineering
//	Email: trannguyenhien29085@gmail.com
//	Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
package main

import (
	"archive/tar"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	quic "github.com/NguyenHien-8/tcq-network-protocol"
)

const (
	protocolName = "tcq-filetransfer-v1"
	protocolVer  = 1
	maxJSONFrame = 1 << 20 // 1 MiB for control messages
)

type request struct {
	Version    int    `json:"version"`
	Token      string `json:"token,omitempty"`
	Op         string `json:"op"`
	RemotePath string `json:"remote_path"`
	SourceKind string `json:"source_kind,omitempty"`
	SourceName string `json:"source_name,omitempty"`
	Overwrite  bool   `json:"overwrite"`
}

type response struct {
	OK      bool        `json:"ok"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Name    string      `json:"name,omitempty"`
	Entries []listEntry `json:"entries,omitempty"`
}

type listEntry struct {
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type progressWriter struct {
	W         io.Writer
	Label     string
	Total     int64
	Written   int64
	LastPrint time.Time
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.W.Write(b)
	p.Written += int64(n)
	now := time.Now()
	if p.LastPrint.IsZero() || now.Sub(p.LastPrint) >= 2*time.Second {
		p.LastPrint = now
		if p.Total > 0 {
			log.Printf("%s: sent %.2f MiB / %.2f MiB", p.Label, float64(p.Written)/(1<<20), float64(p.Total)/(1<<20))
		} else {
			log.Printf("%s: sent %.2f MiB", p.Label, float64(p.Written)/(1<<20))
		}
	}
	return n, err
}

func main() {
	server := flag.String("server", "192.168.1.5:4242", "server UDP address; use 192.168.1.5:4242 on LAN or 100.91.211.76:4242 via Tailscale")
	token := flag.String("token", "", "shared auth token configured on the server")
	overwrite := flag.Bool("overwrite", true, "overwrite existing files on upload/download")
	insecure := flag.Bool("insecure", true, "skip TLS certificate verification for the server self-signed certificate")
	timeout := flag.Duration("timeout", 0, "operation timeout, for example 30s or 10m; 0 disables it")
	bind := flag.String("bind", "auto", "local UDP bind address. auto binds to the local Tailscale IPv4 when -server is 100.64.0.0/10; use 0.0.0.0:0 to disable")
	transport := flag.String("transport", "auto", "transport: auto, quic, or tcp. auto tries QUIC first, then TCP fallback for Tailscale/UDP-blocked networks")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "TCQ file client over QUIC with TCP fallback\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  client.exe [flags] upload <local-file-or-folder> [remote-path]\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  client.exe [flags] download <remote-file-or-folder> [local-path]\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  client.exe [flags] list [remote-path]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  client.exe -server 192.168.1.5:4242 -token secret upload C:\\Data backup\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  client.exe -server 100.91.211.76:4242 -token secret download backup C:\\Download\\backup\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  client.exe -server 100.91.211.76:4242 -token secret list backup\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: *insecure,
		NextProtos:         []string{protocolName},
	}
	quicConf := &quic.Config{
		KeepAlivePeriod:                15 * time.Second,
		MaxIdleTimeout:                 5 * time.Minute,
		InitialStreamReceiveWindow:     4 << 20,
		MaxStreamReceiveWindow:         64 << 20,
		InitialConnectionReceiveWindow: 8 << 20,
		MaxConnectionReceiveWindow:     256 << 20,
	}

	openStream, closeSession, usedTransport, err := connectTCQ(ctx, *server, *bind, *transport, tlsConf, quicConf)
	if err != nil {
		log.Fatalf("connect %s: %v", *server, err)
	}
	defer closeSession()
	log.Printf("connected using %s", usedTransport)

	switch args[0] {
	case "upload", "put":
		if len(args) < 2 || len(args) > 3 {
			log.Fatalf("usage: upload <local-file-or-folder> [remote-path]")
		}
		remote := ""
		if len(args) == 3 {
			remote = args[2]
		}
		if err := upload(ctx, openStream, *token, args[1], remote, *overwrite); err != nil {
			log.Fatal(err)
		}
	case "download", "get":
		if len(args) < 2 || len(args) > 3 {
			log.Fatalf("usage: download <remote-file-or-folder> [local-path]")
		}
		local := ""
		if len(args) == 3 {
			local = args[2]
		}
		if err := download(ctx, openStream, *token, args[1], local, *overwrite); err != nil {
			log.Fatal(err)
		}
	case "list", "ls":
		remote := "."
		if len(args) > 2 {
			log.Fatalf("usage: list [remote-path]")
		}
		if len(args) == 2 {
			remote = args[1]
		}
		if err := listRemote(ctx, openStream, *token, remote); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command %q", args[0])
	}
}

type streamOpener func(context.Context) (io.ReadWriteCloser, error)

func connectTCQ(ctx context.Context, server, bind, transport string, tlsConf *tls.Config, quicConf *quic.Config) (streamOpener, func() error, string, error) {
	transport = strings.ToLower(strings.TrimSpace(transport))
	if transport == "" {
		transport = "auto"
	}
	if transport != "auto" && transport != "quic" && transport != "tcp" {
		return nil, nil, "", fmt.Errorf("unsupported -transport %q; use auto, quic, or tcp", transport)
	}

	if transport == "tcp" {
		return tcpStreamOpener(server, bind), func() error { return nil }, "tcp", nil
	}

	quicConn, closePacketConn, err := dialTCQ(ctx, server, bind, tlsConf, quicConf)
	if err == nil {
		closeFn := func() error {
			quicConn.CloseWithError(0, "client done")
			if closePacketConn != nil {
				return closePacketConn()
			}
			return nil
		}
		open := func(ctx context.Context) (io.ReadWriteCloser, error) {
			return quicConn.OpenStreamSync(ctx)
		}
		return open, closeFn, "quic/udp", nil
	}
	if transport == "quic" {
		return nil, nil, "", err
	}

	log.Printf("QUIC/UDP connect failed: %v", err)
	log.Printf("falling back to TCP on %s. This is useful when Tailscale/OS policy blocks UDP application traffic but allows TCP.", server)
	return tcpStreamOpener(server, bind), func() error { return nil }, "tcp-fallback", nil
}

func tcpStreamOpener(server, bind string) streamOpener {
	return func(ctx context.Context) (io.ReadWriteCloser, error) {
		serverTCP, err := net.ResolveTCPAddr("tcp", server)
		if err != nil {
			return nil, err
		}
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		bind = strings.TrimSpace(strings.ToLower(bind))
		if bind == "auto" && isTailscaleIPv4(serverTCP.IP) {
			ip, err := findLocalTailscaleIPv4(serverTCP.IP)
			if err == nil {
				dialer.LocalAddr = &net.TCPAddr{IP: ip, Port: 0}
				log.Printf("TCP fallback binding local socket to Tailscale IP %s", ip)
			}
		} else if bind != "" && bind != "default" && bind != "auto" && bind != "0.0.0.0:0" && bind != "[::]:0" {
			udpAddr, err := net.ResolveUDPAddr("udp", bind)
			if err != nil {
				return nil, fmt.Errorf("invalid -bind %q: %w", bind, err)
			}
			dialer.LocalAddr = &net.TCPAddr{IP: udpAddr.IP, Port: 0}
		}
		conn, err := dialer.DialContext(ctx, "tcp", serverTCP.String())
		if err != nil {
			return nil, err
		}
		log.Printf("TCP fallback socket: %s -> %s", conn.LocalAddr(), conn.RemoteAddr())
		return conn.(io.ReadWriteCloser), nil
	}
}

func closeWrite(s io.ReadWriteCloser) error {
	if cw, ok := s.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return s.Close()
}

func dialTCQ(ctx context.Context, server, bind string, tlsConf *tls.Config, quicConf *quic.Config) (*quic.Conn, func() error, error) {
	serverUDP, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return nil, nil, err
	}

	bind = strings.TrimSpace(strings.ToLower(bind))
	if bind == "" || bind == "default" || bind == "0.0.0.0:0" || bind == "[::]:0" {
		log.Printf("dialing %s with default UDP routing", server)
		conn, err := quic.DialAddr(ctx, server, tlsConf, quicConf)
		return conn, nil, err
	}

	var localUDP *net.UDPAddr
	if bind == "auto" {
		if isTailscaleIPv4(serverUDP.IP) {
			ip, err := findLocalTailscaleIPv4(serverUDP.IP)
			if err != nil {
				return nil, nil, fmt.Errorf("server is a Tailscale IP, but no local Tailscale IPv4 was found; pass -bind 0.0.0.0:0 to use default routing: %w", err)
			}
			localUDP = &net.UDPAddr{IP: ip, Port: 0}
			log.Printf("server %s is in 100.64.0.0/10; binding local UDP socket to Tailscale IP %s", serverUDP.IP, ip)
		} else {
			log.Printf("server %s is not a Tailscale 100.64.0.0/10 IP; using default UDP routing", serverUDP.IP)
			conn, err := quic.DialAddr(ctx, server, tlsConf, quicConf)
			return conn, nil, err
		}
	} else {
		localUDP, err = net.ResolveUDPAddr("udp", bind)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid -bind %q: %w", bind, err)
		}
		log.Printf("binding local UDP socket to %s", localUDP.String())
	}

	udpConn, err := net.ListenUDP("udp", localUDP)
	if err != nil {
		return nil, nil, fmt.Errorf("listen local UDP %s: %w", localUDP, err)
	}
	log.Printf("local UDP socket: %s -> remote %s", udpConn.LocalAddr(), serverUDP)
	conn, err := quic.Dial(ctx, udpConn, serverUDP, tlsConf, quicConf)
	if err != nil {
		udpConn.Close()
		return nil, nil, err
	}
	return conn, udpConn.Close, nil
}

func isTailscaleIPv4(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	// Tailscale IPv4 addresses are in 100.64.0.0/10.
	return ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127
}

func findLocalTailscaleIPv4(exclude net.IP) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	exclude4 := exclude.To4()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ip4 := ip.To4()
			if ip4 == nil || !isTailscaleIPv4(ip4) {
				continue
			}
			if exclude4 != nil && ip4.Equal(exclude4) {
				continue
			}
			return append(net.IP(nil), ip4...), nil
		}
	}
	return nil, errors.New("no local 100.64.0.0/10 address found")
}

func calcRegularFileBytes(root string) (int64, error) {
	var total int64
	info, err := os.Stat(root)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func upload(ctx context.Context, openStream streamOpener, token, localPath, remotePath string, overwrite bool) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	kind := "file"
	if info.IsDir() {
		kind = "dir"
	}
	name := filepath.Base(localPath)
	if remotePath == "" {
		remotePath = name
	}

	totalBytes, err := calcRegularFileBytes(localPath)
	if err != nil {
		return err
	}
	log.Printf("uploading %s (%s) to server path %q", localPath, kind, remotePath)

	stream, err := openStream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	req := request{
		Version:    protocolVer,
		Token:      token,
		Op:         "upload",
		RemotePath: remotePath,
		SourceKind: kind,
		SourceName: name,
		Overwrite:  overwrite,
	}
	if err := writeJSON(stream, req); err != nil {
		return err
	}
	pw := &progressWriter{W: stream, Label: "upload", Total: totalBytes}
	if err := writeSourceAsTar(pw, localPath, info); err != nil {
		return err
	}
	log.Printf("upload stream finished: sent %.2f MiB", float64(pw.Written)/(1<<20))
	if err := closeWrite(stream); err != nil {
		return err
	}

	var resp response
	if err := readJSON(stream, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	fmt.Printf("OK: %s (%s %s)\n", resp.Message, resp.Kind, resp.Name)
	return nil
}

func download(ctx context.Context, openStream streamOpener, token, remotePath, localPath string, overwrite bool) error {
	stream, err := openStream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	req := request{Version: protocolVer, Token: token, Op: "download", RemotePath: remotePath, Overwrite: overwrite}
	if err := writeJSON(stream, req); err != nil {
		return err
	}
	if err := closeWrite(stream); err != nil {
		return err
	}

	var resp response
	if err := readJSON(stream, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	if resp.Kind != "file" && resp.Kind != "dir" {
		return fmt.Errorf("server returned unsupported kind %q", resp.Kind)
	}
	if resp.Name == "" {
		resp.Name = cleanBaseName(remotePath)
	}
	if resp.Name == "" {
		resp.Name = "download"
	}

	if resp.Kind == "file" {
		target := localPath
		if target == "" {
			target = resp.Name
		} else if looksLikeDir(target) {
			target = filepath.Join(target, resp.Name)
		} else if st, err := os.Stat(target); err == nil && st.IsDir() {
			target = filepath.Join(target, resp.Name)
		}
		if err := extractSingleDownloadedFile(stream, target, overwrite); err != nil {
			return err
		}
		fmt.Printf("OK: downloaded file to %s\n", target)
		return nil
	}

	targetDir := localPath
	if targetDir == "" {
		targetDir = resp.Name
	}
	if err := extractDownloadedDir(stream, targetDir, overwrite); err != nil {
		return err
	}
	fmt.Printf("OK: downloaded directory to %s\n", targetDir)
	return nil
}

func listRemote(ctx context.Context, openStream streamOpener, token, remotePath string) error {
	stream, err := openStream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	req := request{Version: protocolVer, Token: token, Op: "list", RemotePath: remotePath}
	if err := writeJSON(stream, req); err != nil {
		return err
	}
	if err := closeWrite(stream); err != nil {
		return err
	}

	var resp response
	if err := readJSON(stream, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	for _, e := range resp.Entries {
		fmt.Printf("%-6s %12d %s %s\n", e.Kind, e.Size, e.ModTime, e.Path)
	}
	return nil
}

func writeSourceAsTar(w io.Writer, localPath string, info os.FileInfo) error {
	tw := tar.NewWriter(w)
	var err error
	if info.IsDir() {
		err = addDirToTar(tw, localPath)
	} else {
		err = addFileToTar(tw, localPath, filepath.Base(localPath))
	}
	closeErr := tw.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func addDirToTar(tw *tar.Writer, dir string) error {
	return filepath.WalkDir(dir, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == dir {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil // avoid exporting symlinks from Windows junctions or Unix symlinks
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		archiveName := filepath.ToSlash(rel)
		if info.IsDir() && !strings.HasSuffix(archiveName, "/") {
			archiveName += "/"
		}
		if info.IsDir() {
			hdr := &tar.Header{
				Name:     archiveName,
				Mode:     int64(info.Mode().Perm()),
				ModTime:  info.ModTime(),
				Typeflag: tar.TypeDir,
			}
			return tw.WriteHeader(hdr)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return addFileToTar(tw, p, archiveName)
	})
}

func addFileToTar(tw *tar.Writer, filePath, archiveName string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", filePath)
	}
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	hdr.Name = filepath.ToSlash(archiveName)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(tw, f)
	return err
}

func extractSingleDownloadedFile(r io.Reader, target string, overwrite bool) error {
	tr := tar.NewReader(r)
	found := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if found {
				return errors.New("download archive contains more than one file")
			}
			found = true
			if !overwrite {
				if _, err := os.Stat(target); err == nil {
					return fmt.Errorf("target exists: %s", target)
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileModeFromHeader(hdr, 0o640))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			_ = os.Chtimes(target, hdr.ModTime, hdr.ModTime)
		default:
			return fmt.Errorf("unsupported tar entry type %d", hdr.Typeflag)
		}
	}
	if !found {
		return errors.New("download archive did not contain a regular file")
	}
	return nil
}

func extractDownloadedDir(r io.Reader, targetDir string, overwrite bool) error {
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return err
	}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name, err := cleanArchivePath(hdr.Name)
		if err != nil {
			return err
		}
		if name == "" {
			continue
		}
		target, err := safeJoin(targetDir, name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, fileModeFromHeader(hdr, 0o750)); err != nil {
				return err
			}
		case tar.TypeReg:
			if !overwrite {
				if _, err := os.Stat(target); err == nil {
					return fmt.Errorf("target exists: %s", target)
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileModeFromHeader(hdr, 0o640))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			_ = os.Chtimes(target, hdr.ModTime, hdr.ModTime)
		default:
			return fmt.Errorf("unsupported tar entry type %d in %s", hdr.Typeflag, hdr.Name)
		}
	}
}

func cleanArchivePath(p string) (string, error) {
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "/")
	cleaned := path.Clean("/" + p)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("unsafe archive path: %q", p)
	}
	return cleaned, nil
}

func safeJoin(root, rel string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	candidate := filepath.Join(rootAbs, filepath.FromSlash(rel))
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	relToRoot, err := filepath.Rel(rootAbs, candidateAbs)
	if err != nil {
		return "", err
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(os.PathSeparator)) || filepath.IsAbs(relToRoot) {
		return "", fmt.Errorf("path escapes local target: %q", rel)
	}
	return candidateAbs, nil
}

func cleanBaseName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	if name == "." || name == "/" {
		return ""
	}
	return name
}

func fileModeFromHeader(h *tar.Header, fallback os.FileMode) os.FileMode {
	mode := os.FileMode(h.Mode) & 0o777
	if mode == 0 {
		return fallback
	}
	return mode
}

func looksLikeDir(p string) bool {
	return strings.HasSuffix(p, "/") || strings.HasSuffix(p, "\\")
}

func writeJSON(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if len(b) > maxJSONFrame {
		return fmt.Errorf("json frame too large: %d", len(b))
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(b)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func readJSON(r io.Reader, v any) error {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n == 0 || n > maxJSONFrame {
		return fmt.Errorf("invalid json frame length %d", n)
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
