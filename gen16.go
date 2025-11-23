/*----------------------------------------------------------------------

Gerador de imagem (16 cores x 4bits)

Layout de saida:

Offset	size	Pos	Função
0	1	0x0	Valor inicial da Palette (pal[0])
1	4080	0x1	Matriz de controle (16 x 255, iniciando de 1 à 256)
4081	16	480x2	Matriz de apoio ( 4 x 4, palette (2,4,8,16))
4097	2	496x2	Tipo da mascara utilizada (xor, off, shr, srr, none)
4099	2	498x2	Mascara utilizada
4101	18	500x2	Tamanha da imagem (size, format decimal 9 digitos)
4119	2481	520x2	Dados da imagem
1439100 900	900x799 Tail information


Exemplo:


(0,0) Header.1 Palette
(1,0) Header.1 Palette Sync
(961,4) Header.1 Palette (4x4)
(993,4) Header.2 Encode/Mask
(997,4) Header.2 image size: 280 bytes
{1017,4) Image data

IMAGE DATA

(421,799) Tail.1 (tail end)
(900,799) Tail.2 information size: 50
Preparer tail:  MZ:x.sh,zcode_0.bmp,Sem descricao.,LAST,0,MZEOF,,,



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
    color.RGBA{0x00, 0x00, 0x00, 0xff},
    color.RGBA{0xff, 0x00, 0x00, 0xff},
    color.RGBA{0x00, 0xff, 0x00, 0xff},
    color.RGBA{0x00, 0x00, 0xff, 0xff},
    color.RGBA{0xff, 0xff, 0xff, 0xff},
    color.RGBA{0xff, 0xff, 0x00, 0xff},
    color.RGBA{0xff, 0x00, 0xff, 0xff},
    color.RGBA{0x00, 0xff, 0xff, 0xff},
    color.RGBA{0xff, 0x00, 0x80, 0xff},
    color.RGBA{0xff, 0x80, 0x40, 0xff},
    color.RGBA{0x80, 0x40, 0x00, 0xff},
    color.RGBA{0x00, 0x80, 0x80, 0xff},
    color.RGBA{0x80, 0x00, 0x00, 0xff},
    color.RGBA{0x80, 0x00, 0x80, 0xff},
    color.RGBA{0x80, 0x80, 0xff, 0xff},
    color.RGBA{0x80, 0xff, 0x80, 0xff} }



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


var bloco = 0
var isdata = true
var ismoref= false


var deb=false	

var fmask string = ""
var fkeyv string = ""
var nfile string = ""
var descf string = ""




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
img = image.NewRGBA(image.Rect(0,0,1800,801))


fmt.Printf("(%d,%d) Header.1 Palette\n",px,py)
img.Set(px, py, pal[0])
px++

fmt.Printf("(%d,%d) Header.1 Palette Sync\n",px,py)
for i:=1; i< 256 ; i++ {
	for j:=0; j<16 ; j++ {
		//setpixel(img,uint8(i & 0x0f))
                img.Set(px, py, pal[i & 0x0f])
                addpixel()

	}
}

fmt.Printf("(%d,%d) Header.1 Palette (4x4)\n",px,py)
for i:=0; i< 4 ; i++ {
	setpixel(img, 2)
	setpixel(img, 4)
	setpixel(img, 8)
	setpixel(img, 15)
}


fmt.Printf("(%d,%d) Header.2 Encode/Mask\n",px,py)
setpixel(img, 0)
setpixel(img, 0)


// calcular image size do bloco
s := make([]byte, 10)
dsize := (fsize - offset)

if dsize > (2481 * 256) {
	dsize  = (2481 * 256)
	offset = offset + (2481 * 256)
} 
	
s = []byte(fmt.Sprintf("%-9d ", dsize))

fmt.Printf("(%d,%d) Header.2 image size: %d bytes\n",px,py,dsize)
for i:=0; i<10; i++ {
	setpixel(img, s[i])

}

fmt.Printf("{%d,%d) Image data\n",px,py)



//fmt.Println("Current position: ", px, py)


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


	// verificar bloco limite 2484
	if bloco > 2480 {
		fmt.Println("\nBloco limite atingido! (size:", size, ", bloco:", bloco, "). Necessario criar novo arquivo!")
		ismoref = true
		break
	}
}
fmt.Printf("\nTotal de crc: %d\n", pcrc)



// finalizar imagem....
fmt.Printf("(%d,%d) Tail.0 (image end)",px,py)
isdata = false
for y:=py; y < (800 - 1); y++ {
        for x:=px; x< 1800; x++ {
		setpixel(img, 0)
		crc(0, img)
	}
}

fmt.Println("\n")
fmt.Printf("(%d,%d) Tail.1 (tail end)\n",px,py)

// definir nome do arquivo corrente...
filename := fmt.Sprintf("zcode_%d.bmp", fimg)


//Salvar o nome do arquivo original...
seqf   := fmt.Sprintf(",%d", fimg)
iscont := "LAST"
if ismoref {
	iscont = "NEXT"
}
bt := "MZ:" + nfile + "," + filename + "," + descf + "," + iscont + seqf + ",MZEOF,,,"
bc := len(bt)

isdata = false
px=900
py=799
xbc := 0
fmt.Printf("(%d,%d) Tail.2 information size: %d\n",px,py,bc)
fmt.Println("Preparer tail: ", bt)
for x:=px; x < 1800 ; x++ {
	if xbc < bc {
		setpixel(img, bt[xbc])
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






//-----------------------------------------------------
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

//-----------------------------------------------------
//
func addpixel() {
px++

if px >= 1800 && py >= 799 {
	if isdata { 
		fmt.Println("Estouro de limite da imagem (1800 x 800).", size, ", blocos: ", bloco)
	}
}

if px >= 1800 { 
	px = 0
	py++
}

if py >= 800 {  py=799 }

return
}



//-----------------------------------------------------
//
func setpixel(m image.Image, by uint8)  {

hb := (by & 0xf0) >> 4
lb := by & 0x0f

m.(*image.RGBA).Set(px,py, pal[hb])
addpixel()

m.(*image.RGBA).Set(px,py, pal[lb])
addpixel()

return
}

