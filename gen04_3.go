/*----------------------------------------------------------------------

Gerador de imagem (2 cores x 1bits x 3 repetidos ==> tamanho de 1 byte = 8bits x 3 repeat = 24bytes)

Layout de saida:

Offset	size	szpx	Pos	Função
header:
0	1b	1	0x0	Valor inicial da Palette (pal[0])
1	10	240	1x0	Matriz de controle (bit 1 x 128)
241	10	240		Matriz de controle (bit 0 x 128)
481	1	24	481x0	Numero de bits por cor (1 bit)
505	1	24	505x0 	Numero de repeticoes 3
529	9	216	529x0	Tamanho do arquivo (9 bytes)
745	4	96	769x0 	Numero da sequencia (4 bytes)
841	52992	1435752 865x0	Dados da imagem (bloco: 6936, caapcidade para 207) 16x16x207 = 52992

Image Data: 
(207 blocos, 6936 (17x17x24, Dados transmitidos: 16x16x207 = 52992 bytes)

tail:
1438600 141	3384	400x798 Tail de informacoes(MZ:fileorg,filename,desc,STATUS,seq,MZEOF,,,,)



Exemplo:

(0,0) Header.1 Palette
(1,0) Header.1 Palette Sync.
(481,0) Header.2 modelo 1bit
(505,0) Header.2 3 repeat
(529,0) Header.2 image size: 52992 , Offset: 52992  (numero maximo de blocos: 207)
(769,0) Header.2 Sequence: 0000
(865,0) Image data

Current position: 865,0

(217,798) Tail.0 (image end)
(1561,799) Tail.1 (tail end)
(400,798) Tail.2 , information size: 55
Preparer tail:  MZ:zimg.tar,zcode4_0.bmp,Teste marcello,NEXT,0,MZEOF,,,

----------------------------------------------------------------------*/



package main

import (
	"image"
	"image/color"
	"image/color/palette"
	"github.com/zimg"
	//"image/png"
	"os"
	"os/exec"
	"fmt"
	//"strconv"
	"flag"
	//"bufio"
	//"io"
	//"io/ioutil"
	//"encoding/binary"

)


var Pal16 = []color.Color{
        color.RGBA{0x00, 0x00, 0x00, 0xff},    // Lumin: 00
        color.RGBA{0x10, 0x10, 0x10, 0xff},    // Lumin: 10
        color.RGBA{0x20, 0x20, 0x20, 0xff},    // Lumin: 20
        color.RGBA{0x30, 0x30, 0x30, 0xff},    // Lumin: 30
        color.RGBA{0x40, 0x40, 0x40, 0xff},    // Lumin: 40
        color.RGBA{0x50, 0x50, 0x50, 0xff},    // Lumin: 50
        color.RGBA{0x60, 0x60, 0x60, 0xff},    // Lumin: 60
        color.RGBA{0x70, 0x70, 0x70, 0xff},    // Lumin: 70
        color.RGBA{0x80, 0x80, 0x80, 0xff},    // Lumin: 80
        color.RGBA{0x90, 0x90, 0x90, 0xff},    // Lumin: 90
        color.RGBA{0xa0, 0xa0, 0xa0, 0xff},    // Lumin: a0
        color.RGBA{0xb0, 0xb0, 0xb0, 0xff},    // Lumin: b0
        color.RGBA{0xc0, 0xc0, 0xc0, 0xff},    // Lumin: c0
        color.RGBA{0xd0, 0xd0, 0xd0, 0xff},    // Lumin: d0
        color.RGBA{0xe0, 0xe0, 0xe0, 0xff},    // Lumin: e0
        color.RGBA{0xf0, 0xf0, 0xf0, 0xff} }   // Lumin: f0



var pal color.Palette
var px int
var py int
var idx uint8

//CRC
var acrc [17][17]byte
var pcx int
var pcy int

var ccrc int
var pcrc =0
var tcrc uint8
var size = 0
var mask uint8 = 0
var mtip uint8 = 0


var bloco = 0
var isdata = true
var ismoref= false


var deb=false	

var nfile string = ""
var descf string = ""


var nbloco int64 = 207


//---------------------------------------------------------------------------------------------------------

func main() {


flag.StringVar(&nfile, "f", "teste.tar", "Nome do arquivo a ser processado.")
flag.StringVar(&descf, "d", "Sem descricao.", "Descrição do arquivo gerado.")

flag.Parse()



fmt.Println("Inciando a geracao de imagem (base: ", nfile, ")...")


// obter tamanho arquivo de entrada...
fs,_ := os.Lstat(nfile)

fsize := fs.Size()
offset:= int64(0)


// Arquivo de entrada...
z,err  := os.Open(nfile)
if err != nil { panic(err) }

r := make([]byte, 1)

fimg := 0 




for {

fmt.Println("Gerando arquivo de imagem (zcode: ", fimg, ")\n")
size  = 0
bloco = 0

px=0
py=0


pal = palette.Plan9
pal = Pal16

var img *image.RGBA
img = image.NewRGBA(image.Rect(0,0,1919,801))


fmt.Printf("(%d,%d) Header.1 Palette\n",px,py)
setbit(img, 0)

fmt.Printf("(%d,%d) Header.1 Palette Sync.\n",px,py)
for i:=0; i< 10 ; i++ { setpixel(img,255) }
for i:=0; i< 10 ; i++ { setpixel(img,0) }


fmt.Printf("(%d,%d) Header.2 modelo 1bit\n",px,py)
setpixel(img, 1)
fmt.Printf("(%d,%d) Header.2 3 repeat \n",px,py)
setpixel(img, 3)


// calcular image size do bloco
s := make([]byte, 10)
dsize := (fsize - offset)

if dsize > (nbloco * 256) {
	dsize  = (nbloco * 256)
	offset = offset + (nbloco * 256)
} 
	
s = []byte(fmt.Sprintf("%-9d ", dsize))

fmt.Printf("(%d,%d) Header.2 image size: %d , Offset: %d  (numero maximo de blocos: %d)\n",px,py,dsize,offset,nbloco)
for i:=0; i<10; i++ {
	//fmt.Printf("%02X ", s[i])
	setpixel(img, s[i])

}

//fmt.Printf("Size: %s, Offset: %d\n",s, offset)




//Salvar o nome do arquivo original...
seqn   := fmt.Sprintf("%04d", fimg)
		
fmt.Printf("(%d,%d) Header.2 Sequence: %s\n",px,py,seqn[:4])

setpixel(img, seqn[0])
setpixel(img, seqn[1])
setpixel(img, seqn[2])
setpixel(img, seqn[3])

	
//fmt.Printf("%02X %02X %02X %02X %02X ", seqn[0],seqn[1],seqn[2],seqn[3])

fmt.Printf("(%d,%d) Image data \n",px,py)
fmt.Printf("Current position: %d,%d\n", px, py)


// inicializar crc
ccrc=0
pcx=0
pcy=0

for {
	_,err = z.Read(r)
	if err != nil { 
		fmt.Println("EOF")
		ismoref = false
		break 
	}

	size++
	p:= uint8(r[0])

	setpixel(img, p)

if deb { fmt.Printf("(%d):%02X ", ccrc, p)}
	crc(p, img)


// zana
	// verificar bloco limite nbloco blocos de dados
	if bloco >= int(nbloco) {
		fmt.Println("\nBloco limite atingido! (size:", size, ", bloco:", bloco, "). Necessario criar novo arquivo!")
		ismoref = true
		break
	}
}
fmt.Printf("\nTotal de crc: %d\n", pcrc)



// finalizar imagem....
fmt.Printf("(%d,%d) Tail.0 (image end)\n",px,py)
isdata = false
for y:=py; y < (800 - 1); y++ {
        for x:=px; x< 1920; x++ {
		setpixel(img, 0)
		crc(0, img)
	}
}

fmt.Println("\n")
fmt.Printf("(%d,%d) Tail.1 (tail end)\n",px,py)

// definir nome do arquivo corrente...
filename := fmt.Sprintf("zcode4_%d.bmp", fimg)


//Salvar o nome do arquivo original...
seqf   := fmt.Sprintf(",%d", fimg)
iscont := "LAST"
if ismoref {
	iscont = "NEXT"
}
bt := "MZ:" + nfile + "," + filename + "," + descf + "," + iscont + seqf + ",MZEOF,,,"
bc := len(bt)

isdata = false
px=400
py=798
xbc := 0
fmt.Printf("(%d,%d) Tail.2 , information size: %d\n",px,py, bc)
fmt.Println("Preparer tail: ", bt)
for x:=px; x < 1920 ; x++ {
	if xbc < bc {
		setpixel(img, bt[xbc])
		//fmt.Printf("%d x %c\n", xbc, bt[xbc])
		xbc++	
	} else {
		setpixel(img, 0)
	}
}

fmt.Println("")
fmt.Printf("\nTotal de crc: %d\n", pcrc)



// salvar imagem....
fimg++

f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
defer f.Close()

// formato do arquivo de saida...
//png.Encode(f, img)
zimgbmp.Encode(f, img)

fmt.Println("Filename image:", filename)
fmt.Println("Tamanho total: ", dsize, "bytes, offset:", offset)
f.Close()

if 0 == 1 {
	cmd := exec.Command("cp", "teste.bmp", "/cdshell/html/go/go.bmp")
	_ = cmd.Run()
	fmt.Println("Efetuado a copia do arquivo!")
}



// verificar se existe mais dados do arquivo, repetir looping...
if ! ismoref { break }

fmt.Println("________________________________\n\n\n")

}



fmt.Println("Processamento encerrado (", fimg, " arquivos de imagens gerados).\n")
os.Exit(0)
}





//---------------------------------------------------------------------------------------------------------

func addpixel() {
px++

if px >= 1920 && py >= 799 {
	if isdata { 
		fmt.Println("Estouro de limite da imagem (1800 x 800).", size, ", blocos: ", bloco)
	}
}

if px >= 1920 { 
	px = 0
	py++
}

if py >= 800 {  py=799 }

return
}



//-----------------------------------------------------------
//
func crc(v uint8, im image.Image) {


if ccrc < 256 {
	acrc[pcx][pcy] = v

	ccrc++
	pcy++
	if pcy >= 16 {
		pcy =0
		pcx++
	}
} 

if ccrc == 256 {
	// calculate e fflush do crc
	pcx = 0
	pcy = 0
	ccrc= 0
	tcrc = 0

	if deb {
		for i:=0; i<=16; i++ {
			for j:=0; j<=16; j++ {
			fmt.Printf("%02X ", acrc[i][j])
			}
			fmt.Println("")
		}
			fmt.Println("")
	}


	// inicializar CRC´s
	for i:=0; i<16; i++ {
		acrc[16][i]=0
		acrc[i][16]=0
	}


	// calculate CRC linha
	for i:=0; i<16; i++ {
		tcrc = 0
		for j:=0; j<16; j++ { tcrc = acrc[i][j] ^ tcrc }
		acrc[i][16] = tcrc
	
		//fmt.Println(tcrc, i, 16)
	}
			
	// calculate CRC coluna
	for j:=0; j<16; j++ {
		tcrc = 0
		for i:=0; i<16; i++ { tcrc = acrc[i][j] ^ tcrc }
		acrc[16][j] = tcrc

		//fmt.Println(tcrc, 16, j)
	}

	// calculate block CRC
	tcrc = 0
	for i:=0; i<16; i++ {
		tcrc = acrc[16][i] ^ tcrc
		tcrc = acrc[i][16] ^ tcrc
	}
	acrc[16][16] = tcrc
	
//fmt.Println(acrc)

	// save CRC linha		
	for i:=0; i<16; i++ {
		setpixel(im, acrc[i][16])
		//fmt.Println("ADD",acrc[i][16],"x", i)
//fmt.Printf("%02X ", acrc[i][16])
	}
//fmt.Println("")

	// save CRC coluna		
	for i:=0; i<16; i++ {
		setpixel(im, acrc[16][i])
		//fmt.Println("ADD",acrc[16][i],"y", i)
//fmt.Printf("%02X ", acrc[16][i])
	}

	// Save CRC block
	setpixel(im, acrc[16][16])
//fmt.Printf("%02X ", acrc[16][16])


	if  deb {
		for i:=0; i<=16; i++ {
		for j:=0; j<=16; j++ {
			fmt.Printf("%02X ", acrc[i][j])
			}
			fmt.Println("")
		}
			fmt.Println("==========================")
	}


	pcrc++
	//size = size + 33
	bloco++
	fmt.Printf(".")
	return
}


return 
}



//-----------------------------------------------------------
//
func setbit(m image.Image, by uint8)  {

m.(*image.RGBA).Set(px,py, pal[by])
addpixel()
}

//-----------------------------------------------------------
//
func setpixel(m image.Image, by uint8)  {



za := (by & 0xf0) >> 4
for j:=0; j<4; j++ { setbit(m,za) }

zb := (by & 0x0f)
for j:=0; j<4; j++ { setbit(m,zb) }


return
}

