//go:build ceph

package targets

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func init() {
	defaultFactory.RegisterProvider("rbd", RBDCreateTarget)
}

type RBDTarget struct {
	ioctx *rados.IOContext
	image *rbd.Image
	conn  *rados.Conn

	imageName string
}

func (t *RBDTarget) GetSize() uint64 {
	sz, _ := t.image.GetSize()
	return sz
}

func (t *RBDTarget) Queue(r *IORequest) TargetError {
	switch r.Command {
	case IORequestCmdRead:
		offset := int64(r.Lba * 512)

		for i := range r.Buffers() {
			cnt, err := t.image.ReadAt(r.SGL[i].Data, offset)
			if err != nil {
				fmt.Printf("Read Error: %s\n", err.Error())
				return r.Complete(TargetErrorUnsupported)
			}
			if cnt != len(r.SGL[i].Data) {
				panic(false)
			}
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdWrite:
		offset := int64(r.Lba * 512)

		for i := range r.Buffers() {
			cnt, err := t.image.WriteAt(r.SGL[i].Data, offset)
			if err != nil {
				fmt.Printf("Write Error: %s\n", err.Error())
				return r.Complete(TargetErrorUnsupported)
			}
			if cnt != len(r.SGL[i].Data) {
				panic(false)
			}
			offset += int64(len(r.SGL[i].Data))
		}
		return r.Complete(TargetErrorNone)

	case IORequestCmdFlush:
		// FIXME: need to do a proper flush
		return r.Complete(TargetErrorNone)

	case IORequestCmdTrim, IORequestCmdWriteZero:
		return r.Complete(TargetErrorNone)
		/*
			offset := int64(r.Lba * 512)
			cnt, err := t.image.Trim(offset, r.Length)
			if err != nil {
				return r.Complete(TargetErrorNone)
			}
			return r.Complete(TargetErrorNone)*/

	default:
		return r.Complete(TargetErrorUnsupported)
	}
}

func (t *RBDTarget) Start() error {
	return nil
}

func (t *RBDTarget) Close() error {
	t.image.Close()
	t.ioctx.Destroy()
	t.conn.Shutdown()
	t.conn = nil
	return nil
}

func (t *RBDTarget) GetRuntimeDetails() []KV {
	return []KV{
		{
			Key:   "Image",
			Value: t.imageName,
		},
	}
}

func RBDCreateTarget(options Options) (Target, error) {
	var cluster, user, pool, image string
	var ok bool

	if cluster, ok = options["cluster"]; !ok {
		return nil, errors.New("no cluster name provided")
	}

	if user, ok = options["user"]; !ok {
		return nil, errors.New("no cluster user provided")
	}

	if pool, ok = options["pool"]; !ok {
		return nil, errors.New("no cluster pool provided")
	}

	if image, ok = options["image"]; !ok {
		return nil, errors.New("no cluster image provided")
	}

	conn, err := rbdConnect(cluster, user, 0)
	if err != nil {
		return nil, err
	}

	ioctx, err := conn.OpenIOContext(pool)
	if err != nil {
		conn.Shutdown()
		return nil, fmt.Errorf("can't open IOContext")
	}

	rbdImage := rbd.GetImage(ioctx, image)
	if rbdImage == nil {
		ioctx.Destroy()
		conn.Shutdown()
		return nil, fmt.Errorf("could not find image %s", image)
	}

	if err := rbdImage.Open(false); err != nil {
		conn.Shutdown()
		ioctx.Destroy()
		return nil, fmt.Errorf("could not find image %s", image)
	}

	t := &RBDTarget{
		ioctx:     ioctx,
		image:     rbdImage,
		conn:      conn,
		imageName: fmt.Sprintf("%s/%s/%s", cluster, pool, image),
	}

	return NewWorkQueue(options, t), nil
}

func rbdConnect(cluster string, user string, opTimeout float64) (*rados.Conn, error) {
	conn, err := rados.NewConnWithUser(user)
	if err != nil {
		return nil, err
	}

	configPath := fmt.Sprintf("/etc/ceph/%s.conf", cluster)
	if err := conn.ReadConfigFile(configPath); err != nil {
		return nil, err
	}

	if opTimeout > 0 {
		err := conn.SetConfigOption("rados_osd_op_timeout",
			strconv.FormatFloat(opTimeout, 'f', 2, 64))
		if err != nil {
			return nil, err
		}
	}

	if err := conn.Connect(); err != nil {
		return nil, err
	}

	return conn, nil
}
