package log

import (
	syslog "log"
	"reflect"
	"time"
)

type LogWriter struct {
	writers  []Writer
	tunnel   chan *Record
	c        chan bool
	stopchan chan struct{}
}

func (this *LogWriter) Init() {
	this.writers = make([]Writer, 0, 2)
	this.tunnel = make(chan *Record, tunnel_size_default)
	this.stopchan = make(chan struct{})
	this.c = make(chan bool, 1)

	go this.boostrapLogWriter()
}

func (this *LogWriter) AddLogFile(filename string, redirecterr bool) {
	//	fmt.Printf("log filename,%s \n", filename)
	filebasename := filename
	filename += ".%Y%M%D-%H"
	w := NewFileWriter()
	if redirecterr {
		w.Redirecterr = true
	}
	err := w.SetPathPattern(filebasename, filename)
	if err != nil {
	}
	this.registerLogWriter(w)
}

func (this *LogWriter) ChangeLogFile(filename string) {
	filebasename := filename
	filename += ".%Y%M%D-%H"
	for i := 0; i < len(this.writers); i++ {
		w := this.writers[i]
		if reflect.TypeOf(w).String() == "*log.FileWriter" {
			if r, ok := w.(Rotater); ok {
				err := r.SetPathPattern(filebasename, filename)
				if err != nil {
				}
				err = r.Rotate()
				if err != nil {
				}
			}
		}
	}
}

func (this *LogWriter) RemoveConsoleLog() {
	newlist := make([]Writer, 0, 2)
	for i := 0; i < len(this.writers); i++ {
		w := this.writers[i]
		//		Debug("start RemoveConsoleLog, %s", reflect.TypeOf(w).String())
		if reflect.TypeOf(w).String() != "*log.ConsoleWriter" {
			newlist = append(newlist, w)
		}
	}
	this.writers = newlist
}

func (this *LogWriter) registerLogWriter(w Writer) {
	if err := w.Init(); err != nil {
		panic(err)
	}
	this.writers = append(this.writers, w)
}

func (this *LogWriter) Close() {
	select {
	case <-this.stopchan:
		return
	default:
		close(this.stopchan)
		// close(this.tunnel)
		break
	}
	select {
	case <-this.c:
		break
	}

	for _, w := range this.writers {
		if f, ok := w.(Flusher); ok {
			if err := f.Flush(); err != nil {
				syslog.Println(err)
			}
		}
	}
}

func (this *LogWriter) write(r *Record) {
	select {
	case <-this.stopchan:
		return
	default:
	}
	select {
	case <-this.stopchan:
		break
	case this.tunnel <- r:
		break
	}
}

func (this *LogWriter) boostrapLogWriter() {
	var (
		r  *Record
		ok bool
	)

	if r, ok = <-this.tunnel; !ok {
		this.c <- true
		return
	}

	for _, w := range this.writers {
		if err := w.Write(r); err != nil {
			syslog.Println(err)
		}
	}

	flushTimer := time.NewTimer(time.Millisecond * 500)
	rotateTimer := time.NewTimer(time.Millisecond * 100)
	//	rotateTimer := time.NewTimer(time.Second * 10)
	lastRecordTimeUnix60 := int64(0)
	lastRecordTime := time.Now()
	for {
		select {
		case r, ok = <-this.tunnel:
			if !ok {
				this.c <- true
				return
			}
			needTryRotate := false
			// 如果上一条记录的时间和这条记录的时间不是同一分钟，需要尝试增加log文件
			nowTimeUnix60 := r.timeUnix / 60
			if lastRecordTimeUnix60 != nowTimeUnix60 {
				needTryRotate = true
				lastRecordTime = time.Unix(r.timeUnix, 0)
				// 更新最后一条记录的时间
				lastRecordTimeUnix60 = nowTimeUnix60
			}
			for _, w := range this.writers {
				if needTryRotate {
					// 需要增加log文件
					if r, ok := w.(Rotater); ok {
						if err := r.RotateByTime(&lastRecordTime); err != nil {
							syslog.Println(err)
						}
					}
				}
				// 写入log
				if err := w.Write(r); err != nil {
					syslog.Println(err)
				}
			}
			recordPool.Put(r)
		case <-this.stopchan:
			this.c <- true
			return
		case <-flushTimer.C:
			for _, w := range this.writers {
				if f, ok := w.(Flusher); ok {
					if err := f.Flush(); err != nil {
						syslog.Println(err)
					}
				}
			}
			flushTimer.Reset(time.Millisecond * 500)
		case <-rotateTimer.C:
			//	fmt.Printf("start rotate file,actions, 1111\n")
			this.tryRotate()
			rotateTimer.Reset(time.Second * 60)
		}
	}
}

func (this *LogWriter) tryRotate() {
	for _, w := range this.writers {
		if r, ok := w.(Rotater); ok {
			if err := r.Rotate(); err != nil {
				syslog.Println(err)
			}
		}
	}
}
