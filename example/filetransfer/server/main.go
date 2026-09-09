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
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	SourceKind string `json:"source_kind,omitempty"` // file or dir, only used by upload
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

type limitReader struct {
	R io.Reader
	N int64
}

func (l *limitReader) Read(p []byte) (int, error) {
	if l.N <= 0 {
		var one [1]byte
		n, err := l.R.Read(one[:])
		if n > 0 {
			return 0, errors.New("upload exceeds configured server limit")
		}
		return 0, err
	}
	if int64(len(p)) > l.N {
		p = p[:l.N]
	}
	n, err := l.R.Read(p)
	l.N -= int64(n)
	return n, err
}

func main() {
	listenAddrsFlag := flag.String("listen", "0.0.0.0:4242", "comma-separated UDP listen address(es). For forced LAN + Tailscale use: 192.168.1.5:4242,100.91.211.76:4242")
	rootDir := flag.String("root", "./tcq-storage", "server storage root directory")
	token := flag.String("token", "", "shared auth token; leave empty only for a trusted lab LAN")
	maxUpload := flag.Int64("max-upload", 0, "maximum upload bytes per request; 0 means unlimited")
	flag.Parse()

	root, err := prepareStorageRoot(*rootDir)
	if err != nil {
		log.Fatalf("storage root is not writable: %v", err)
	}

	listenAddrs := splitListenAddrs(*listenAddrsFlag)
	if len(listenAddrs) == 0 {
		log.Fatalf("no listen address configured")
	}

	quicConf := &quic.Config{
		KeepAlivePeriod:                15 * time.Second,
		MaxIdleTimeout:                 5 * time.Minute,
		MaxIncomingStreams:             128,
		InitialStreamReceiveWindow:     4 << 20,
		MaxStreamReceiveWindow:         64 << 20,
		InitialConnectionReceiveWindow: 8 << 20,
		MaxConnectionReceiveWindow:     256 << 20,
	}
	tlsConf := generateTLSConfig()

	log.Printf("Storage root: %s", root)
	log.Printf("Upload rule: upload E:\\VIDEOS with no remote path is saved into %s", filepath.Join(root, "VIDEOS"))
	log.Printf("LAN client example:       client.exe -server 192.168.1.5:4242 -token <token> upload C:\\Data backup")
	log.Printf("Tailscale client example: client.exe -server 100.91.211.76:4242 -token <token> upload C:\\Data backup")
	if *token == "" {
		log.Printf("WARNING: no auth token configured; use -token for anything outside an isolated lab")
	}

	for _, addr := range listenAddrs {
		listener, err := quic.ListenAddr(addr, tlsConf, quicConf)
		if err != nil {
			log.Fatalf("listen %s: %v", addr, err)
		}
		log.Printf("TCQ file server listening on UDP %s", listener.Addr())
		go acceptLoop(listener, root, *token, *maxUpload)

		tcpListener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("tcp listen %s: %v", addr, err)
		}
		log.Printf("TCQ file server TCP fallback listening on %s", tcpListener.Addr())
		go acceptTCPLoop(tcpListener, root, *token, *maxUpload)
	}

	select {}
}

func splitListenAddrs(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		addr := strings.TrimSpace(part)
		if addr == "" {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out
}

func acceptLoop(listener *quic.Listener, root, token string, maxUpload int64) {
	defer listener.Close()
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("accept QUIC connection on %s: %v", listener.Addr(), err)
			continue
		}
		go handleConn(conn, root, token, maxUpload)
	}
}

func acceptTCPLoop(listener net.Listener, root, token string, maxUpload int64) {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept TCP connection on %s: %v", listener.Addr(), err)
			continue
		}
		log.Printf("accepted TCP fallback connection from %s", conn.RemoteAddr())
		go func(c net.Conn) {
			defer func() {
				log.Printf("closed TCP fallback connection from %s", c.RemoteAddr())
			}()
			handleStream(c, root, token, maxUpload)
		}(conn)
	}
}

func prepareStorageRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("empty storage root")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", abs)
	}
	test, err := os.CreateTemp(abs, ".tcq-write-test-*")
	if err != nil {
		return "", fmt.Errorf("cannot write to %s: %w", abs, err)
	}
	name := test.Name()
	if _, err := test.WriteString("ok"); err != nil {
		test.Close()
		os.Remove(name)
		return "", fmt.Errorf("cannot write test file in %s: %w", abs, err)
	}
	if err := test.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	if err := os.Remove(name); err != nil {
		return "", err
	}
	return abs, nil
}

func handleConn(conn *quic.Conn, root, token string, maxUpload int64) {
	log.Printf("accepted QUIC connection from %s", conn.RemoteAddr())
	defer func() {
		log.Printf("closed QUIC connection from %s", conn.RemoteAddr())
		conn.CloseWithError(0, "server done")
	}()
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			return
		}
		go handleStream(stream, root, token, maxUpload)
	}
}

func handleStream(stream io.ReadWriteCloser, root, token string, maxUpload int64) {
	defer stream.Close()

	var req request
	if err := readJSON(stream, &req); err != nil {
		_ = writeJSON(stream, response{OK: false, Error: "bad request: " + err.Error()})
		return
	}

	if req.Version != protocolVer {
		_ = writeJSON(stream, response{OK: false, Error: fmt.Sprintf("unsupported protocol version %d", req.Version)})
		return
	}
	if token != "" && subtle.ConstantTimeCompare([]byte(req.Token), []byte(token)) != 1 {
		_ = writeJSON(stream, response{OK: false, Error: "unauthorized"})
		return
	}

	var err error
	switch req.Op {
	case "upload":
		log.Printf("upload request: remote_path=%q source_kind=%q source_name=%q", req.RemotePath, req.SourceKind, req.SourceName)
		err = handleUpload(stream, root, req, maxUpload)
	case "download":
		err = handleDownload(stream, root, req)
	case "list":
		err = handleList(stream, root, req)
	default:
		err = fmt.Errorf("unknown op %q", req.Op)
	}
	if err != nil {
		_ = writeJSON(stream, response{OK: false, Error: err.Error()})
		return
	}
}

func handleUpload(rw io.ReadWriter, root string, req request, maxUpload int64) error {
	kind := strings.ToLower(req.SourceKind)
	if kind != "file" && kind != "dir" {
		return fmt.Errorf("upload requires source_kind=file or source_kind=dir")
	}

	reader := io.Reader(rw)
	if maxUpload > 0 {
		reader = &limitReader{R: rw, N: maxUpload}
	}

	sourceName := cleanBaseName(req.SourceName)
	remoteLooksDir := req.RemotePath == "" || req.RemotePath == "." || strings.HasSuffix(req.RemotePath, "/") || strings.HasSuffix(req.RemotePath, "\\")
	remote, err := cleanRemotePath(req.RemotePath)
	if err != nil {
		return err
	}

	if kind == "file" {
		if sourceName == "" {
			return errors.New("upload file requires source_name")
		}
		if remote == "" || remoteLooksDir {
			remote = path.Join(remote, sourceName)
		}
		target, err := safeJoin(root, remote)
		if err != nil {
			return err
		}
		log.Printf("saving uploaded file to %s", target)
		if err := extractSingleFile(reader, target, req.Overwrite); err != nil {
			return err
		}
		log.Printf("uploaded file saved: %s", target)
		return writeJSON(rw, response{OK: true, Message: "uploaded file", Kind: "file", Name: filepath.Base(target)})
	}

	if remote == "" {
		remote = sourceName
	}
	if remote == "" {
		return errors.New("upload dir requires a remote_path or source_name")
	}
	targetDir, err := safeJoin(root, remote)
	if err != nil {
		return err
	}
	log.Printf("saving uploaded directory to %s", targetDir)
	if err := extractDir(reader, targetDir, req.Overwrite); err != nil {
		return err
	}
	log.Printf("uploaded directory saved: %s", targetDir)
	return writeJSON(rw, response{OK: true, Message: "uploaded directory", Kind: "dir", Name: filepath.Base(targetDir)})
}

func handleDownload(w io.Writer, root string, req request) error {
	remote, err := cleanRemotePath(req.RemotePath)
	if err != nil {
		return err
	}
	target, err := safeJoin(root, remote)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	name := info.Name()
	kind := "file"
	if info.IsDir() {
		kind = "dir"
	}
	if err := writeJSON(w, response{OK: true, Message: "download follows as tar stream", Kind: kind, Name: name}); err != nil {
		return err
	}

	tw := tar.NewWriter(w)
	if info.IsDir() {
		err = addDirToTar(tw, target)
	} else {
		err = addFileToTar(tw, target, name)
	}
	closeErr := tw.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func handleList(w io.Writer, root string, req request) error {
	remote, err := cleanRemotePath(req.RemotePath)
	if err != nil {
		return err
	}
	target, err := safeJoin(root, remote)
	if err != nil {
		return err
	}
	entries := make([]listEntry, 0, 128)
	err = filepath.WalkDir(target, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == target {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(target, p)
		if err != nil {
			return err
		}
		kind := "file"
		if info.IsDir() {
			kind = "dir"
		} else if info.Mode()&os.ModeSymlink != 0 {
			kind = "symlink"
		}
		entries = append(entries, listEntry{
			Path:    filepath.ToSlash(rel),
			Kind:    kind,
			Size:    info.Size(),
			ModTime: info.ModTime().Format(time.RFC3339),
		})
		if len(entries) > 20000 {
			return errors.New("too many entries to list")
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return writeJSON(w, response{OK: true, Message: "list ok", Entries: entries})
}

func extractSingleFile(r io.Reader, target string, overwrite bool) error {
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
				return errors.New("file upload archive contains more than one file")
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
			mode := fileModeFromHeader(hdr, 0o640)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
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
		return errors.New("file upload archive did not contain a regular file")
	}
	return nil
}

func extractDir(r io.Reader, targetDir string, overwrite bool) error {
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
			return nil // avoid following or exporting symlinks
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

func cleanRemotePath(p string) (string, error) {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if p == "" || p == "." || p == "/" {
		return "", nil
	}
	cleaned := path.Clean("/" + p)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("unsafe remote path: %q", p)
	}
	return cleaned, nil
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
		return "", fmt.Errorf("path escapes server root: %q", rel)
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

func generateTLSConfig() *tls.Config {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("192.168.1.5"),
			net.ParseIP("100.91.211.76"),
		},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, priv.Public(), priv)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{certDER}, PrivateKey: priv}},
		NextProtos:   []string{protocolName},
	}
}
