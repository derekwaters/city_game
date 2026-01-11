package main

import (
	"math"
	"math/rand"
	"time"
	"fmt"
	"log/slog"

	_ "image/png"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)



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
	mouseTxt := text.New(pixel.V(100, 450), basicAtlas)

	basicTxt.Color = colornames.White
	fmt.Fprintln(basicTxt, "Hello, city!")
	mouseTxt.Color = colornames.White

	spritesheet, err := loadPicture("resources/images/cityTiles_sheet.png")
	if err != nil {
		panic(err)
	}
	atlas, err := loadTextureAtlas("resources/images/cityTiles_sheet.xml")
	if err != nil {
		panic(err)
	}

	// Create a batch for better drawing performance
	batch := pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)

	var (
		currentTileGroup	= rand.Intn(len(atlas.TileGroups))
		currentTile 		= 0

		camPos			= pixel.ZV
		camSpeed		= 500.0
		camZoom			= 1.0
		camZoomSpeed	= 1.2
//		tiles			[]*pixel.Sprite
//		matrices		[]pixel.Matrix
		frames			= 0
		second			= time.Tick(time.Second)
		debugMode		= false
	)

	const BOARD_SIZE = 10
	var boardTiles [BOARD_SIZE][BOARD_SIZE]*pixel.Sprite
	var tileWidth = 132.0
	var tileHeight = 66.0

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		debugMode = false
		if win.JustPressed(pixel.KeyD) {
			debugMode = true
		}

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if debugMode {
			slog.Info("Camera: ", "camPos", camPos, "camZoom", camZoom)
		}

		if win.JustPressed(pixel.MouseButtonRight) {
			currentTile++
			if currentTile >= len(atlas.TileGroups[currentTileGroup].SubTextures) {
				currentTile = 0
			}
		}

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
		
		if debugMode {
			slog.Info("Current Tile: ", "currentTileGroup", currentTileGroup, 
				"currentTile", currentTile, "tileBounds", tileBounds) 
		}

		win.Clear(colornames.Grey)
		batch.Clear()

		mouseOrig := cam.Unproject(win.MousePosition())

		if debugMode {
			slog.Info("Mouse Orig Info: ", "win.MousePosition", win.MousePosition(), "mouse", mouseOrig)
		}

		// Need to snap mouse to the "nearest" tile pos (note we subtract tileHeight because we're
		// adjusting the yPos of the drawn tiles by their height later).
		logicalX := int(math.Round((mouseOrig.X / tileWidth) - ((mouseOrig.Y - tileHeight) / tileHeight)))
		logicalY := int(math.Round((mouseOrig.X / tileWidth) + ((mouseOrig.Y - tileHeight) / tileHeight)))
		if logicalX < 0 {
			logicalX = 0
		}
		if logicalX >= BOARD_SIZE {
			logicalX = BOARD_SIZE - 1
		}
		if logicalY < 0 {
			logicalY = 0
		}
		if logicalY >= BOARD_SIZE {
			logicalY = BOARD_SIZE - 1
		}

		if debugMode {
			slog.Info("Logical position: ", "X", logicalX, "Y", logicalY)
		}

		xpos := (float64(logicalX) * tileWidth / 2.0) +
			(float64(logicalY) * tileWidth / 2.0)
		ypos := (-1.0 * float64(logicalX) * tileHeight / 2.0) +
			(float64(logicalY) * tileHeight / 2.0) +
			tile.Frame().Max.Y - tile.Frame().Min.Y
		mouse := pixel.V(xpos, ypos)


		if debugMode {
			slog.Info("Adjusted Tile Pos: ", "mouse", mouse)
		}

		if win.JustPressed(pixel.MouseButtonLeft) && 
			boardTiles[logicalX][logicalY] == nil {

			if debugMode {
				slog.Info("Adding new tile: ", "win.MousePosition", win.MousePosition(), "logicalX", logicalX, "logicalY", logicalY)
			}

			boardTiles[logicalX][logicalY] = tile
			
			currentTileGroup = rand.Intn(len(atlas.TileGroups))
			currentTile = 0
		}
		
		for x := 0; x < BOARD_SIZE; x++ {
			for y := BOARD_SIZE - 1; y >= 0; y-- {
				if boardTiles[x][y] != nil {
					xpos := (float64(x) * tileWidth / 2.0) +
						(float64(y) * tileWidth / 2.0)
					ypos := (-1.0 * float64(x) * tileHeight / 2.0) +
						(float64(y) * tileHeight / 2.0) + 
						boardTiles[x][y].Frame().Max.Y -
						boardTiles[x][y].Frame().Min.Y
			
					if debugMode {
						slog.Info("Tile: ", "X", x, "Y", y, "width", boardTiles[x][y].Frame().W(), "height", boardTiles[x][y].Frame().H())
						slog.Info("Tile Draw Pos: ", "X", xpos, "Y", ypos)
					}

					// Might need to also add the height of the tile here...
					mat := pixel.IM.Moved(pixel.V(xpos, ypos))

					boardTiles[x][y].Draw(batch, mat)
				}

				if x == logicalX && y == logicalY {
					tile.Draw(batch, pixel.IM.Moved(mouse))
				}
			}
		}
		
		batch.Draw(win)

		frames++
		select {
		case <-second:
			basicTxt.Clear()
			fmt.Fprintf(basicTxt, "%d fps", frames)
			// win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
		basicTxt.Draw(win, pixel.IM)
		mouseTxt.Clear()
		fmt.Fprintf(mouseTxt, "Mouse: %d, %d", int(mouseOrig.X), int(mouseOrig.Y))
		mouseTxt.Draw(win, pixel.IM)

		win.Update()
	}
}

func main() {
	opengl.Run(run)
}