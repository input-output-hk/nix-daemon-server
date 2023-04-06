package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kr/pretty"
	"github.com/nix-community/go-nix/pkg/nar"
	"github.com/nix-community/go-nix/pkg/wire"
	"github.com/pkg/errors"
)

const (
	StderrLast      = 0x616C7473 // stla
	StderrError     = 0x63787470 // ptxc
	WorkerMagic1    = 0x6E697863 // cxin
	WorkerMagic2    = 0x6478696F // ioxd
	ProtocolVersion = 1<<8 | 34  // 1.34
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("no DATABASE_URL set")
	}

	dbConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		panic(err)
	}

	db, err := pgxpool.ConnectConfig(context.Background(), dbConfig)
	if err != nil {
		panic(err)
	}

	c := client{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		db:     db,
	}

	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	if err := c.handshake(); err != nil {
		panic(err)
	} else if err := c.handleOperations(); err != nil {
		panic(err)
	}
}

type client struct {
	db     *pgxpool.Pool
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	err    error
}

func (c *client) handshake() error {
	if wm1, err := wire.ReadUint64(c.stdin); err != nil {
		return errors.WithMessage(err, "reading worker magic 1")
	} else if wm1 != WorkerMagic1 {
		return errors.Errorf("worker magic 1 mismatch: '%x' should be '%x'", wm1, WorkerMagic1)
	} else if err := wire.WriteUint64(c.stdout, WorkerMagic2); err != nil {
		return errors.WithMessage(err, "while writing worker magic 2")
	} else if err := wire.WriteUint64(c.stdout, ProtocolVersion); err != nil {
		return errors.WithMessage(err, "while writing server protocol version")
	} else if clientProtocolVersion, err := wire.ReadUint64(c.stdin); err != nil {
		return errors.WithMessage(err, "while reading client protocol version")
	} else if err := wire.WriteString(c.stdout, "2.11.2"); err != nil {
		return errors.WithMessage(err, "while writing nix version")
	} else if err := wire.WriteUint64(c.stdout, StderrLast); err != nil {
		return errors.WithMessage(err, "while writing StderrLast")
	} else {
		_, _ = wire.ReadUint64(c.stdin) // cpu affinity
		_, _ = wire.ReadUint64(c.stdin) // reserve space
		io.WriteString(c.stderr, pretty.Sprint(clientProtocolVersion)+"\n")
	}

	return nil
}

func (c *client) handleOperations() error {
	for {
		if operation, err := wire.ReadUint64(c.stdin); err != nil {
			if err == io.EOF {
				io.WriteString(c.stderr, "EOF\n")
				return nil
			}
			return errors.WithMessage(err, "while reading operation")
		} else {
			workerOperation := WorkerOperation(operation)
			io.WriteString(c.stderr, workerOperation.String()+"\n")

			switch workerOperation {
			case WOPQueryValidPaths:
				c.queryValidPaths()
			case WOPRegisterDrvOutput:
				c.registerDrvOutput()
			case WOPAddMultipleToStore:
				c.addMultipleToStore()
			case WOPAddTempRoot:
				c.addTempRoot()
			case WOPQueryMissing:
				c.queryMissing()
			case WOPIsValidPath:
				c.isValidPath()
			case WOPQueryPathInfo:
				c.queryPathInfo()
			default:
				return errors.Errorf("unknown operation: %s", workerOperation.String())
			}

			if c.err != nil {
				return err
			}
		}
	}
}

func (c *client) queryPathInfo() {
	storePath := c.readString(1024 * 4)
	c.debug("queryPathInfo:", storePath)

	if c.err == nil {
		info := validPathInfo{}
		if err := pgxscan.Select(
			context.Background(), c.db, &info, `SELECT * FROM ValidPaths WHERE path = $1;`, storePath,
		); err != nil {
			c.err = err
		}
	}

	c.writeStderrLast()
	c.writeBool(true)
}

func (c *client) queryValidPaths() {
	paths := c.readStrings()
	substitute := c.readBool()
	c.writeStderrLast()
	c.writeStrings([]string{})
	c.debug("paths:", paths, "substitute:", substitute)
}

func (c *client) isValidPath() {
	storePath := c.readString(1024 * 4)
	c.debug("isValidPath:", storePath)
	c.writeStderrLast()
	c.writeBool(true)
}

func (c *client) addTempRoot() {
	storePath := c.readString(1024 * 4)
	c.debug("tmproot:", storePath)
	c.writeStderrLast()
	c.writeInt(1)
}

func (c *client) queryMissing() {
	targets := c.readStrings()
	c.debug("targets:", targets)
	c.writeStderrLast()

	willBuild := []string{}
	willSubstitute := []string{}
	unknown := []string{}

	unknown = append(willSubstitute, targets...)

	c.writeStrings(willBuild)
	c.writeStrings(willSubstitute)
	c.writeStrings(unknown)
	c.writeInt(100)
	c.writeInt(100)
	c.writeStderrLast()
}

func (c *client) addMultipleToStore() {
	repair := c.readBool()
	dontCheckSigs := c.readBool()
	c.debug("repair:", repair, "dontCheckSigs:", dontCheckSigs)
	narSource := newFramedSource(c.stdin)
	if err := c.parseSource(narSource); err != nil {
		c.err = err
	}
	c.writeStderrLast()
	if n := c.readInt(); n != 0 {
		c.err = errors.New("invalid result status")
	}
}

func (c *client) registerDrvOutput() {
	realisation := c.readString(1024 * 10)
	c.writeStderrLast()
	c.debug("realisation:", realisation)
}

func (c *client) parseSource(s io.Reader) error {
	expected, err := wire.ReadUint64(s)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return errors.WithMessage(err, "reading expected")
	}

	c.debug("expected:", expected)

	for i := uint64(0); i < expected; i += 1 {
		if info, err := readNarinfo(s); err != nil {
			c.debug("err:", err.Error())
			return errors.WithMessage(err, "reading Narinfo")
		} else {
			c.debug("narinfo:", info.OutPath)
			swallow(s, int64(info.NarSize))
		}
	}

	return nil
}

func swallow(s io.Reader, n int64) {
	fd, err := os.Create("swallow")
	if err != nil {
		panic(err)
	}

	copied, err := io.Copy(fd, io.LimitReader(s, n))

	if err != nil {
		panic(err)
	}

	if copied != n {
		panic(fmt.Sprintf("copied %d of %d bytes", copied, n))
	}
}

type validPathInfo struct {
	OutPath          string `db:path`
	Deriver          string `db:deriver`
	NarHash          string `db:hash`
	References       []string
	RegistrationTime time.Time `db:registration_rime`
	NarSize          uint64    `db:nar_size`
	Ultimate         bool      `db:ultimate`
	Sigs             []string  `db:sigs`
	CA               string    `db:ca`
}

func readNarinfo(s io.Reader) (*validPathInfo, error) {
	info := &validPathInfo{}
	var err error

	if info.OutPath, err = wire.ReadString(s, 1024*10); err != nil {
		return nil, errors.WithMessage(err, "reading StorePath")
	} else if info.Deriver, err = wire.ReadString(s, 1024*10); err != nil {
		return nil, errors.WithMessage(err, "reading Deriver")
	} else if info.NarHash, err = wire.ReadString(s, 1024); err != nil {
		return nil, errors.WithMessage(err, "reading NarHash")
	} else if info.References, err = readStrings(s); err != nil {
		return nil, errors.WithMessage(err, "reading References")
	}

	registrationTimeUnix, err := wire.ReadUint64(s)
	if err != nil {
		return nil, errors.WithMessage(err, "reading registrationTime")
	}
	info.RegistrationTime = time.Unix(int64(registrationTimeUnix), 0)

	if info.NarSize, err = wire.ReadUint64(s); err != nil {
		return nil, errors.WithMessage(err, "reading narSize")
	} else if info.Ultimate, err = wire.ReadBool(s); err != nil {
		return nil, errors.WithMessage(err, "reading ultimate")
	}

	if info.Sigs, err = readStrings(s); err != nil {
		return nil, errors.WithMessage(err, "reading Sigs")
	}

	if info.CA, err = wire.ReadString(s, 1024); err != nil {
		return nil, errors.WithMessage(err, "reading CA")
	}

	return info, nil
}

func readNar(s io.Reader) error {
	n, err := nar.NewReader(s)
	if err != nil {
		return err
	}
	for {
		header, err := n.Next()
		pp(header, err)

		if err != nil {
			if err.Error() == "unexpected EOF" {
				return nil
			}
			return errors.WithMessage(err, "getting NAR header")
		}

		switch header.Type {
		case nar.TypeSymlink:
			pp("sym", header.Path, header.LinkTarget)
		case nar.TypeDirectory:
			pp("dir", header.Path)
		case nar.TypeRegular:
			pp("reg", header.Path, header.Size, header.Executable)
			// buf := bytes.Buffer{}
			// pp(io.Copy(&buf, n))
		}
	}
}

func (c *client) readString(max uint64) (out string) {
	if c.err == nil {
		out, c.err = wire.ReadString(c.stdin, max)
		return
	}
	return
}

func (c *client) readStrings() (out []string) {
	if c.err == nil {
		out, c.err = readStrings(c.stdin)
		return
	}
	return
}

func (c *client) readInt() (out uint64) {
	if c.err == nil {
		out, c.err = wire.ReadUint64(c.stdin)
		return
	}
	return
}

func (c *client) readBool() (out bool) {
	if c.err == nil {
		out, c.err = wire.ReadBool(c.stdin)
		return
	}
	return
}

func (c *client) writeBool(value bool) {
	if c.err == nil {
		c.err = wire.WriteBool(c.stdout, value)
	}
}

func (c *client) writeInt(value uint64) {
	if c.err == nil {
		c.err = wire.WriteUint64(c.stdout, value)
	}
}

func (c *client) writeStderrLast() {
	c.writeInt(StderrLast)
}

func (c *client) writeStrings(value []string) {
	if c.err == nil {
		c.err = writeStrings(c.stdout, value)
	}
}

func (c *client) debug(value ...any) {
	if c.err == nil {
		io.WriteString(c.stderr, pretty.Sprint(value...)+"\n")
	}
}

func readStrings(s io.Reader) ([]string, error) {
	size, err := wire.ReadUint64(s)
	if err != nil {
		return nil, err
	}

	output := make([]string, size)

	for i := uint64(0); i < size; i += 1 {
		path, err := wire.ReadString(s, 2048)
		if err != nil {
			return nil, err
		}
		output[i] = path
	}

	return output, nil
}

func writeStrings(s io.Writer, strings []string) error {
	if err := wire.WriteUint64(s, uint64(len(strings))); err != nil {
		return err
	}

	for _, str := range strings {
		if err := wire.WriteString(s, str); err != nil {
			return err
		}
	}

	return nil
}

func pp(args ...any) {
	pretty.Println(args...)
}
