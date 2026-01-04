package main

import (
	"image"
	"os"
	"math"
//	"math/rand"
	"time"
	"fmt"
	"io/ioutil"
	"encoding/xml"

	_ "image/png"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

/*******
 * ImageMap Definitions
 ********/
type SpriteSheet_SubTexture struct {
	XMLName xml.Name 	`xml:"SubTexture"`
	Name	string		`xml:"name,attr"`
	X		int			`xml:"x,attr"`
	Y		int			`xml:"y,attr"`
	Width	int			`xml:"width,attr"`
	Height	int			`xml:"height,attr"`
} 

type SpriteSheet_TileGroup struct {
	XMLName		xml.Name					`xml:"TileGroup"`
	Name		string						`xml:"name,attr"`
	SubTextures	[]SpriteSheet_SubTexture	`xml:"SubTexture"`
}

type SpriteSheet_TextureAtlas struct {
	XMLName		xml.Name					`xml:"TextureAtlas"`
	TileGroups	[]SpriteSheet_TileGroup		`xml:"TileGroup"`
}

func loadTextureAtlas(path string) (SpriteSheet_TextureAtlas, error) {

	var atlas SpriteSheet_TextureAtlas

	file, err := os.Open(path)
	if err != nil {
		return atlas, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return atlas, err
	}

	err = xml.Unmarshal(bytes, &atlas)
	if err != nil {
		return atlas, err
	}

	return atlas, nil
}


func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func run() {
	// All game code here!
	cfg := opengl.WindowConfig{
		Title: "City Game!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync: true,
	}

	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	win.SetSmooth(true)

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(pixel.V(100, 400), basicAtlas)

	basicTxt.Color = colornames.White
	fmt.Fprintln(basicTxt, "Hello, city!")

	spritesheet, err := loadPicture("resources/images/cityTiles_sheet.png")
	if err != nil {
		panic(err)
	}
	atlas, err := loadTextureAtlas("resources/images/cityTiles_sheet.xml")
	if err != nil {
		panic(err)
	}

	/*
	var cityTiles []pixel.Rect
	for _, texture := range atlas.SubTextures {
		cityTiles = append(cityTiles, pixel.R(
			spritesheet.Bounds().Min.X + float64(texture.X), 
			spritesheet.Bounds().Max.Y - float64(texture.Y) - float64(texture.Height), 
			spritesheet.Bounds().Min.X + float64(texture.X + texture.Width), 
			spritesheet.Bounds().Max.Y - float64(texture.Y)))
	}
			*/

	// Create a batch for better drawing performance
	batch := pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)

	var (
		currentTileGroup	= 0
		currentTile 		= 0

		camPos			= pixel.ZV
		camSpeed		= 500.0
		camZoom			= 1.0
		camZoomSpeed	= 1.2
//		tiles			[]*pixel.Sprite
//		matrices		[]pixel.Matrix
		frames			= 0
		second			= time.Tick(time.Second)
	)
/*
	tileWidth := 132.0
	tileHeight := 66.0
	for x := 0; x < 10; x++ {
		for y := 10; y > 0; y-- {
			xpos := (float64(x) * tileWidth / 2.0) +
				(float64(y) * tileWidth / 2.0)
			ypos := (-1.0 * float64(x) * tileHeight / 2.0) +
				(float64(y) * tileHeight / 2.0)
			
			whichTile := cityTiles[rand.Intn(len(cityTiles))]
			tile := pixel.NewSprite(spritesheet, whichTile)
			tiles = append(tiles, tile)
			matrices = append(matrices, pixel.IM.Moved(pixel.V(xpos, ypos + whichTile.Bounds().Max.Y - whichTile.Bounds().Min.Y)))
		}
	}
*/
	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()


		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.JustPressed(pixel.MouseButtonLeft) {
			currentTileGroup++
			currentTile = 0
			if currentTileGroup >= len(atlas.TileGroups) {
				currentTileGroup = 0
			}
		}
		if win.JustPressed(pixel.MouseButtonRight) {
			currentTile++
			if currentTile >= len(atlas.TileGroups[currentTileGroup].SubTextures) {
				currentTile = 0
			}
		}

/*		if win.JustPressed(pixel.MouseButtonLeft) {
			tile := pixel.NewSprite(spritesheet, cityTiles[rand.Intn(len(cityTiles))])
			tiles = append(tiles, tile)
			mouse := cam.Unproject(win.MousePosition())
			matrices = append(matrices, pixel.IM.Moved(mouse))
		}
*/
		if win.Pressed(pixel.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixel.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixel.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		if win.Pressed(pixel.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		if win.Pressed(pixel.KeyEscape) {
			return;
		}

		// Get the current tile bounds and create a sprite
		texture := atlas.TileGroups[currentTileGroup].SubTextures[currentTile]
		tileBounds := pixel.R(
			spritesheet.Bounds().Min.X + float64(texture.X), 
			spritesheet.Bounds().Max.Y - float64(texture.Y) - float64(texture.Height), 
			spritesheet.Bounds().Min.X + float64(texture.X + texture.Width), 
			spritesheet.Bounds().Max.Y - float64(texture.Y))
		tile := pixel.NewSprite(spritesheet, tileBounds)
		
		win.Clear(colornames.Grey)
		batch.Clear()

		mouse := cam.Unproject(win.MousePosition())
		// for i, tile := range tiles {
			tile.Draw(batch, pixel.IM.Moved(mouse
				))
		//}
		batch.Draw(win)

		frames++
		select {
		case <-second:
			basicTxt.Clear()
			fmt.Fprintf(basicTxt, "%d fps", frames)
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
		basicTxt.Draw(win, pixel.IM)

		win.Update()
	}
}

func main() {
	opengl.Run(run)
}