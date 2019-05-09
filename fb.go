package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/veandco/go-sdl2/img"

	"github.com/veandco/go-sdl2/ttf"

	"github.com/veandco/go-sdl2/sdl"
)

const WINHEIGHT = 600
const WINWIDTH = 800
const FOLDERHEIGHT = 100
const FOLDERWIDTH = 100
const TEXTHEIGHT = WINHEIGHT / 24
const TEXTPADDING = WINWIDTH / 64

//File contains the info and the current position and size of the folder
type File struct {
	info os.FileInfo
	rect sdl.Rect
}

//Pane is the currently viewed directory
type Pane struct {
	rect       sdl.Rect
	files      []File
	txt        []sdl.Rect
	renderFont bool
}

var cwd string
var window *sdl.Window
var rend *sdl.Renderer
var font *ttf.Font
var smallFont *ttf.Font
var folder *sdl.Texture
var txt *sdl.Texture
var blank *sdl.Texture
var sur *sdl.Surface

func init() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	err = ttf.Init()
	if err != nil {
		panic(err)
	}
	window, err = sdl.CreateWindow("fb", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WINWIDTH, WINHEIGHT, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	rend, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		panic(err)
	}
	font, err = ttf.OpenFont("./assets/Lato.ttf", 20)
	if err != nil {
		panic(err)
	}
	smallFont, err = ttf.OpenFont("./assets/Lato-light.ttf", 16)
	if err != nil {
		panic(err)
	}
	sur, err = img.Load("./assets/folder.png")
	if err != nil {
		panic(err)
	}
	folder, err = rend.CreateTextureFromSurface(sur)
	if err != nil {
		panic(err)
	}
	sur, err = img.Load("./assets/txt.png")
	if err != nil {
		panic(err)
	}
	txt, err = rend.CreateTextureFromSurface(sur)
	if err != nil {
		panic(err)
	}
	sur, err = img.Load("./assets/blank.png")
	if err != nil {
		panic(err)
	}
	blank, err = rend.CreateTextureFromSurface(sur)
	if err != nil {
		panic(err)
	}
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	sur, err = window.GetSurface()
	if err != nil {
		panic(err)
	}
}

func launchError(errString string) {
	buttons := []sdl.MessageBoxButtonData{
		{sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, 1, "Okay"},
	}

	messageboxdata := sdl.MessageBoxData{
		sdl.MESSAGEBOX_INFORMATION,
		nil,
		"Error",
		errString,
		buttons,
		nil,
	}

	sdl.ShowMessageBox(&messageboxdata)
}

func (pane *Pane) getFiles() {
	cwd, err := os.Getwd()
	if err != nil {
		launchError(err.Error())
		return
	}
	files, err := ioutil.ReadDir(cwd)
	pane.files = nil
	pane.txt = nil
	for i := 0; i < len(files); i++ {
		pane.files = append(pane.files, File{info: files[i]})
		pane.txt = append(pane.txt, sdl.Rect{})
	}
}

func (pane *Pane) setPosition() {
	nFit := int((pane.rect.W / FOLDERWIDTH))
	if len(pane.files)/nFit == 0 && len(pane.files) != 0 {
		for j := 0; j < len(pane.files); j++ {
			pane.files[j].rect = sdl.Rect{X: (int32(j) * FOLDERWIDTH) + pane.rect.X, Y: pane.rect.Y, W: FOLDERWIDTH, H: FOLDERHEIGHT}
			pane.txt[j] = sdl.Rect{X: ((int32(j) * FOLDERWIDTH) + pane.rect.X) + TEXTPADDING, Y: (pane.rect.Y + FOLDERHEIGHT), W: FOLDERWIDTH - TEXTPADDING, H: FOLDERHEIGHT}
		}
	} else if len(pane.files) != 0 {
		for i := 0; i <= len(pane.files)/nFit; i++ {
			for j := 0; j < nFit && (i*nFit)+j < len(pane.files); j++ {
				pane.files[(i*nFit)+j].rect = sdl.Rect{X: (int32(j) * FOLDERWIDTH) + pane.rect.X, Y: (int32(i) * (FOLDERHEIGHT + TEXTHEIGHT)) + pane.rect.Y, W: FOLDERWIDTH, H: FOLDERHEIGHT}
				pane.txt[(i*nFit)+j] = sdl.Rect{X: ((int32(j) * FOLDERWIDTH) + pane.rect.X) + TEXTPADDING, Y: (int32(i) * (FOLDERHEIGHT + TEXTHEIGHT)) + (pane.rect.Y + FOLDERHEIGHT), W: FOLDERWIDTH - TEXTPADDING, H: FOLDERHEIGHT}
			}
		}
	}
}

func (pane *Pane) drawFolders() {
	for i := 0; i < len(pane.files); i++ {
		if pane.files[i].info.IsDir() {
			rend.Copy(folder, nil, &pane.files[i].rect)
		} else {
			if len(pane.files[i].info.Name()) > 4 {
				switch pane.files[i].info.Name()[len(pane.files[i].info.Name())-4:] {
				case ".txt":
					rend.Copy(txt, nil, &pane.files[i].rect)
				default:
					rend.Copy(blank, nil, &pane.files[i].rect)
				}
			} else {
				rend.Copy(blank, nil, &pane.files[i].rect)
			}
		}
		if pane.renderFont {
			solid, err := smallFont.RenderUTF8Blended(pane.files[i].info.Name(), sdl.Color{R: 0xF5, G: 0xF5, B: 0xF5, A: 0})
			if err != nil {
				panic(err)
			}
			err = solid.Blit(nil, sur, &pane.txt[i])
			if err != nil {
				panic(err)
			}
		}
	}
	pane.renderFont = false
}

func (pane *Pane) cd(path string) {
	err := os.Chdir(path)
	if err != nil {
		launchError(err.Error())
		return
	}
	cwd, err = os.Getwd()
	if err != nil {
		launchError(err.Error())
		return
	}
	pane.getFiles()
	pane.setPosition()
	pane.refresh()
}

func (pane *Pane) refresh() {
	rend.Clear()
	addrBar := sdl.Rect{X: (WINWIDTH / 16), Y: 0, W: WINWIDTH, H: (WINHEIGHT / 24)}
	sur.FillRect(&addrBar, 0xffEEEEEE)

	solid, err := font.RenderUTF8Blended(cwd, sdl.Color{R: 21, G: 21, B: 21, A: 0})
	if err != nil {
		panic(err)
	}
	err = solid.Blit(nil, sur, &addrBar)
	if err != nil {
		panic(err)
	}

	pane.rect = sdl.Rect{X: (WINWIDTH / 16), Y: (WINHEIGHT / 24), W: WINWIDTH - (WINWIDTH / 16), H: WINHEIGHT}
	sur.FillRect(&pane.rect, 0xff424242)
}

func (pane *Pane) getFile(x, y int32) (File, error) {
	for i := 0; i < len(pane.files); i++ {
		if pane.files[i].rect.X < x && pane.files[i].rect.X+pane.files[i].rect.W > x && pane.files[i].rect.Y < y && pane.files[i].rect.Y+pane.files[i].rect.H > y {
			return pane.files[i], nil
		}
	}
	return File{}, errors.New("No file found")
}

func main() {
	var pane Pane
	pane.renderFont = true
	sur.FillRect(nil, 0)
	pane.refresh()
	pane.getFiles()
	pane.setPosition()

	running := true
	for running {
		pane.drawFolders()
		window.UpdateSurface()
		rend.Present()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			case *sdl.MouseButtonEvent:
				if t.State == 1 {
					file, err := pane.getFile(t.X, t.Y)
					if err != nil {
						break
					}
					if file.info.IsDir() {
						pane.renderFont = true
						pane.cd(file.info.Name())
					} else {
						cmd := exec.Command(cwd + "/" + file.info.Name())
						err := cmd.Start()
						if err != nil {
							launchError(err.Error())
						}
					}
				}
				break
			case *sdl.KeyboardEvent:
				switch t.Keysym.Sym {
				case sdl.K_BACKSPACE:
					if t.State == 1 {
						pane.renderFont = true
						pane.cd("..")
					}
				}
			}
		}
		sdl.Delay(100)
	}
	sdl.Quit()
}
