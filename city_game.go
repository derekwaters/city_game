package main

import (
	"math"
	"time"
	"fmt"
	"log/slog"

	_ "image/png"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"golang.org/x/image/colornames"
)



func run() {
	// Initiate the window
	cfg := opengl.WindowConfig{
		Title: "City Game!",
		Bounds: pixel.R(0, 0, 1280, 768),
		VSync: true,
	}

	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	win.SetSmooth(true)


	// Setup the game data
	gameData, err := InitGameElements()
	if err != nil {
		panic(err)
	}




	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()





		gameData.debugMode = false
		if win.JustPressed(pixel.KeyD) {
			gameData.debugMode = true
		}

		cam := pixel.IM.Scaled(gameData.camPos, gameData.camZoom).Moved(win.Bounds().Center().Sub(gameData.camPos))
		win.SetMatrix(cam)

		if gameData.debugMode {
			slog.Info("Camera: ", "camPos", gameData.camPos, "camZoom", gameData.camZoom)
		}

		if win.JustPressed(pixel.MouseButtonRight) {
			gameData.cycleNextTile()
		}

		if win.Pressed(pixel.KeyLeft) {
			gameData.scrollLeft(dt)
		}
		if win.Pressed(pixel.KeyRight) {
			gameData.scrollRight(dt)
		}
		if win.Pressed(pixel.KeyUp) {
			gameData.scrollUp(dt)
		}
		if win.Pressed(pixel.KeyDown) {
			gameData.scrollDown(dt)
		}
		gameData.zoom(win.MouseScroll().Y)

		if win.Pressed(pixel.KeyEscape) {
			return;
		}

		tile := gameData.getCurrentTile()

		if gameData.debugMode {
			slog.Info("Current Tile: ", "currentTileGroup", gameData.currentTileGroup, 
				"currentTile", gameData.currentTile, "tileBounds", tile.sprite.Frame()) 
		}

		win.Clear(colornames.Grey)
		gameData.imdraw.Draw(win)
		gameData.tileBatch.Clear()

		// Need to snap mouse to the "nearest" tile pos (note we subtract TILE_HEIGHT because we're
		// adjusting the yPos of the drawn tiles by their height later).
		mouseOrig := cam.Unproject(win.MousePosition())
		logicalX := int(math.Round((mouseOrig.X / TILE_WIDTH) - ((mouseOrig.Y - TILE_HEIGHT) / TILE_HEIGHT)))
		logicalY := int(math.Round((mouseOrig.X / TILE_WIDTH) + ((mouseOrig.Y - TILE_HEIGHT) / TILE_HEIGHT)))
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

		if gameData.debugMode {
			slog.Info("Mouse Orig Info: ", "win.MousePosition", win.MousePosition(), "mouse", mouseOrig)
			slog.Info("Logical position: ", "X", logicalX, "Y", logicalY)
		}

		mousePos := gameData.mapTilePos(logicalX, logicalY)
		mousePos.Y += 
			((tile.sprite.Frame().Max.Y - tile.sprite.Frame().Min.Y) / 2.0) + 8.0
			// A little extra to make the "moving" tile float a bit.

		if gameData.debugMode {
			slog.Info("Adjusted Tile Pos: ", "mouse", mousePos)
		}

		if win.JustPressed(pixel.MouseButtonLeft) {
			gameData.checkAddCurrentTile(logicalX, logicalY, tile)
			if gameData.debugMode {
				slog.Info("Possibly add new tile: ", "win.MousePosition", win.MousePosition(), "logicalX", logicalX, "logicalY", logicalY)
			}
		}
		
		for x := 0; x < BOARD_SIZE; x++ {
			for y := BOARD_SIZE - 1; y >= 0; y-- {

				boardTile := gameData.boardTiles[x][y]

				if boardTile.sprite != nil {
					mappedPos := gameData.mapTilePos(x, y)
					mappedPos.Y += 
						(boardTile.sprite.Frame().Max.Y -
						boardTile.sprite.Frame().Min.Y) / 2.0
			
					if gameData.debugMode {
						slog.Info("Tile: ", "X", x, "Y", y, "width", boardTile.sprite.Frame().W(), "height", boardTile.sprite.Frame().H())
						slog.Info("Tile Draw Pos: ", "X", mappedPos.X, "Y", mappedPos.Y)
					}

					// Might need to also add the height of the tile here...
					mat := pixel.IM.Moved(mappedPos)

					boardTile.sprite.Draw(gameData.tileBatch, mat)
				}

				if x == logicalX && y == logicalY {
					tile.sprite.Draw(gameData.tileBatch, pixel.IM.Moved(mousePos))
				}
			}
		}
		
		gameData.tileBatch.Draw(win)

		gameData.frames++
		select {
		case <-gameData.second:
			gameData.fpsText.Clear()
			fmt.Fprintf(gameData.fpsText, "%d fps", gameData.frames)
			// win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			gameData.frames = 0
		default:
		}
		gameData.scoreText.Clear()
		fmt.Fprintf(gameData.scoreText, "Score: %d", gameData.score)
		gameData.scoreText.Draw(win, pixel.IM)

		gameData.fpsText.Draw(win, pixel.IM)
		gameData.mousePosText.Clear()
		fmt.Fprintf(gameData.mousePosText, "Mouse: %d, %d", int(mouseOrig.X), int(mouseOrig.Y))
		gameData.mousePosText.Draw(win, pixel.IM)

		gameData.scoreText.Clear()
		fmt.Fprintf(gameData.scoreText, "Score: %d", gameData.score)
		gameData.scoreText.Draw(win, pixel.IM)

		gameData.debugText.Clear()
		fmt.Fprintf(gameData.debugText, "Group: %s, Offset: %d, TL: %s, TR: %s, BR: %s, BL: %s", 
			gameData.getCurrentTileGroupName(),
			gameData.currentTile,
			gameData.getCurrentTileJoinTL(),
			gameData.getCurrentTileJoinTR(),
			gameData.getCurrentTileJoinBR(),
			gameData.getCurrentTileJoinBL())
		gameData.debugText.Draw(win, pixel.IM)


		win.Update()
	}
}

func main() {
	opengl.Run(run)
}