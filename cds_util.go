// Ref. https://www.sobyte.net/post/2021-07/data-reading-in-golang/
//


package cds_util

import (
	"fmt"
	"log"
	"net"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
	"github.com/mattn/go-shellwords"
)


const BUFFTP = 60000
const BUFFER = 60000

const ctlver = "03"


var deb = false


type ftpbuf struct {
        header	string
        control string
        versao  string
        data    []byte
}



//------------------------------------------------
// 
func Println(v ...any) {
	if deb { log.Println(v) }
}


//------------------------------------------------
// 
func Printf(format string, v ...any) {
	if deb { log.Println(format, v) }
}


//------------------------------------------------
// 
func Send_msg(conn net.Conn, msg string) string {

//Printf("SEND> %s.\n", msg)

pmsg   := ""
sz     := len(msg)
offset := 0

Println("LEN:", sz)

for {
	if sz == 0 { break }
	ofx := offset

	if sz > BUFFER {
		pmsg = msg[offset:(offset + BUFFER)]

		sz     = sz - BUFFER
		offset = offset + BUFFER
	} else {
		pmsg = msg[offset:(offset + sz)]
		sz   = 0
	}
		
	Println("XMIT: offset:", ofx, ", size:", len(pmsg), ", remain:", sz)

	c,err := conn.Write(encode_msg(pmsg))
	if err != nil {
		Println("Falha ao transmitir mensagem.", c)
		return "Falha ao transmitir mensagem."
	}
}
return ""
}


//------------------------------------------------
// 
func Recv_msg(conn net.Conn) string {
rsz := 0

var buffer []byte

hdr := make([]byte, 7)     	// using small tmo buffer for demonstrating

n, err := conn.Read(hdr)
if err != nil {
	Println("read error:", err)
}
buffer = append(buffer, hdr[:n]...)



// ajustar o tamanho do buffer
shdr := string(hdr)

if shdr[5:7] == "ZM" {
	n,_ = strconv.Atoi(shdr[0:5])
}








tmp := make([]byte, n)     	// using small tmo buffer for demonstrating
rsz, err = conn.Read(tmp)
if err != nil {
	Println("read error:", err)
}
	
buffer = append(buffer, tmp[:rsz]...)
rsz = rsz + 7


mbuf := decode_msg(buffer, rsz)
if mbuf == "" {
	Println("Erro ao receber mensagem [", string(buffer), "].")
	return ""
}
Printf("RECV> %s.\n", mbuf)
return mbuf
}





//------------------------------------------------
// 
func encode_msg(msg string) []byte {

msgcode := fmt.Sprintf("%05dZM", len(msg))

return []byte(msgcode + msg)
}


//------------------------------------------------
// 
func decode_msg(msg []byte, rsz int) string {

zmsg := string(msg)

sz,err := strconv.Atoi(zmsg[0:5])
if err != nil {
	Println("Falha obter tamanho da mensagem.")
	return ""
}

ctl := zmsg[5:7]
if ctl != "ZM" {
	Println("Falha ao obter controle da mensagem.", ctl)
	return ""
}

if rsz != (sz + 7) {
	Println("Tamanho invalido da mensagem (", rsz, "x", (sz + 7), ").")
	return ""
}

return zmsg[7:(sz + 7)]
}




//------------------------------------------------------------
// Parse - Efetua o parse, quebrando a command line e uma matriz de argumentos
// ref. https://github.com/mattn/go-shellwords


func Parse(line string) ([]string, error) {
        return shellwords.NewParser().Parse(line)
}



//------------------------------------------------------------
// Exec

func Exec(zcmd string, label bool) (string, string) {
e := ""
x,err := Parse(zcmd)
if err != nil {
	e = fmt.Sprintf("Error: Falha no parsing do comando (%s), %s", zcmd, err.Error())
	eout := "Falha ao executar comando!\n"
	return eout,e
}

app  := x[0]
args := x[1:]

if label {
        fmt.Printf("______________________________________\n")
        fmt.Printf("Comando: %s\n", zcmd)
        fmt.Printf("Args...: [%s] [%s]\n", app, args)

        fmt.Println(app," ==> ", args)
}
cmd := exec.Command(app, args...)
cmd.Env = append(os.Environ(),
	"CDSHELL=1",
)

out, err := cmd.CombinedOutput()
if err != nil {
	//fmt.Println("ERR>>", string(out))
        e = fmt.Sprintf("Error: %s", err.Error())
}
//fmt.Printf("\n%s\n", string(out))
return string(out), e
}






//------------------------------------------------
// 
func Recv_ctrl(conn net.Conn) []byte {


hdr := make([]byte, 7)
rr  := io.LimitReader(conn, 7)

for {
	n, err := io.ReadAtLeast(rr, hdr, 7)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println("read error:", n, err)
		}
		break
	}
}

/* ---------
n, err := io.LimitRead(conn, hdr, 7)
if err != nil || n != 7 {
	fmt.Println("read error:", n, err)
}
*/

fmt.Println("READER_HDR>", string(hdr[0:7]))
sz,err := strconv.Atoi(string(hdr[0:5]))
if err != nil {
	fmt.Println("Atoi error [", string(hdr[0:5]), "):", err)
	fmt.Println("Falha no recebimento de informacoes.")
}



tmp := make([]byte, sz)
rd  := io.LimitReader(conn, int64(sz))

for {
	n, err := io.ReadAtLeast(rd, tmp, sz)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println("read error:", n, err)
		}
		fmt.Println("READER_DATA>", n)
		break
	}
}

fmt.Println("READER_EDATA>", sz)
/*
rsz, err := conn.Read(tmp)
fmt.Println("READER_DATA>", rsz)
if err != nil || rsz != sz {
	fmt.Println("read error:", rsz, err)
}
*/
	
//Printf("RECV> [%s]\n", string(hdr))
//Printf("RECV> (%d) \n", rsz)
return tmp
}






//------------------------------------------------
// 
func Send_ctrl(conn net.Conn, control string, data []byte, size int) string {

prehdr := fmt.Sprintf("%05dZM", (size + 3))
//Println(prehdr, size)


xbuff := make([]byte, 0)
xbuff = append(xbuff[:], []byte(prehdr)...)

xbuff = append(xbuff[:], []byte(control)...)

xbuff = append(xbuff[:], []byte(ctlver)...)

xbuff = append(xbuff[:], data[:size]...)


c,err := conn.Write(xbuff)
if err != nil {
	Println("Falha ao transmitir mensagem.", c)
	return "Falha ao transmitir mensagem."
}
//Println("Enviado ", c, "bytes")
return ""
}



//------------------------------------------------
// 
func ConfigureKeepAlive(conn net.Conn) error {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("conexão não é TCP")
	}

	if err := tcpConn.SetKeepAlive(true); err != nil {
		return err
	}

	if err := tcpConn.SetKeepAlivePeriod(60 * time.Second); err != nil {
		return err
	}

	return nil
}
